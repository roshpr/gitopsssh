package poller

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"mymodule/internal/config"
	"mymodule/internal/git"
	"mymodule/internal/ssh"
	"mymodule/internal/store"
)

func Poll(ctx context.Context, cfg *config.Config, db *sql.DB) {
	ticker := time.NewTicker(time.Duration(cfg.Polling.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("Polling all servers...")
			servers, err := store.GetAllServers(db)
			if err != nil {
				log.Printf("Error getting servers: %v", err)
				continue
			}

			var wg sync.WaitGroup
			limiter := make(chan struct{}, cfg.Polling.MaxConcurrency)

			for _, server := range servers {
				wg.Add(1)
				limiter <- struct{}{}

				go func(server store.Server) {
					defer wg.Done()
					defer func() { <-limiter }()

					log.Printf("Polling server: %s", server.Name)
					if err := pollServer(db, cfg, server); err != nil {
						log.Printf("Error polling server %s: %v", server.Name, err)
					}
				}(server)
			}

			wg.Wait()
			log.Println("Polling finished.")
		}
	}
}

func pollServer(db *sql.DB, cfg *config.Config, server store.Server) error {
	serverConfig, err := cfg.GetServerConfig(server.ProductID, server.ID)
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}

	files, err := store.GetMonitoredFilesForServer(db, server.ProductID, server.ID)
	if err != nil {
		return fmt.Errorf("failed to get monitored files: %w", err)
	}

	sshClient, err := ssh.NewClient(server.Host, server.Port, server.User, serverConfig.SSHKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create ssh client: %w", err)
	}
	defer sshClient.Close()

	gitManager := git.NewManager(cfg.Git.RepoPath)

	for _, file := range files {
		log.Printf("Checking file: %s on server %s", file.DestPath, server.Name)

		state := store.FileState{
			ID:              fmt.Sprintf("%s:%s", file.ID, time.Now().UTC().Format(time.RFC3339Nano)),
			MonitoredFileID: file.ID,
			LastCheckedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		}

		desiredHash, err := gitManager.GetFileHash(file.RepoRelPath)
		if err != nil {
			log.Printf("Error getting desired hash for %s: %v", file.RepoRelPath, err)
			state.Status = "error"
			state.Error = sql.NullString{String: err.Error(), Valid: true}
			if err := store.UpdateFileState(db, state); err != nil {
				log.Printf("Error updating file state: %v", err)
			}
			continue
		}

		remoteHash, err := ssh.GetFileHash(sshClient, file.DestPath, server.Sudo)
		if err != nil {
			log.Printf("Error getting remote hash for %s on %s: %v", file.DestPath, server.Name, err)
			state.Status = "error"
			state.Error = sql.NullString{String: err.Error(), Valid: true}
			if err := store.UpdateFileState(db, state); err != nil {
				log.Printf("Error updating file state: %v", err)
			}
			continue
		}

		if desiredHash != remoteHash {
			state.Status = "drifted"
			// TODO: Generate and store diff
		} else {
			state.Status = "in_sync"
		}

		log.Printf("File %s on server %s is %s", file.DestPath, server.Name, state.Status)

		if err := store.UpdateFileState(db, state); err != nil {
				log.Printf("Error updating file state: %v", err)
		}
	}

	return nil
}
