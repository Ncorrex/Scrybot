.PHONY: help build run docker-build docker-run clean test lint deps ui-dev ui-build

help:
	@echo "Scrybot - Scryfall Alert Microservice"
	@echo ""
	@echo "Available commands:"
	@echo "  make build        - Build the Go application"
	@echo "  make run          - Run the application (requires DISCORD_WEBHOOK_URL env var)"
	@echo "  make docker-build - Build the Docker image"
	@echo "  make docker-run   - Run the Docker container"
	@echo "  make clean        - Remove build artifacts and compiled binary"
	@echo "  make test         - Run Go tests (if any)"
	@echo "  make lint         - Run Go linter"
	@echo "  make deps         - Download and verify dependencies"
	@echo "  make all          - Build everything (compile, lint, test)
  make ui-dev       - Start the Vue dev server (proxies API to :8080)
  make ui-build     - Build the Vue frontend into ui/dist"

ui-build:
	@echo "Building UI..."
	cd ui && npm install && npm run build
	@echo "✓ UI build complete: ui/dist"

ui-dev:
	@echo "Starting Vue dev server (proxy → :8080)..."
	cd ui && npm run dev

build: ui-build
	@echo "Building scrybot..."
	go build -o scrybot .
	@echo "✓ Build complete: ./scrybot"

run: build
	@if [ -z "$$DISCORD_WEBHOOK_URL" ]; then \
		echo "Error: DISCORD_WEBHOOK_URL environment variable not set"; \
		exit 1; \
	fi
	@echo "Starting Scrybot..."
	./scrybot

docker-build:
	@echo "Building Docker image..."
	docker build -t scrybot:latest .
	@echo "✓ Docker image built: scrybot:latest"

docker-run: docker-build
	@if [ -z "$$DISCORD_WEBHOOK_URL" ]; then \
		echo "Error: DISCORD_WEBHOOK_URL environment variable not set"; \
		exit 1; \
	fi
	@echo "Starting Docker container..."
	docker run -d \
		--name scrybot \
		-e DISCORD_WEBHOOK_URL="$$DISCORD_WEBHOOK_URL" \
		-e POLL_INTERVAL="1h" \
		-v scrybot-data:/root/data \
		scrybot:latest
	@echo "✓ Container running. View logs: docker logs -f scrybot"

clean:
	@echo "Cleaning build artifacts..."
	rm -f scrybot
	@echo "✓ Clean complete"

test:
	@echo "Running tests..."
	go test -v ./...

lint:
	@echo "Running linter..."
	@if ! which golangci-lint > /dev/null; then \
		echo "golangci-lint not installed. Install with: brew install golangci-lint"; \
		exit 1; \
	fi
	golangci-lint run ./...

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify
	@echo "✓ Dependencies verified"

all: deps lint build test
	@echo "✓ All checks passed"

# Development helpers
dev-run:
	POLL_INTERVAL=5m go run .

dev-watch:
	@if ! which entr > /dev/null; then \
		echo "entr not installed. Install with: brew install entr"; \
		exit 1; \
	fi
	find . -name '*.go' | entr make build

