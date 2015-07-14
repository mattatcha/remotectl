// Package ssh is based on github.com/crosbymichael/slex/blob/master/ssh.go
package ssh

import (
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// ClientConfig stores the configuration
// and the ssh agent to forward authentication requests
type ClientConfig struct {
	// agent is the connection to the ssh agent
	agent agent.Agent
	*ssh.ClientConfig
}

// Session stores the open session and connection to execute a command.
type Session struct {
	// conn is the ssh client that started the session.
	conn *ssh.Client
	*ssh.Session
}

// NewSession creates a new ssh session with the host.
// It forwards authentication to the agent when it's configured.
func (c *ClientConfig) NewSession(host string) (*Session, error) {
	conn, err := ssh.Dial("tcp", host, c.ClientConfig)
	if err != nil {
		return nil, err
	}

	if c.agent != nil {
		if err := agent.ForwardToAgent(conn, c.agent); err != nil {
			return nil, err
		}
	}

	session, err := conn.NewSession()
	if c.agent != nil {
		err = agent.RequestAgentForwarding(session)
	}

	return &Session{
		conn:    conn,
		Session: session,
	}, err
}

// RunWaitGroup will run a command and notify when done.
// TODO: Remove this. It doesn't do enough to actually be useful.
func (s *Session) RunWaitGroup(g *sync.WaitGroup, cmd string) error {
	defer g.Done()

	return s.Run(cmd)
}

// NewClientConfig returns a config using an ssh agent.
// Will also use an identity file later.
func NewClientConfig(user string) (*ClientConfig, error) {
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}

	agent := agent.NewClient(sock)
	signers, err := agent.Signers()
	if err != nil {
		return nil, err
	}

	cfg := &ClientConfig{
		agent: agent,
		ClientConfig: &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{ssh.PublicKeys(signers...)},
		},
	}

	return cfg, nil
}

//
// func WaitForSSH(addr string) error {
// 	for {
// 		log.Printf("testing TCP connection to: %s", addr)
// 		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
//
// 		if err != nil {
// 			continue
// 		}
//
// 		defer conn.Close()
// 		return nil
// 	}
// }
