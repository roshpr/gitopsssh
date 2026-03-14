-- 001_initial_schema.sql

CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS servers (
    product_id TEXT NOT NULL,
    id TEXT NOT NULL,
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    user TEXT NOT NULL,
    sudo INTEGER NOT NULL,
    last_poll_at DATETIME,
    last_error TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (product_id, id),
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS monitored_files (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    server_id TEXT NOT NULL,
    dest_path TEXT NOT NULL,
    repo_rel_path TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (product_id, server_id, dest_path),
    FOREIGN KEY (product_id, server_id) REFERENCES servers(product_id, id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS file_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    monitored_file_id TEXT NOT NULL,
    status TEXT NOT NULL,
    desired_hash TEXT,
    remote_hash TEXT,
    desired_size INTEGER,
    remote_size INTEGER,
    remote_mode TEXT,
    remote_owner TEXT,
    remote_group TEXT,
    last_checked_at DATETIME,
    last_drift_at DATETIME,
    error_text TEXT,
    FOREIGN KEY (monitored_file_id) REFERENCES monitored_files(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS diff_cache (
    monitored_file_id TEXT PRIMARY KEY,
    desired_hash TEXT NOT NULL,
    remote_hash TEXT NOT NULL,
    diff_text TEXT NOT NULL,
    truncated INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (monitored_file_id) REFERENCES monitored_files(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    actor TEXT NOT NULL,
    action TEXT NOT NULL,
    server_id TEXT NOT NULL,
    monitored_file_id TEXT NOT NULL,
    dest_path TEXT NOT NULL,
    before_hash TEXT,
    after_hash TEXT,
    git_commit_sha TEXT,
    result TEXT NOT NULL,
    details TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_monitored_files_server ON monitored_files(product_id, server_id);
CREATE INDEX IF NOT EXISTS idx_file_state_status ON file_state(status);
CREATE INDEX IF NOT EXISTS idx_audit_server ON audit_log(server_id, created_at);
