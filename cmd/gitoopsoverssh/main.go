package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"mymodule/internal/config"
	"mymodule/internal/git"
	"mymodule/internal/http"
	"mymodule/internal/poller"
	"mymodule/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("gitoopsOverSsh server starting...")

	cfg, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := store.NewDB("gitoops.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := store.Migrate(db, "internal/store/migrations"); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	if err := setupDatabase(db, cfg); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	cloner := git.NewCloner(cfg.Git.RepoPath, cfg.Git.Remote, cfg.Git.Branch)
	if err := cloner.EnsureCloned(); err != nil {
		log.Fatalf("Failed to clone repository: %v", err)
	}

	go func() {
		ticker := time.NewTicker(time.Duration(cfg.Git.RefreshIntervalSeconds) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := cloner.Pull(); err != nil {
				log.Printf("Error pulling repository: %v", err)
			}
		}
	}()

	ctx := context.Background()
	go poller.Poll(ctx, cfg, db)

	httpServer := http.NewServer(db)
	log.Println("Starting HTTP server on :8080")
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func setupDatabase(db *sql.DB, cfg *config.Config) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback on error

	for _, p := range cfg.Products {
		log.Printf("Syncing product: %s", p.Name)
		if _, err := tx.Exec("INSERT OR IGNORE INTO products (id, name) VALUES (?, ?)", p.ID, p.Name); err != nil {
			return fmt.Errorf("failed to insert product: %w", err)
		}

		for _, s := range p.Servers {
			log.Printf("Syncing server: %s for product %s", s.Name, p.Name)

			sshKeyPath := s.SSHKeyPath
			if sshKeyPath == "" {
				sshKeyPath = p.Global.SSHKeyPath
			}

			serverData := store.Server{
				ProductID:  p.ID,
				ID:         s.ID,
				Name:       s.Name,
				Host:       s.Host,
				Port:       s.Port,
				User:       s.User,
				Sudo:       s.Sudo,
				SSHKeyPath: sshKeyPath,
			}
			if _, err := tx.Exec(
				"INSERT OR REPLACE INTO servers (product_id, id, name, host, port, user, sudo, ssh_key_path) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				serverData.ProductID, serverData.ID, serverData.Name, serverData.Host, serverData.Port, serverData.User, serverData.Sudo, serverData.SSHKeyPath,
			); err != nil {
				return fmt.Errorf("failed to insert server: %w", err)
			}

			effectiveFiles, err := cfg.GetEffectiveFiles(p.ID, s.ID)
			if err != nil {
				return fmt.Errorf("failed to get effective files: %w", err)
			}

			for _, f := range effectiveFiles {
				log.Printf("Syncing file: %s for server %s", f.Dest, s.Name)
				fileID := fmt.Sprintf("%s:%s:%s", p.ID, s.ID, f.Dest)
				if _, err := tx.Exec(
					"INSERT OR REPLACE INTO monitored_files (id, product_id, server_id, dest_path, repo_rel_path) VALUES (?, ?, ?, ?, ?)",
					fileID, p.ID, s.ID, f.Dest, f.RepoRelPath,
				); err != nil {
					return fmt.Errorf("failed to insert monitored file: %w", err)
				}
			}
		}
	}

	return tx.Commit()
}
