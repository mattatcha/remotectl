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

	ext "github.com/MattAitchison/remotectl/providers"
	sshutil "github.com/MattAitchison/remotectl/ssh"

	// Enabled Providers
	_ "github.com/MattAitchison/remotectl/aws"
	_ "github.com/MattAitchison/remotectl/digitalocean"
	_ "github.com/MattAitchison/remotectl/stdin"

	"github.com/MattAitchison/envconfig"
)

var (
	// Version number that gets set at compile time.
	Version string

	curUser, _ = user.Current()

	sshPort = envconfig.Int("remotectl_port", 22, "ssh port to use. Note: all ports must be the same on hosts for a run.")
	// Show default ident.
	ident         = envconfig.String("remotectl_identity", "", "private key file")
	usr           = envconfig.String("remotectl_user", curUser.Username, "user to connect as")
	provider      = envconfig.String("remotectl_provider", "do", "name or comma-sep list of provider modules to use for selecting hosts")
	namespace     = envconfig.String("remotectl_namespace", "", "")
	prefixTmplStr = envconfig.String("remotectl_prefix", "{{.Name}}: ", "prefix template for host log output")
	prefixTmpl    = template.Must(template.New("prefix").Parse(prefixTmplStr))

	showVersion = flag.Bool("version", false, "show version")
	showHelp    = flag.Bool("help", false, "show this help message")
	showList    = flag.Bool("list", false, "lists selected ips and names. /etc/hosts friendly output")

	// Use current working path.
	profile = flag.String("profile", "", "sources a bash profile to load a config") // Maybe a name will default to a file in ~/.remotectl
)

func fatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func parseArgs(args []string) (q, cmd string) {
	i := flag.NFlag() + 1

	cmd = strings.Join(flag.Args(), " ")
	if len(args) <= i {
		return "", cmd
	}

	if args[i] != "--" {
		return flag.Args()[0], strings.Join(flag.Args()[1:], " ")
	}

	return "", cmd
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

	query, cmd := parseArgs(os.Args)

	var hosts []ext.Host
	for _, p := range ext.Providers.Select(strings.Fields(provider)) {
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
		printHosts(os.Stderr, hosts)
		return
	}

	cfg, err := sshutil.NewClientConfig(usr)
	fatalErr(err)

	var wg sync.WaitGroup
	wg.Add(len(hosts))
	for _, host := range hosts {
		go func(h ext.Host) {
			s := setupSession(cfg, h)

			defer func() {
				s.Close()
				wg.Done()
			}()

			// Should probably use something else other than run
			s.Run(cmd)
		}(host)
	}
	wg.Wait()
}

func setupSession(cfg *sshutil.ClientConfig, host ext.Host) *sshutil.Session {
	addr := fmt.Sprint(host.Addr, ":", sshPort)
	s, err := cfg.NewSession(addr)
	if err != nil {
		log.Fatalln(host.Name+":", addr, err)
	}

	outPipe, err := s.StdoutPipe()
	fatalErr(err)

	errPipe, err := s.StderrPipe()
	fatalErr(err)

	// Don't like this
	go hostStreamer(host, outPipe, os.Stdout)
	go hostStreamer(host, errPipe, os.Stderr)

	return s
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

func printHosts(w io.Writer, hosts []ext.Host) {
	for _, h := range hosts {
		fmt.Fprintf(w, "%-20s %s.%s\n", h.Addr, h.Name, h.Provider)
	}
}

func helpCmd() {
	usage := `Usage: remotectl <flags> <query> [--] <cmd>

Providers:
%s

Environment Vars:
`
	fmt.Printf(usage, ext.Providers.Names())
	envconfig.PrintDefaults()
	fmt.Println("\nFlags:")
	flag.PrintDefaults()

}
