# GitOps over SSH

This application implements a GitOps workflow over SSH. It periodically polls a set of servers, checks for file content drift, and corrects it.

## Core Functionality

The application performs the following core functions:

*   **Regular Polling:** It polls all configured servers at a regular interval.
*   **File Content Hashing:** For each monitored file, it calculates the SHA256 hash of the content on the remote server.
*   **Drift Detection:** It compares the hash of the remote file with the hash of the file in the local Git repository.
*   **Drift Correction:** If the hashes don't match, it overwrites the remote file with the content from the local Git repository.
*   **State Logging:** It logs the state of each file (in_sync, drift_detected, drift_corrected, error) to a local SQLite database.

## Architecture

The application is composed of the following components:

*   **Main Application (`cmd/gitoopsoverssh/main.go`):** The entry point of the application. It initializes the configuration, database, and poller, and starts the HTTP server.
*   **Configuration (`internal/config/config.go`):** Loads the application configuration from a YAML file (`config.yml`).
*   **Database (`internal/store/db.go`):** Manages the SQLite database connection and runs schema migrations.
*   **Git Manager (`internal/git/manager.go`):** Manages the local Git repository, including cloning and pulling updates.
*   **SSH Client (`internal/ssh/client.go`):** Manages SSH connections to the remote servers.
*   **Poller (`internal/poller/poller.go`):** The core of the application. It orchestrates the polling process, including drift detection and correction.
*   **HTTP Server (`internal/http/server.go`):** Exposes an API for querying the status of the monitored files.

## Code Flow

The following steps describe a single polling cycle:

1.  **The `Poll` function in `internal/poller/poller.go` is called.**
2.  **The poller retrieves a list of all servers from the database.**
3.  **For each server, the poller retrieves a list of all monitored files.**
4.  **For each monitored file, the poller performs the following steps:**
    *   It establishes an SSH connection to the remote server.
    *   It calculates the SHA256 hash of the remote file.
    *   It retrieves the content of the file from the local Git repository.
    *   It calculates the SHA256 hash of the local file.
    *   It compares the two hashes.
    *   If the hashes are different, it overwrites the remote file with the content of the local file.
    *   It updates the file state in the database.

## Configuration

The application is configured using a YAML file (`config.yml`). The following options are available:

*   **`repo_url`:** The URL of the Git repository to clone.
*   **`repo_path`:** The local path where the Git repository will be cloned.
*   **`ssh_key_path`:** The path to the SSH private key to use for cloning the Git repository.
*   **`poll_interval_seconds`:** The interval at which to poll the servers.
*   **`servers`:** A list of servers to poll. Each server has the following properties:
    *   **`name`:** The name of the server.
    *   **`host`:** The hostname or IP address of the server.
    *   **`port`:** The SSH port of the server.
    *   **`user`:** The SSH user to use for connecting to the server.
    *   **`ssh_key_path`:** The path to the SSH private key to use for connecting to the server.
    *   **`files`:** A list of files to monitor on the server. Each file has the following properties:
        *   **`repo_path`:** The path to the file in the Git repository.
        *   **`remote_path`:** The path to the file on the remote server.

## Database

The application uses a SQLite database to store the state of the monitored files. The database schema is defined in `internal/store/migrations/0001_initial_schema.sql`.

The following tables are used:

*   **`servers`:** Stores the configuration of the servers to poll.
*   **`monitored_files`:** Stores the configuration of the files to monitor.
*   **`file_states`:** Stores the state of each monitored file.

## Getting Started

To build and run the application, follow these steps:

1.  **Clone the repository.**
2.  **Install the dependencies:** `go mod download`
3.  **Create a `config.yml` file.**
4.  **Run the application:** `go run cmd/gitoopsoverssh/main.go`
