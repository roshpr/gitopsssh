# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# CGO_ENABLED=1 is required for the go-sqlite3 driver
RUN CGO_ENABLED=1 go build -o gitoopsoverssh cmd/gitoopsoverssh/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

# Install runtime dependencies
RUN apk add --no-cache sqlite

# Copy the binary from the builder stage
COPY --from=builder /app/gitoopsoverssh .

# Copy the config file and migrations. The application will create the database file.
COPY config.yml .
COPY internal/store/migrations ./internal/store/migrations

# Command to run the application
CMD ["./gitoopsoverssh"]
