package stdin

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/MattAitchison/remotectl/providers"
)

func init() {
	providers.Providers.Register(new(StdinProvider), "stdin")
}

// StdinProvider is a provider for digitalocean
type StdinProvider struct {
}

func (*StdinProvider) Init() {}
func (*StdinProvider) Get() ([]providers.Host, error) {
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal("error reading from stdin")
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		return nil, err
	}
	hosts := []providers.Host{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		re := regexp.MustCompile("(?s)#.*")
		txt := re.ReplaceAllLiteralString(scanner.Text(), "")
		txt = strings.TrimSpace(txt)
		if len(txt) == 0 {
			continue
		}
		host := providers.Host{
			Addr: txt,
		}
		hosts = append(hosts, host)
	}
	if err := scanner.Err(); err != nil {
		// Assuming that an err scanning stdin is fatal and we
		// shouldn't attempt to use a provider.
		log.Fatal("reading standard input:", err)
	}

	return hosts, nil
}
