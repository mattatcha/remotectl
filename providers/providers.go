package providers

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

var providers = make(map[string]HostProvider)

func Register(provider HostProvider, name string) {
	if provider == nil {
		panic("driver is nil")
	}

	if _, dup := providers[name]; dup {
		panic("register called twice for provider " + name)
	}
	providers[name] = provider
}

func Providers() []string {
	var list []string
	for name := range providers {
		list = append(list, name)
	}

	sort.Strings(list)
	return list
}

func Get(name string, setup bool) (HostProvider, error) {
	p, found := providers[name]
	if !found {
		return nil, fmt.Errorf("Provider not registered: %s", name)
	}
	if setup {
		return p, p.Setup()
	} else {
		return p, nil
	}
}

type HostProvider interface {
	Setup() error
	Query(namespace, query string) ([]Host, error)
}

// Host represents a host from a provider
type Host struct {
	Name     string
	Addr     string
	Provider string
	Index    int
	Group    int
}

// Match attempts a filepath match then a tag match
func Match(pattern, name string) bool {
	matched, _ := filepath.Match(pattern, name)
	if pattern == "" || matched {
		return true
	}

	for _, part := range strings.Split(name, ".") {
		if pattern == part {
			return true
		}
	}

	return false
}
