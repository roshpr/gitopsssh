package store

import (
	"database/sql"
	"fmt"
)

func GetMonitoredFilesForServer(db *sql.DB, productID, serverID string) ([]MonitoredFile, error) {
	rows, err := db.Query("SELECT id, product_id, server_id, dest_path, repo_rel_path FROM monitored_files WHERE product_id = ? AND server_id = ?", productID, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query monitored files: %w", err)
	}
	defer rows.Close()

	var files []MonitoredFile
	for rows.Next() {
		var f MonitoredFile
		if err := rows.Scan(&f.ID, &f.ProductID, &f.ServerID, &f.DestPath, &f.RepoRelPath); err != nil {
			return nil, fmt.Errorf("failed to scan monitored file row: %w", err)
		}
		files = append(files, f)
	}

	return files, nil
}
