package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/user"
	"strings"
	"sync"
	"text/template"
	"unicode"

	ext "github.com/MattAitchison/remotectl/providers"

	// Enabled Providers
	_ "github.com/MattAitchison/remotectl/digitalocean"
	_ "github.com/MattAitchison/remotectl/stdin"

	env "github.com/MattAitchison/envconfig"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var (
	// Version number that gets set at compile time.
	Version string

	curUser, _ = user.Current()

	sshPort       = env.Int("remotectl_port", 22, "ssh port to use. Note: all ports must be the same on hosts for a run.")
	ident         = env.String("remotectl_identity", "", "private key file")
	usr           = env.String("remotectl_user", strings.ToLower(curUser.Username), "user to connect as")
	provider      = env.String("remotectl_provider", "do,stdin", "name or comma-sep list of provider modules to use for selecting hosts")
	hook          = env.String("remotectl_hook", "", "hook command to use as provider")
	prefixTmplStr = env.String("remotectl_prefix", "{{.Name}}: ", "prefix template for host log output")
	prefixTmpl    = template.Must(template.New("prefix").Parse(prefixTmplStr))

	showVersion = flag.Bool("version", false, "show version")
	showHelp    = flag.Bool("help", false, "show this help message")
	verbose     = flag.Bool("verbose", false, "enable verbose status output")
	localMode   = flag.Bool("local", false, "runs a local shell with <cmd> instead of ssh")
	showList    = flag.Bool("list", false, "lists selected ips and names. /etc/hosts friendly output")
	profile     = flag.String("profile", "", "sources a bash profile to load a config") // Maybe a name will default to a file in ~/.remotectl
)

func fatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Show filename and line number in logs.
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	switch {
	case *showHelp:
		helpCmd()
		os.Exit(0)
	case *showVersion:
		fmt.Printf("remotectl: %s", Version)
		os.Exit(0)
	}

	hosts := []ext.Host{}
	// Get all of the hosts from each provider.
	// This will have a way to filter soon.
	f := func(c rune) bool {
		return c == ',' || unicode.IsSpace(c)
	}
	fmt.Println(strings.FieldsFunc(provider, f))
	for _, p := range ext.Providers.Select(strings.FieldsFunc(provider, f)) {
		if p == nil {
			log.Fatal("provider undefined")
		}
		// Call init on provider.
		// This allows each provider to do any setup that is needed.
		p.Init()
		extHosts, err := p.Get()
		fatalErr(err)
		hosts = append(hosts, extHosts...)
	}

	if len(hosts) == 0 {
		log.Fatal("no hosts")
	}

	if *showList {
		log.Println(hosts)
		return
	}

	cfg, err := newSSHClientConfig()
	fatalErr(err)

	group := &sync.WaitGroup{}
	for _, host := range hosts {
		group.Add(1)
		go runSSHCmd(host, cfg, group)
	}

	// Wait until all ssh cmds are done running.
	group.Wait()
}

// newSSHClientConfig returns a config using an ssh agent.
func newSSHClientConfig() (*ssh.ClientConfig, error) {
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}

	agent := agent.NewClient(sock)
	signers, err := agent.Signers()
	if err != nil {
		return nil, err
	}

	cfg := &ssh.ClientConfig{
		User: usr,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signers...)},
	}

	return cfg, nil
}

func hostLogger(host ext.Host, reader io.Reader) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		prefixTmpl.Execute(os.Stdout, host)
		fmt.Printf("%s \n", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error while reading from Writer: %s", err)
	}

}

func runSSHCmd(host ext.Host, cfg *ssh.ClientConfig, group *sync.WaitGroup) {
	defer group.Done()

	// Append port to host.Addr
	hostAddr := fmt.Sprintf("%s:%v", host.Addr, sshPort)
	client, err := ssh.Dial("tcp", hostAddr, cfg)
	fatalErr(err)

	// Create an ssh session.
	session, err := client.NewSession()
	fatalErr(err)
	defer session.Close()

	// Setup logging
	outPipe, err := session.StdoutPipe()
	fatalErr(err)

	errPipe, err := session.StderrPipe()
	fatalErr(err)

	go hostLogger(host, outPipe)
	go hostLogger(host, errPipe)

	// Run command that was passed into remotectl
	session.Run(strings.Join(os.Args[1:], " "))
}

func helpCmd() {
	usage := `Usage: remotectl <flags> <query> [--] <cmd>

Providers:
%s

Environment Vars:
`
	fmt.Printf(usage, ext.Providers.Names())
	env.PrintDefaults()
	fmt.Println("\nFlags:")
	flag.PrintDefaults()

}
