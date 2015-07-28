package providers

import (
	"fmt"
	"testing"
)

func TestQueryMatch(t *testing.T) {
	names := []string{
		"web.dev",
		"web",
		"web.prod",
		"web.prod.1",
		"db.prod.1",
		"db.dev.1",
		"db.dev.2",
	}
	fmt.Println(names)
	for _, q := range []string{"", "web", "dev", "prod", "web.prod", "web.prod*", "web.prod.*", "db", "db.dev", "db.dev.*", "db.dev*", "web.dev", "dev.web"} {
		fmt.Println("\nquery:", q)
		fmt.Println("matches:")
		for _, n := range names {
			if Match(q, n) {
				fmt.Println(n)
			}
		}

	}

}
