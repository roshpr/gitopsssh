--- file_states table
CREATE TABLE IF NOT EXISTS file_states (
    id TEXT PRIMARY KEY,
    monitored_file_id TEXT NOT NULL,
    status TEXT NOT NULL,
    last_checked_at DATETIME NOT NULL,
    diff TEXT,
    error TEXT,
    FOREIGN KEY (monitored_file_id) REFERENCES monitored_files (id)
);
