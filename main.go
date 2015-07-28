package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/MattAitchison/bashenv"
	"github.com/MattAitchison/envconfig"
	ext "github.com/MattAitchison/remotectl/providers"
	sshutil "github.com/MattAitchison/remotectl/ssh"

	// Enabled Providers
	_ "github.com/MattAitchison/remotectl/digitalocean"
)

var (
	// Version number that gets set at compile time.
	Version string

	curUser, _ = user.Current()

	sshPort       = envconfig.Int("remotectl_port", 22, "port used to connect to each host")
	_             = envconfig.String("SSH_AUTH_SOCK", "", "ssh agent socket")
	ident         = envconfig.String("remotectl_identity", "", "file from which the identity (private key) for public key authentication is read.")
	usr           = envconfig.String("remotectl_user", curUser.Username, "user to connect as")
	provider      = envconfig.String("remotectl_provider", "do", "comma-sep list of provider modules to use for selecting hosts")
	namespace     = envconfig.String("remotectl_namespace", "", "namespace is a prefix which is matched and removed from hosts")
	prefixTmplStr = envconfig.String("remotectl_prefix", "{{.Name}}: ", "prefix template for host log output")
	prefixTmpl    = template.Must(template.New("prefix").Parse(prefixTmplStr))
	profile       = flag.String("profile", "", "bash profile to source for env config") // Maybe a name will default to a file in ~/.remotectl

	showVersion = flag.Bool("version", false, "show version")
	showHelp    = flag.Bool("help", false, "show this help message")
	showList    = flag.Bool("list", false, "lists selected ips and names. /etc/hosts friendly output")
)

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// The following command `remotectl -profile test.sh -list` would result in
// len(args) <= i returning false.
func parseArgs(args []string) (query, cmd string) {
	i := flag.NFlag() + 1

	cmd = strings.Join(flag.Args(), " ")

	if len(args) > i && args[i] != "--" {
		query = flag.Args()[0]
		cmd = strings.Join(flag.Args()[1:], " ")
	}

	return query, cmd
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	if *profile != "" {
		path := filepath.Clean(*profile)
		path, err := filepath.Abs(path)
		fatal(err)

		fatal(bashenv.Source(path))
	}

	switch {
	case *showHelp:
		helpCmd()
		os.Exit(0)
	case *showVersion:
		log.Printf("remotectl: %s", Version)
		os.Exit(0)
	}

	query, cmd := parseArgs(os.Args)

	var hosts []ext.Host
	for _, p := range ext.Providers.Select(strings.Fields(provider)) {
		if p == nil {
			log.Fatal("unknown provider")
		}

		// Setup the provider
		fatal(p.Setup())

		// Query the provider for hosts
		extHosts, err := p.Query(namespace, query)
		fatal(err)

		hosts = append(hosts, extHosts...)
	}

	if len(hosts) == 0 {
		log.Fatal("no hosts selected")
	}

	if *showList {
		printHosts(os.Stdout, hosts)
		return
	}

	cfg, err := sshutil.NewClientConfig(ident, usr)
	fatal(err)

	var wg sync.WaitGroup
	wg.Add(len(hosts))
	for _, host := range hosts {
		go func(h ext.Host) {
			s := newSession(cfg, h)

			defer func() {
				s.Close()
				wg.Done()
			}()

			// Should probably use something else other than run
			// Using run and stdoutpipe/stderrpipe could result in lost output
			s.Run(cmd)
		}(host)
	}
	wg.Wait()
}

func newSession(cfg *sshutil.ClientConfig, host ext.Host) *sshutil.Session {
	addr := fmt.Sprint(host.Addr, ":", sshPort)
	s, err := cfg.NewSession(addr)
	if err != nil {
		log.Printf("error connecting to host: %s with user: %s", host.Name, cfg.User)
		log.Fatal(err)
	}

	outPipe, err := s.StdoutPipe()
	fatal(err)

	errPipe, err := s.StderrPipe()
	fatal(err)

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
		fmt.Fprintln(w, scanner.Text())
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
