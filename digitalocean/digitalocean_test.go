package digitalocean

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	do := DOProvider{}
	do.Setup()

	hosts, err := do.Query()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(hosts)
}
