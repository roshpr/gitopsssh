package ssh

import (
	"fmt"
	"io/ioutil"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewClient(host string, port int, user, keyPath string) (*ssh.Client, error) {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //FIXME: In a real application, you should use a proper host key callback.
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %w", err)
	}

	return client, nil
}

func GetFileHash(client *ssh.Client, path string, sudo bool) (string, error) {
	var cmd string
	if sudo {
		cmd = fmt.Sprintf("sudo sha256sum %s", path)
	} else {
		cmd = fmt.Sprintf("sha256sum %s", path)
	}

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.Output(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	// The output of sha256sum is in the format "<hash>  <filename>"
	parts := split(string(output), " ")
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "", fmt.Errorf("unexpected output from sha256sum: %s", output)
}

func split(s, sep string) []string {
	var result []string
	for _, part := range s {
		if string(part) != sep {
			result = append(result, string(part))
		}
	}
	return result
}
