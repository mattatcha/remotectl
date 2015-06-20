package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strings"
	"sync"
	"text/template"
	"unicode"

	ext "github.com/MattAitchison/remotectl/providers"
	sshutil "github.com/MattAitchison/remotectl/ssh"

	// Enabled Providers
	_ "github.com/MattAitchison/remotectl/digitalocean"
	_ "github.com/MattAitchison/remotectl/stdin"

	env "github.com/MattAitchison/envconfig"
)

var (
	// Version number that gets set at compile time.
	Version string

	curUser, _ = user.Current()

	sshPort = env.Int("remotectl_port", 22, "ssh port to use. Note: all ports must be the same on hosts for a run.")
	// Show default ident.
	ident         = env.String("remotectl_identity", "", "private key file")
	usr           = env.String("remotectl_user", strings.ToLower(curUser.Username), "user to connect as")
	provider      = env.String("remotectl_provider", "do", "name or comma-sep list of provider modules to use for selecting hosts")
	namespace     = env.String("remotectl_namespace", "", "")
	prefixTmplStr = env.String("remotectl_prefix", "{{.Name}}: ", "prefix template for host log output")
	prefixTmpl    = template.Must(template.New("prefix").Parse(prefixTmplStr))

	showVersion = flag.Bool("version", false, "show version")
	showHelp    = flag.Bool("help", false, "show this help message")
	verbose     = flag.Bool("verbose", false, "enable verbose status output")
	localMode   = flag.Bool("local", false, "runs a local shell with <cmd> instead of ssh")
	showList    = flag.Bool("list", false, "lists selected ips and names. /etc/hosts friendly output")

	// Use current working path.
	profile = flag.String("profile", "", "sources a bash profile to load a config") // Maybe a name will default to a file in ~/.remotectl
)

func fatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func fields(s string) []string {
	f := func(c rune) bool {
		return c == ',' || unicode.IsSpace(c)
	}
	return strings.FieldsFunc(s, f)
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	switch {
	case *showHelp:
		helpCmd()
		os.Exit(0)
	case *showVersion:
		fmt.Printf("remotectl: %s", Version)
		os.Exit(0)
	}

	query := flag.Args()[0]
	cmd := strings.Join(flag.Args()[1:], " ")

	// Get all of the hosts from each provider.
	// This will have a way to filter soon.
	var hosts []ext.Host
	for _, p := range ext.Providers.Select(fields(provider)) {
		if p == nil {
			log.Fatal("unknown provider")
		}

		// Setup the provider
		p.Setup()

		// Query the provider for hosts
		extHosts, err := p.Query(namespace, query)
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

	cfg, err := sshutil.NewClientConfig(usr)
	fatalErr(err)

	log.Println("Connecting...")
	var sessions []*sshutil.Session
	for _, host := range hosts {

		addr := fmt.Sprint(host.Addr, ":", sshPort)
		sess, err := cfg.NewSession(addr)
		if err != nil {
			log.Fatalln(host.Name+":", addr, err)
		}

		outPipe, err := sess.StdoutPipe()
		fatalErr(err)

		errPipe, err := sess.StderrPipe()
		fatalErr(err)

		go hostStreamer(host, outPipe, os.Stdout)
		go hostStreamer(host, errPipe, os.Stderr)
		sessions = append(sessions, sess)
	}

	if len(sessions) == 0 {
		log.Fatal("no sessions")
	}

	group := &sync.WaitGroup{}
	for _, s := range sessions {
		group.Add(1)
		// TODO: Close s when done.
		go s.RunWaitGroup(group, cmd)
	}
	group.Wait()
}

func hostStreamer(host ext.Host, r io.Reader, w io.Writer) {
	// Locking? One writer?
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		prefixTmpl.Execute(w, host)
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("err streaming logs %s", err)
	}

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
