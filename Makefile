.PHONY: help start-dynamo stop-dynamo create-table seed-data run-local clean-local test-local
# Go commands
build:
	@echo "Building application..."
	go build -o bin/api cmd/api/main.go

run:
	@echo "Running application..."
	go run cmd/api/main.go

test:
	@echo "Running tests..."
	go test -v ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Vetting code..."
	go vet ./...

lint:
	@echo "Linting code..."
	golangci-lint run

# Default target
help:
	@echo "Available commands:"
	@echo "  start-dynamo    - Start DynamoDB Local container"
	@echo "  stop-dynamo     - Stop DynamoDB Local container"
	@echo "  create-table    - Create DynamoDB table"
	@echo "  seed-data       - Seed sample data"
	@echo "  run-local       - Run the application locally"
	@echo "  clean-local     - Clean up local DynamoDB data"
	@echo "  test-local      - Run tests against local DynamoDB"

# Start DynamoDB Local
start-dynamo:
	@echo "Starting DynamoDB Local..."
	docker-compose up -d dynamodb-local
	@echo "DynamoDB Local is running on http://localhost:8000"

# Stop DynamoDB Local
stop-dynamo:
	@echo "Stopping DynamoDB Local..."
	docker-compose down

# Create table
create-table:
	@echo "Creating DynamoDB table..."
	go run scripts/create-table.go

# Seed sample data
seed-data:
	@echo "Seeding sample data..."
	go run scripts/seed-data.go

# Run application locally
run-local:
	@echo "Running application locally..."
	DYNAMODB_ENDPOINT=http://localhost:8000 TABLE_NAME=notes-local go run main.go

# Clean local data (restart container)
clean-local:
	@echo "Cleaning local DynamoDB data..."
	docker-compose down
	docker-compose up -d dynamodb-local
	@echo "Local DynamoDB data cleaned"

# Run tests against local DynamoDB
test-local:
	@echo "Running tests against local DynamoDB..."
	DYNAMODB_ENDPOINT=http://localhost:8000 TABLE_NAME=notes-test go test ./...

# Full local setup
setup-local: start-dynamo create-table seed-data
	@echo "Local development environment is ready!"
	@echo "Run 'make run-local' to start the application"