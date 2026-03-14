package git

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Manager struct {
	RepoPath string
}

func NewManager(repoPath string) *Manager {
	return &Manager{RepoPath: repoPath}
}

func (m *Manager) GetFileContent(repoRelPath string) ([]byte, error) {
	path := filepath.Join(m.RepoPath, repoRelPath)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return content, nil
}

func (m *Manager) GetFileHash(repoRelPath string) (string, error) {
	path := filepath.Join(m.RepoPath, repoRelPath)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // File doesn't exist, so no hash
		}
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
