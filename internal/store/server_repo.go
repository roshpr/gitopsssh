package store

import (
	"database/sql"
	"fmt"
	"time"

	"mymodule/internal/config"
	"mymodule/internal/domain"

	_ "github.com/mattn/go-sqlite3"
)

type ServerRepo struct {
	db *sql.DB
}

func NewServerRepo(db *sql.DB) *ServerRepo {
	return &ServerRepo{db: db}
}

func (r *ServerRepo) Sync(products []config.ProductConfig) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Keep track of active monitored files
	activeMonitoredFiles := make(map[string]bool)

	for _, product := range products {
		// Sync product
		_, err := tx.Exec("INSERT INTO products (id, name) VALUES (?, ?) ON CONFLICT(id) DO UPDATE SET name = excluded.name", product.ID, product.Name)
		if err != nil {
			return fmt.Errorf("failed to sync product %s: %w", product.ID, err)
		}

		for _, server := range product.Servers {
			// Sync server
			_, err := tx.Exec(`
				INSERT INTO servers (id, product_id, name, host, port, user, sudo, ssh_key_path, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(id) DO UPDATE SET
					product_id = excluded.product_id,
					name = excluded.name,
					host = excluded.host,
					port = excluded.port,
					user = excluded.user,
					sudo = excluded.sudo,
					ssh_key_path = excluded.ssh_key_path,
					updated_at = excluded.updated_at
			`, server.ID, product.ID, server.Name, server.Host, server.Port, server.User, server.Sudo, server.SSHKeyPath, time.Now(), time.Now())
			if err != nil {
				return fmt.Errorf("failed to sync server %s: %w", server.ID, err)
			}

			// Sync monitored files for the server
			monitoredFiles := r.determineMonitoredFiles(product, server)
			for _, file := range monitoredFiles {
				monitoredFileID := fmt.Sprintf("%s:%s:%s", product.ID, server.ID, file.Dest)
				_, err := tx.Exec(`
					INSERT INTO monitored_files (id, product_id, server_id, dest_path, repo_rel_path, enabled, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)
					ON CONFLICT(id) DO UPDATE SET
						dest_path = excluded.dest_path,
						repo_rel_path = excluded.repo_rel_path,
						enabled = TRUE,
						updated_at = excluded.updated_at
				`, monitoredFileID, product.ID, server.ID, file.Dest, file.RepoRelPath, true, time.Now(), time.Now())
				if err != nil {
					return fmt.Errorf("failed to sync monitored file %s: %w", monitoredFileID, err)
				}
				activeMonitoredFiles[monitoredFileID] = true
			}
		}
	}

	// Disable monitored files that are no longer in the config
	rows, err := tx.Query("SELECT id FROM monitored_files WHERE enabled = TRUE")
	if err != nil {
		return fmt.Errorf("failed to query monitored files: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan monitored file id: %w", err)
		}
		if !activeMonitoredFiles[id] {
			_, err := tx.Exec("UPDATE monitored_files SET enabled = FALSE, updated_at = ? WHERE id = ?", time.Now(), id)
			if err != nil {
				return fmt.Errorf("failed to disable monitored file %s: %w", id, err)
			}
		}
	}

	return tx.Commit()
}

func (r *ServerRepo) GetEnabledServers() ([]domain.Server, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.product_id, s.name, s.host, s.port, s.user, s.sudo, s.ssh_key_path, s.last_poll_at, s.last_error, s.created_at, s.updated_at
		FROM servers s
		JOIN products p ON s.product_id = p.id
		WHERE EXISTS (
			SELECT 1 FROM monitored_files mf WHERE mf.server_id = s.id AND mf.enabled = TRUE
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled servers: %w", err)
	}
	defer rows.Close()

	var servers []domain.Server
	for rows.Next() {
		var s domain.Server
		if err := rows.Scan(&s.ID, &s.ProductID, &s.Name, &s.Host, &s.Port, &s.User, &s.Sudo, &s.SSHKeyPath, &s.LastPollAt, &s.LastError, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}

		monitoredFiles, err := r.getMonitoredFilesForServer(s.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get monitored files for server %s: %w", s.ID, err)
		}
		s.MonitoredFiles = monitoredFiles

		servers = append(servers, s)
	}

	return servers, nil
}

func (r *ServerRepo) getMonitoredFilesForServer(serverID string) ([]domain.MonitoredFile, error) {
	rows, err := r.db.Query(`
		SELECT id, product_id, server_id, dest_path, repo_rel_path, enabled, created_at, updated_at
		FROM monitored_files
		WHERE server_id = ? AND enabled = TRUE
	`, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query monitored files: %w", err)
	}
	defer rows.Close()

	var files []domain.MonitoredFile
	for rows.Next() {
		var f domain.MonitoredFile
		if err := rows.Scan(&f.ID, &f.ProductID, &f.ServerID, &f.DestPath, &f.RepoRelPath, &f.Enabled, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan monitored file: %w", err)
		}
		files = append(files, f)
	}

	return files, nil
}

func (r *ServerRepo) determineMonitoredFiles(product config.ProductConfig, server config.ServerConfig) []config.FileConfig {
	filesToMonitor := make(map[string]config.FileConfig)

	// Add global files first
	for _, file := range product.Global.Files {
		filesToMonitor[file.Dest] = file
	}

	// Add server-specific files, they will override global files if destination is the same
	for _, file := range server.Files {
		filesToMonitor[file.Dest] = file
	}

	// Remove excluded files
	for _, excludedPath := range server.ExcludeFiles {
		delete(filesToMonitor, excludedPath)
	}

	// Convert map to slice
	result := make([]config.FileConfig, 0, len(filesToMonitor))
	for _, file := range filesToMonitor {
		result = append(result, file)
	}

	return result
}
