package digitalocean

import (
	"log"
	"strings"

	"code.google.com/p/goauth2/oauth"
	env "github.com/MattAitchison/envconfig"
	"github.com/MattAitchison/remotectl/providers"
	"github.com/digitalocean/godo"
)

var doToken = env.String("do_access_token", "", "digitalocean PAT token")

func init() {
	providers.Providers.Register(new(DOProvider), "do")
}

// DOProvider is a provider for digitalocean
type DOProvider struct {
	client *godo.Client
}

// Setup will get the DO key and login.
func (p *DOProvider) Setup() {
	doToken = strings.TrimSpace(doToken)
	if len(doToken) == 0 {
		log.Fatal("access key required")
	}

	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: doToken},
	}

	p.client = godo.NewClient(t.Client())
}

func getPublicIP(droplet godo.Droplet) string {
	for _, net := range droplet.Networks.V4 {
		if net.Type == "public" {
			return net.IPAddress
		}
	}

	// FIXME: Shouldnt return an empty string if an IP address can't be found.
	return ""
}

// Query will retrieve all digitalocean droplets
// Rename to Query with namespace and query string as args.
func (p *DOProvider) Query(namespace, query string) ([]providers.Host, error) {
	drops, err := p.dropletList()
	if err != nil {
		return nil, err
	}

	hosts := []providers.Host{}

	for _, drop := range drops {
		if len(namespace) != 0 && !strings.HasPrefix(drop.Name, namespace) {
			continue
		}
		if len(query) != 0 && !strings.HasPrefix(drop.Name, query) {
			continue
		}

		host := providers.Host{
			Name: drop.Name,
			Addr: getPublicIP(drop),
		}

		hosts = append(hosts, host)
	}

	return hosts, nil
}

func (p *DOProvider) dropletList() ([]godo.Droplet, error) {
	// create a list to hold our droplets
	list := []godo.Droplet{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := p.client.Droplets.List(opt)
		if err != nil {
			return nil, err
		}

		// append the current page's droplets to our list
		for _, d := range droplets {
			list = append(list, d)
		}

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}
