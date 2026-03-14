package ssh

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewClient(host string, port int, user, keyPath string) (*ssh.Client, error) {
	log.Printf("Attempting to use SSH key at path: %s for user %s@%s:%d", keyPath, user, host, port)

	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key from %s: %w", keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key from %s. If the key is passphrase-protected, this is not supported: %w", keyPath, err)
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
	log.Printf("Connecting to SSH server at %s", addr)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to %s: %w", addr, err)
	}

	log.Printf("Successfully connected to SSH server at %s", addr)
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
	parts := strings.Fields(string(output))
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "", fmt.Errorf("unexpected output from sha256sum: %s", output)
}
