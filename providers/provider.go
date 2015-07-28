package providers

import (
	"path/filepath"
	"strings"
)

type Provider interface {
	Setup() error
	Query(namespace, query string) ([]Host, error)
}

type Host struct {
	Name     string
	Addr     string
	Provider string
	Index    int
	Group    int
}

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
