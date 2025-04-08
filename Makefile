.PHONY: build run test clean docker docker-compose

# Default build target
build:
	go build -o cereja-corp ./cmd/main.go

# Run the application
run:
	go run ./cmd/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f cereja-corp

# Build Docker image
docker:
	docker build -t cereja-corp .

# Run with Docker Compose
docker-compose:
	docker-compose up -d

# Stop Docker Compose services
docker-compose-down:
	docker-compose down

# Generate Go docs
docs:
	godoc -http=:6060

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golint ./...

# Tidy Go modules
tidy:
	go mod tidy

# Help command
help:
	@echo "Available commands:"
	@echo "  make build            - Build the application"
	@echo "  make run              - Run the application"
	@echo "  make test             - Run tests"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make docker           - Build Docker image"
	@echo "  make docker-compose   - Run with Docker Compose"
	@echo "  make docker-compose-down - Stop Docker Compose services"
	@echo "  make docs             - Generate Go docs"
	@echo "  make fmt              - Format code"
	@echo "  make lint             - Lint code"
	@echo "  make tidy             - Tidy Go modules"

# Migrate database
migrate:
	@echo "Migrating database..."
	psql -U $(DB_USER) -h $(DB_HOST) -d $(DB_NAME) -f internal/receipts/migrations/001_create_receipts_tables.sql

# Run this rule to initialize database for receipt scanner
init-receipts: migrate
	@echo "Receipt scanner database initialized" 