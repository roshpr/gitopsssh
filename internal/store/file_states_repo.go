package store

import (
	"database/sql"
	"fmt"
)

func UpdateFileState(db *sql.DB, state FileState) error {
	_, err := db.Exec(`
		INSERT INTO file_states (monitored_file_id, status, last_checked_at, diff, error)
		VALUES (?, ?, ?, ?, ?)
	`, state.MonitoredFileID, state.Status, state.LastCheckedAt, state.Diff, state.Error)
	if err != nil {
		return fmt.Errorf("failed to insert file state: %w", err)
	}
	return nil
}

func GetAllFileStates(db *sql.DB) ([]FileState, error) {
	rows, err := db.Query(`
		SELECT id, monitored_file_id, status, last_checked_at, diff, error
		FROM file_states
		ORDER BY last_checked_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query file states: %w", err)
	}
	defer rows.Close()

	var states []FileState
	for rows.Next() {
		var s FileState
		if err := rows.Scan(&s.ID, &s.MonitoredFileID, &s.Status, &s.LastCheckedAt, &s.Diff, &s.Error); err != nil {
			return nil, fmt.Errorf("failed to scan file state row: %w", err)
		}
		states = append(states, s)
	}

	return states, nil
}
