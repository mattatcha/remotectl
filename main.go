package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"strconv"
	"strings"
	"sync"

	ext "github.com/MattAitchison/remotectl/providers"

	// Enabled Providers
	_ "github.com/MattAitchison/remotectl/digitalocean"

	env "github.com/MattAitchison/envconfig"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var (
	curUser, _ = user.Current()

	sshPort    = env.Int("remotectl_port", 22, "ssh port to use. Note: all ports must be the same on hosts for a run.")
	ident      = env.String("remotectl_identity", "", "private key file")
	usr        = env.String("remotectl_user", strings.ToLower(curUser.Username), "user to connect as")
	provider   = env.String("remotectl_provider", "do", "name or comma-sep list of provider modules to use for selecting hosts")
	hook       = env.String("remotectl_hook", "", "hook command to use as provider")
	prefixTmpl = env.String("remotectl_prefix", "{{cc.Name}}:", "prefix template for host log output")

	showVersion = flag.Bool("version", false, "show version")
	showHelp    = flag.Bool("help", false, "show this help message")
	verbose     = flag.Bool("verbose", false, "enable verbose status output")
	localMode   = flag.Bool("local", false, "runs a local shell with <cmd> instead of ssh")
	list        = flag.Bool("list", false, "lists selected ips and names. /etc/hosts friendly output")
	profile     = flag.String("profile", "", "sources a bash profile to load a config") // Maybe a name will default to a file in ~/.remotectl
)

// type Host struct {
// 	Name  string
// 	Addr  string
// 	Index int
// 	Group int
// }

func runSSHCmd(host ext.Host, cfg *ssh.ClientConfig, group *sync.WaitGroup) {
	defer group.Done()

	// lgr := log.New(os.Stdout, host.Name, 0)
	parts := strings.SplitN(host.Addr, ":", 1)
	if len(parts) != 2 {
		parts = append(parts, strconv.Itoa(sshPort))
	}

	hostAddr := strings.Join(parts, ":")
	client, err := ssh.Dial("tcp", hostAddr, cfg)
	if err != nil {
		log.Fatal(err)
	}

	session, err := client.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	session.Run(strings.Join(os.Args[1:], " "))

}

// func hostsFromStdin() (hosts []string, ok bool) {
// 	fi, err := os.Stdin.Stat()
// 	if err != nil {
// 		log.Fatal("error reading from stdin")
// 	}
// 	if fi.Mode()&os.ModeNamedPipe == 0 {
// 		ok = false
// 		return
// 	}
//
// 	scanner := bufio.NewScanner(os.Stdin)
// 	for scanner.Scan() {
// 		re := regexp.MustCompile("(?s)#.*")
// 		txt := re.ReplaceAllLiteralString(scanner.Text(), "")
// 		txt = strings.TrimSpace(txt)
// 		if len(txt) == 0 {
// 			continue
// 		}
// 		log.Println(txt)
// 		hosts = append(hosts, txt)
// 	}
// 	if err := scanner.Err(); err != nil {
// 		// Assuming that an err scanning stdin is fatal and we
// 		// shouldn't attempt to use a provider.
// 		log.Fatal("reading standard input:", err)
// 	}
//
// 	return hosts, true
// }

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	switch {
	case *showHelp:
		helpCmd()
		os.Exit(0)
	case *showVersion:
		fmt.Println("remotectl: v0.1.0")
		os.Exit(0)
	}

	hosts := []ext.Host{}
	for _, p := range ext.Providers.All() {
		p.Init()

		extHosts, _ := p.Get()

		hosts = append(hosts, extHosts...)
	}

	// hosts, ok := hostsFromStdin()
	if len(hosts) > 0 {
		log.Printf("size %v", len(hosts))
		log.Print(hosts)
	}

	cfg, err := newSSHClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	group := &sync.WaitGroup{}
	for _, host := range hosts {
		group.Add(1)
		go runSSHCmd(host, cfg, group)
	}

	group.Wait()

}

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

func helpCmd() {
	usage := `Usage: remotectl <flags> <query> [--] <cmd>
Environment Vars:
`

	fmt.Printf(usage)
	env.PrintDefaults()
}
