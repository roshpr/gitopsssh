package git

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type Cloner struct {
	RepoPath string
	Remote   string
	Branch   string
}

func NewCloner(repoPath, remote, branch string) *Cloner {
	return &Cloner{
		RepoPath: repoPath,
		Remote:   remote,
		Branch:   branch,
	}
}

func (c *Cloner) EnsureCloned() error {
	// Check if repo exists
	if _, err := os.Stat(c.RepoPath); os.IsNotExist(err) {
		log.Printf("Cloning %s into %s...", c.Remote, c.RepoPath)
		cmd := exec.Command("git", "clone", "--branch", c.Branch, c.Remote, c.RepoPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repo: %w", err)
		}
		return nil
	}

	// TODO: Check if remote matches
	log.Printf("Repo already exists at %s. Skipping clone.", c.RepoPath)
	return nil
}

func (c *Cloner) Pull() error {
	log.Printf("Pulling latest changes for %s...", c.RepoPath)
	cmd := exec.Command("git", "-C", c.RepoPath, "pull", "origin", c.Branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull changes: %w", err)
	}
	return nil
}
