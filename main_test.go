package main

import (
	"flag"
	"testing"
)

func TestArgParse(t *testing.T) {
	var tests = []struct {
		args  []string
		query string
		cmd   string
	}{
		{
			args: []string{
				"--list",
				"--profile",
				"test",
				"--",
				"uname",
			},
			query: "",
			cmd:   "uname",
		},
		{
			args: []string{
				"--",
				"uname",
			},
			query: "",
			cmd:   "uname",
		},
		{
			args: []string{
				"--list",
				"--profile",
				"test",
				"web",
				"uname",
			},
			query: "web",
			cmd:   "uname",
		},
		{
			args:  []string{},
			query: "",
			cmd:   "",
		},
		{
			args: []string{
				"--list",
			},
			query: "",
			cmd:   "",
		},
		{
			args: []string{
				"--list",
				"--",
				"test",
			},
			query: "",
			cmd:   "test",
		},
		{
			args: []string{
				"--profile",
				"test.sh",
				"--list",
			},
			query: "",
			cmd:   "",
		},
	}

	for k, a := range tests {
		args := []string{"/remotectl"}
		args = append(args, a.args...)

		flag.CommandLine.Parse(a.args)
		q, c := parseArgs(args)

		if q != a.query {
			t.Errorf("test: %v query: expected: %s got: %s", k, a.query, q)
		}
		if c != a.cmd {
			t.Errorf("test: %v cmd: expected: %s got: %s", k, a.cmd, q)
		}
	}
}
