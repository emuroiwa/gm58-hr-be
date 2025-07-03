.PHONY: build run test clean docker-up docker-down migrate-up migrate-down

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary name
BINARY_NAME=gm58-hr-backend
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v cmd/server/main.go

# Run the application
run:
	$(GOCMD) run cmd/server/main.go

# Test the application
test:
	$(GOTEST) -v ./...

# Test with coverage
test-coverage:
	$(GOTEST) -cover ./...

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v cmd/server/main.go

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

docker-logs:
	docker-compose logs -f backend

# Database migrations
migrate-up:
	go run scripts/migrate.go -cmd=up

migrate-down:
	go run scripts/migrate.go -cmd=down

migrate-status:
	go run scripts/migrate.go -cmd=status

# Development setup
setup-dev:
	./scripts/setup-dev.sh

# Format code
fmt:
	go fmt ./...

# Run all checks
check: fmt test

# Check environment configuration
check-env:
	$(GOCMD) run scripts/check-env.go

# Test database connection
test-db:
	$(GOCMD) run scripts/test-db.go

# Verify complete setup
verify: check-env test-db migrate-status
	@echo "✅ All checks passed!"

# Show help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Clean build files"
	@echo "  deps           - Download dependencies"
	@echo "  docker-up      - Start Docker containers"
	@echo "  docker-down    - Stop Docker containers"
	@echo "  migrate-up     - Run database migrations"
	@echo "  migrate-down   - Rollback database migrations"
	@echo "  setup-dev      - Setup development environment"
	@echo "  fmt            - Format code"
	@echo "  check          - Run all checks"
	@echo "  help           - Show this help"
