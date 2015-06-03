package main

// type SSHClient struct {
// 	Agent  agent.Agent
// 	Config *ssh.ClientConfig
// }
//
// func (sc *SSHClient) DialAgent() error {
// 	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
// 	if err != nil {
// 		return err
// 	}
//
// 	sc.Agent = agent.NewClient(sock)
// 	// signers, err := agent.Signers()
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// func NewSSHClientConfig() (*ssh.ClientConfig, error) {
// 	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	agent := agent.NewClient(sock)
// 	signers, err := agent.Signers()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	cfg := &ssh.ClientConfig{
// 		User: usr,
// 		Auth: []ssh.AuthMethod{ssh.PublicKeys(signers...)},
// 	}
//
// 	return cfg, nil
// }
