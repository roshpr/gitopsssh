.PHONY: build run docker-build docker-run clean help

APP_NAME=gitoopsoverssh
SRC=cmd/gitoopsoverssh/main.go

build:
	CGO_ENABLED=1 go build -o $(APP_NAME) $(SRC)

run:
	./$(APP_NAME)

docker-build:
	docker build -t $(APP_NAME) .

docker-run:
	docker run -it -p 8080:8080 $(APP_NAME)

clean:
	rm -f $(APP_NAME)

help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  build:         Build the application"
	@echo "  run:           Run the application"
	@echo "  docker-build:  Build the Docker image"
	@echo "  docker-run:    Run the Docker container"
	@echo "  clean:         Remove the built application"
	@echo "  help:          Show this help message"
