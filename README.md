# gitoopsOverSsh

`gitoopsOverSsh` is a Go-based tool for monitoring file integrity on remote servers. It ensures that files on your servers match the versions stored in a central Git repository.

## Architecture

The application is composed of the following main components:

*   **Poller:** Periodically checks the file integrity on each server.
*   **Git Manager:** Manages the local clone of the Git repository.
*   **SSH Client:** Connects to remote servers to check file hashes.
*   **Store:** A SQLite database that stores the application's state, including servers, monitored files, and file state history.
*   **HTTP Server:** Provides a simple web-based UI to view the status of monitored files.

## Code Structure

```
.
├── cmd
│   └── gitoopsoverssh
│       └── main.go         # Main application entry point
├── internal
│   ├── config
│   │   └── config.go       # Configuration loading
│   ├── git
│   │   ├── cloner.go       # Git repository cloning
│   │   └── manager.go      # Git repository management
│   ├── http
│   │   └── server.go       # HTTP server and UI
│   ├── poller
│   │   └── poller.go       # File integrity polling logic
│   ├── ssh
│   │   └── client.go       # SSH client for remote commands
│   └── store
│       ├── db.go           # Database initialization and migrations
│       ├── files_repo.go   # Monitored file repository
│       ├── models.go       # Database models
│       └── server_repo.go  # Server repository
├── config.yml              # Application configuration
└── README.md
```

## What we are doing

`gitoopsOverSsh` addresses the problem of configuration drift in a server infrastructure. It ensures that the files on your servers are in the desired state, as defined by a Git repository. This is achieved by:

1.  **Cloning a Git repository:** The application maintains a local clone of a Git repository that contains the desired state of your monitored files.
2.  **Polling servers:** It periodically connects to each configured server via SSH.
3.  **Comparing file hashes:** For each monitored file, it calculates the hash of the local version (from the Git repository) and the remote version (on the server). 
4.  **Reporting status:** If the hashes don't match, the file is marked as "drifted." The status of all monitored files is displayed in a web UI.

## How to set up the environment

### Development

1.  **Install Go:** Make sure you have Go installed and configured on your system.
2.  **Install SQLite:** `gitoopsOverSsh` uses SQLite as its database. Install it using your system's package manager:
    ```bash
    sudo apt-get update && sudo apt-get install -y sqlite3
    ```
3.  **Configure `config.yml`:** See the `config.yml` details section below.
4.  **Run the application:**
    ```bash
    CGO_ENABLED=1 go run cmd/gitoopsoverssh/main.go
    ```

### Deployment

1.  **Build the binary:**
    ```bash
    CGO_ENABLED=1 go build -o gitoopsoverssh cmd/gitoopsoverssh/main.go
    ```
2.  **Deploy the binary and `config.yml`:** Copy the `gitoopsoverssh` binary and your `config.yml` file to your server.
3.  **Run the application:**
    ```bash
    ./gitoopsoverssh
    ```

## How to view the UI

Once the application is running, you can view the UI by opening your web browser and navigating to `http://localhost:8080` (or the appropriate address if you're running it on a remote server).

## `config.yml` details

The `config.yml` file is the central configuration for the application. Here's a breakdown of its structure:

```yaml
git:
  remote: "https://github.com/user/repo.git"
  branch: "main"
  repoPath: "/tmp/gitoops_repo"
  refreshIntervalSeconds: 300

polling:
  intervalSeconds: 60
  maxConcurrency: 10

products:
  - id: "product-a"
    name: "Product A"
    global:
      sshKeyPath: "/path/to/your/ssh/key"
    servers:
      - id: "server-1"
        name: "Server 1"
        host: "1.2.3.4"
        port: 22
        user: "user"
        sudo: true
    files:
      - dest: "/etc/nginx/nginx.conf"
        repoRelPath: "nginx/nginx.conf"
```

*   **`git`:** Configures the Git repository to be monitored.
    *   `remote`: The URL of the Git repository.
    *   `branch`: The branch to be monitored.
    *   `repoPath`: The local path where the repository will be cloned.
    *   `refreshIntervalSeconds`: How often to pull the latest changes from the remote repository.
*   **`polling`:** Configures the file polling behavior.
    *   `intervalSeconds`: How often to check the file integrity on each server.
    *   `maxConcurrency`: The maximum number of servers to poll concurrently.
*   **`products`:** A list of products, each with its own set of servers and files to be monitored.
    *   `id`: A unique identifier for the product.
    *   `name`: A human-readable name for the product.
    *   `global`: Global settings for the product.
        *   `sshKeyPath`: The default path to the SSH key to be used for all servers in this product.
    *   `servers`: A list of servers to be monitored for this product.
        *   `id`: A unique identifier for the server.
        *   `name`: A human-readable name for the server.
        *   `host`: The hostname or IP address of the server.
        *   `port`: The SSH port on the server.
        *   `user`: The SSH user.
        *   `sudo`: Whether to use `sudo` when checking file hashes.
    *   `files`: A list of files to be monitored for this product.
        *   `dest`: The absolute path to the file on the remote server.
        *   `repoRelPath`: The relative path to the file in the Git repository.
