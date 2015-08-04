package providers

import (
	"path/filepath"
	"strings"
)

type Provider interface {
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
