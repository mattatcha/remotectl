package digitalocean

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	do := DOProvider{}
	do.Init()

	hosts, err := do.Get()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(hosts)
}
