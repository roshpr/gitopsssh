package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Git      GitConfig      `yaml:"git"`
	Polling  PollingConfig  `yaml:"polling"`
	Products []ProductConfig `yaml:"products"`
}

type GitConfig struct {
	RepoPath              string `yaml:"repo_path"`
	Remote                string `yaml:"remote"`
	Branch                string `yaml:"branch"`
	RefreshIntervalSeconds int    `yaml:"refresh_interval_seconds"`
}

type PollingConfig struct {
	IntervalSeconds     int `yaml:"interval_seconds"`
	MaxConcurrency      int `yaml:"max_concurrency"`
	PerServerConcurrency int `yaml:"per_server_concurrency"`
	SSHTimeoutSeconds   int `yaml:"ssh_timeout_seconds"`
	DiffMaxBytes        int `yaml:"diff_max_bytes"`
	ContentMaxBytes     int `yaml:"content_max_bytes"`
	DiffCacheTTLSeconds int `yaml:"diff_cache_ttl_seconds"`
}

type ProductConfig struct {
	ID      string        `yaml:"id"`
	Name    string        `yaml:"name"`
	Global  GlobalConfig  `yaml:"global"`
	Servers []ServerConfig `yaml:"servers"`
}

type GlobalConfig struct {
	Files      []FileConfig `yaml:"files"`
	SSHKeyPath string       `yaml:"ssh_key_path"`
}

type ServerConfig struct {
	ID           string       `yaml:"id"`
	Name         string       `yaml:"name"`
	Host         string       `yaml:"host"`
	Port         int          `yaml:"port"`
	User         string       `yaml:"user"`
	SSHKeyPath   string       `yaml:"ssh_key_path"`
	Sudo         bool         `yaml:"sudo"`
	ExcludeFiles []string     `yaml:"exclude_files"`
	Files        []FileConfig `yaml:"files"`
}

type FileConfig struct {
	Dest        string `yaml:"dest"`
	RepoRelPath string `yaml:"repo_rel_path"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) GetEffectiveFiles(productID, serverID string) ([]FileConfig, error) {
	var product *ProductConfig
	for i := range c.Products {
		if c.Products[i].ID == productID {
			product = &c.Products[i]
			break
		}
	}
	if product == nil {
		return nil, fmt.Errorf("product not found: %s", productID)
	}

	var server *ServerConfig
	for i := range product.Servers {
		if product.Servers[i].ID == serverID {
			server = &product.Servers[i]
			break
		}
	}
	if server == nil {
		return nil, fmt.Errorf("server not found: %s for product %s", serverID, productID)
	}

	effectiveFiles := make(map[string]FileConfig)

	// Global files
	for _, f := range product.Global.Files {
		if f.RepoRelPath == "" {
			f.RepoRelPath = filepath.Join("products", product.ID, "global", "files", f.Dest)
		}
		effectiveFiles[f.Dest] = f
	}

	// Server-specific files override global ones
	for _, f := range server.Files {
		if f.RepoRelPath == "" {
			f.RepoRelPath = filepath.Join("products", product.ID, "servers", server.ID, "files", f.Dest)
		}
		effectiveFiles[f.Dest] = f
	}

	// Exclude files
	for _, excludedPath := range server.ExcludeFiles {
		delete(effectiveFiles, excludedPath)
	}

	var result []FileConfig
	for _, f := range effectiveFiles {
		result = append(result, f)
	}

	return result, nil
}
