CREATE TABLE products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE servers (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    user TEXT NOT NULL,
    sudo BOOLEAN NOT NULL,
    ssh_key_path TEXT NOT NULL,
    last_poll_at DATETIME,
    last_error TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE TABLE monitored_files (
    id TEXT PRIMARY KEY, -- sha256(product_id:server_id:dest)
    product_id TEXT NOT NULL,
    server_id TEXT NOT NULL,
    dest_path TEXT NOT NULL,
    repo_rel_path TEXT NOT NULL,
    enabled BOOLEAN NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE (product_id, server_id, dest_path),
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);

CREATE TABLE file_state (
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
    last_checked_at DATETIME NOT NULL,
    last_drift_at DATETIME,
    error_text TEXT,
    FOREIGN KEY (monitored_file_id) REFERENCES monitored_files(id) ON DELETE CASCADE
);

CREATE TABLE diff_cache (
    monitored_file_id TEXT PRIMARY KEY,
    desired_hash TEXT NOT NULL,
    remote_hash TEXT NOT NULL,
    diff_text TEXT NOT NULL,
    truncated BOOLEAN NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (monitored_file_id) REFERENCES monitored_files(id) ON DELETE CASCADE
);

CREATE TABLE audit_log (
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
    created_at DATETIME NOT NULL,
    FOREIGN KEY (monitored_file_id) REFERENCES monitored_files(id) ON DELETE CASCADE
);

CREATE INDEX idx_monitored_files_server ON monitored_files(server_id);
CREATE INDEX idx_file_state_status ON file_state(status);
CREATE INDEX idx_audit_server ON audit_log(server_id, created_at);
