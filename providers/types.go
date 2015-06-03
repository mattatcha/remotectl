//go:generate go-extpoints . Provider
package providers

type Provider interface {
	Init()
	Get() ([]Host, error)
}

type Host struct {
	Name  string
	Addr  string
	Index int
	Group int
}
