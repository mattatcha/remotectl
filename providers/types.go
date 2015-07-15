//go:generate go-extpoints . Provider
package providers

type Provider interface {
	Setup() error
	Query(namespace, query string) ([]Host, error)
}

// remove extpoints
type Host struct {
	Name     string
	Addr     string
	Provider string
	Index    int
	Group    int
}
