# Build stage
FROM golang:1.22 AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y gcc libsqlite3-dev

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# CGO_ENABLED=1 is required for the go-sqlite3 driver
RUN CGO_ENABLED=1 go build -o gitoopsoverssh cmd/gitoopsoverssh/main.go

# Final stage
FROM debian:12-slim

WORKDIR /root/

# Install runtime dependencies and clean up to reduce image size
RUN apt-get update && \
    apt-get install -y --no-install-recommends sqlite3 && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=builder /app/gitoopsoverssh .

# Copy the config file and migrations. The application will create the database file.
COPY config.yml .
COPY internal/store/migrations ./internal/store/migrations

# Command to run the application
CMD ["./gitoopsoverssh"]
