// Package ssh is based on github.com/crosbymichael/slex/blob/master/ssh.go
package ssh

import (
	"io/ioutil"
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
	a agent.Agent
	*ssh.ClientConfig
	ForwardAgent bool
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

	if c.a != nil && c.ForwardAgent {
		if err := agent.ForwardToAgent(conn, c.a); err != nil {
			return nil, err
		}
	}

	session, err := conn.NewSession()
	// Move this up
	if c.a != nil && c.ForwardAgent {
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

func agentSigners() (*agent.Agent, []ssh.Signer, error) {
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, nil, err
	}

	a := agent.NewClient(sock)
	signers, err := a.Signers()
	if err != nil {
		return nil, nil, err
	}

	return &a, signers, nil
}

func pemSigner(file string) (ssh.Signer, error) {
	pemBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return ssh.ParsePrivateKey(pemBytes)
}

// NewClientConfig returns a config using an ssh agent unless ident is not empty.
func NewClientConfig(ident string, user string) (*ClientConfig, error) {
	// This should all be able to be simplified by using PublicKeysCallback
	cfg := &ClientConfig{
		ClientConfig: &ssh.ClientConfig{
			User: user,
		},
	}

	if len(ident) > 0 {
		s, err := pemSigner(ident)
		if err != nil {
			return nil, err
		}

		cfg.ClientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(s)}

		return cfg, nil
	}

	a, s, err := agentSigners()
	if err != nil {
		return nil, err
	}

	cfg.a = *a
	cfg.ClientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(s...)}

	return cfg, nil
}
