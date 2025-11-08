.PHONY: help start-dynamo stop-dynamo create-table seed-data run-local clean-local test-local cli-build cli-run cli-install cli-uninstall test coverage
# CLI tool commands
cli-build:
	@echo "Building CLI tool..."
	go build -o bin/create-go-service cmd/create-go-service/main.go
	@ln -sf create-go-service bin/cgs
	@echo "✓ Built create-go-service and alias cgs"

cli-run:
	@echo "Running CLI tool..."
	go run cmd/create-go-service/main.go

cli-install:
	@echo "Installing create-go-service to system..."
	@go build -o bin/create-go-service cmd/create-go-service/main.go
	@if [ -w /usr/local/bin ]; then \
		cp bin/create-go-service /usr/local/bin/create-go-service; \
		ln -sf create-go-service /usr/local/bin/cgs; \
		echo "✓ Installed to /usr/local/bin/create-go-service"; \
		echo "✓ Created alias /usr/local/bin/cgs"; \
	else \
		echo "Installing to /usr/local/bin (requires sudo)..."; \
		sudo cp bin/create-go-service /usr/local/bin/create-go-service; \
		sudo ln -sf create-go-service /usr/local/bin/cgs; \
		echo "✓ Installed to /usr/local/bin/create-go-service"; \
		echo "✓ Created alias /usr/local/bin/cgs"; \
	fi
	@echo "You can now run 'create-go-service' or 'cgs' from anywhere!"

cli-uninstall:
	@echo "Uninstalling create-go-service from system..."
	@if [ -f /usr/local/bin/create-go-service ]; then \
		if [ -w /usr/local/bin ]; then \
			rm -f /usr/local/bin/create-go-service /usr/local/bin/cgs; \
		else \
			sudo rm -f /usr/local/bin/create-go-service /usr/local/bin/cgs; \
		fi; \
		echo "✓ Uninstalled from /usr/local/bin/create-go-service and /usr/local/bin/cgs"; \
	else \
		echo "create-go-service is not installed"; \
	fi

test:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "Coverage report generated: coverage.out"
	@echo "Run 'make coverage' to view the coverage report"

coverage: test
	@echo "Coverage report:"
	@go tool cover -func=coverage.out | tail -1
	@echo ""
	@echo "Run 'go tool cover -html=coverage.out' to view detailed HTML report"

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
	@echo ""
	@echo "CLI Tool:"
	@echo "  cli-build       - Build the create-go-service CLI tool (with cgs alias)"
	@echo "  cli-run         - Run the create-go-service CLI tool"
	@echo "  cli-install     - Install create-go-service to /usr/local/bin (with cgs alias)"
	@echo "  cli-uninstall   - Uninstall create-go-service from system"
	@echo ""
	@echo "Testing:"
	@echo "  test            - Run tests with coverage"
	@echo "  coverage        - View coverage report"
	@echo ""
	@echo "Local Development:"
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
	docker compose up -d dynamodb-local
	@echo "DynamoDB Local is running on http://localhost:8000"

# Stop DynamoDB Local
stop-dynamo:
	@echo "Stopping DynamoDB Local..."
	docker compose down

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
	docker compose down
	docker compose up -d dynamodb-local
	@echo "Local DynamoDB data cleaned"

# Run tests against local DynamoDB
test-local:
	@echo "Running tests against local DynamoDB..."
	DYNAMODB_ENDPOINT=http://localhost:8000 TABLE_NAME=notes-test go test ./...

# Full local setup
setup-local: start-dynamo create-table seed-data
	@echo "Local development environment is ready!"
	@echo "Run 'make run-local' to start the application"