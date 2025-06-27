# Makefile for Internal Transfers Application

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Application parameters
BINARY_NAME=internal-transfers
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=./cmd/server
DOCKER_IMAGE=internal-transfers

# Test parameters
COVERAGE_DIR=coverage
COVERAGE_FILE=$(COVERAGE_DIR)/coverage.out
COVERAGE_HTML=$(COVERAGE_DIR)/coverage.html

# Database parameters
DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=internal_transfers
TEST_DB_NAME=internal_transfers_test
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
TEST_DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(TEST_DB_NAME)?sslmode=disable

.PHONY: all build clean test test-unit test-integration test-coverage deps fmt vet lint run dev docker-build docker-run docker-stop setup-db migrate help

# Default target
all: clean deps test build

# Help target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the application
build: ## Build the application binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) -v $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

# Clean build artifacts
clean: ## Clean build artifacts and temporary files
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf bin/
	@rm -rf $(COVERAGE_DIR)/
	@rm -f *.prof
	@echo "Clean complete"

# Install dependencies
deps: ## Download and install dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

# Format code
fmt: ## Format Go source code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "Code formatted"

# Vet code
vet: ## Run go vet on the codebase
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "Vet complete"

# Lint code (requires golangci-lint)
lint: ## Run golangci-lint on the codebase
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
	fi

# Run all tests
test: test-unit test-integration ## Run all tests

# Run unit tests
test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) ./tests/unit/... -v -race -coverprofile=$(COVERAGE_DIR)/unit-coverage.out

# Run integration tests
test-integration: ## Run integration tests (requires database)
	@echo "Running integration tests..."
	@mkdir -p $(COVERAGE_DIR)
	TEST_DATABASE_URL=$(TEST_DB_URL) RUN_INTEGRATION_TESTS=true \
	$(GOTEST) ./tests/integration/... -v -race -coverprofile=$(COVERAGE_DIR)/integration-coverage.out

# Generate test coverage report
test-coverage: test ## Generate and display test coverage report
	@echo "Generating coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	@echo "mode: atomic" > $(COVERAGE_FILE)
	@if [ -f $(COVERAGE_DIR)/unit-coverage.out ]; then \
		tail -n +2 $(COVERAGE_DIR)/unit-coverage.out >> $(COVERAGE_FILE); \
	fi
	@if [ -f $(COVERAGE_DIR)/integration-coverage.out ]; then \
		tail -n +2 $(COVERAGE_DIR)/integration-coverage.out >> $(COVERAGE_FILE); \
	fi
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep total
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Run tests with custom script
test-script: ## Run tests using the custom test script
	@chmod +x scripts/run-tests.sh
	./scripts/run-tests.sh -t all -c -v

# Benchmark tests
benchmark: ## Run benchmark tests
	@echo "Running benchmark tests..."
	$(GOTEST) -bench=. -benchmem ./tests/...

# Run the application
run: build ## Build and run the application
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_PATH)

# Development mode with auto-reload (requires air)
dev: ## Run in development mode with auto-reload
	@echo "Starting development server..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not found. Install it with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to normal run..."; \
		make run; \
	fi

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-run: ## Run application in Docker with docker-compose
	@echo "Starting application with docker-compose..."
	docker-compose up --build

docker-stop: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	docker-compose down

docker-clean: ## Clean Docker images and containers
	@echo "Cleaning Docker images and containers..."
	docker-compose down --rmi all --volumes --remove-orphans

# Database targets
setup-db: ## Set up the main database
	@echo "Setting up database..."
	createdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME) || echo "Database may already exist"
	@echo "Database setup complete"

setup-test-db: ## Set up the test database
	@echo "Setting up test database..."
	createdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(TEST_DB_NAME) || echo "Test database may already exist"
	@echo "Test database setup complete"

migrate: ## Run database migrations
	@echo "Running migrations..."
	psql $(DB_URL) -f migrations/001_create_accounts.sql
	psql $(DB_URL) -f migrations/002_create_transactions.sql
	@echo "Migrations complete"

migrate-test: ## Run database migrations for test database
	@echo "Running test migrations..."
	psql $(TEST_DB_URL) -f migrations/001_create_accounts.sql
	psql $(TEST_DB_URL) -f migrations/002_create_transactions.sql
	@echo "Test migrations complete"

# Quality assurance targets
qa: fmt vet lint test ## Run all quality assurance checks

# Security scan
security: ## Run security vulnerability scan
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install it with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Performance profiling
profile-cpu: ## Run CPU profiling
	@echo "Running CPU profiling..."
	$(GOTEST) -cpuprofile=cpu.prof -bench=. ./tests/...
	$(GOCMD) tool pprof cpu.prof

profile-mem: ## Run memory profiling
	@echo "Running memory profiling..."
	$(GOTEST) -memprofile=mem.prof -bench=. ./tests/...
	$(GOCMD) tool pprof mem.prof

# Release targets
release-build: ## Build release binary with optimizations
	@echo "Building release binary..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) \
		-ldflags="-w -s" -a -installsuffix cgo \
		-o $(BINARY_PATH)-linux-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) \
		-ldflags="-w -s" -a -installsuffix cgo \
		-o $(BINARY_PATH)-darwin-amd64 $(MAIN_PATH)
	@echo "Release binaries built"

# Install tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCMD) install github.com/cosmtrek/air@latest
	$(GOCMD) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "Tools installed"

# Check tools
check-tools: ## Check if required tools are installed
	@echo "Checking required tools..."
	@command -v go >/dev/null 2>&1 || { echo "go is required but not installed"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "docker is required but not installed"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "docker-compose is required but not installed"; exit 1; }
	@command -v psql >/dev/null 2>&1 || { echo "psql is recommended but not installed"; }
	@echo "Tool check complete"

# Environment setup
setup: check-tools install-tools setup-db setup-test-db migrate migrate-test deps ## Complete environment setup

# CI/CD targets
ci: qa test-coverage ## Run CI pipeline locally

# Show project status
status: ## Show project status and information
	@echo "Project Status:"
	@echo "  Go version: $$(go version)"
	@echo "  Binary path: $(BINARY_PATH)"
	@echo "  Docker image: $(DOCKER_IMAGE)"
	@echo "  Database URL: $(DB_URL)"
	@echo "  Test DB URL: $(TEST_DB_URL)"
	@if [ -f $(BINARY_PATH) ]; then \
		echo "  Binary exists: Yes ($$(ls -lh $(BINARY_PATH) | awk '{print $$5}'))"; \
	else \
		echo "  Binary exists: No"; \
	fi
	@echo "  Dependencies: $$(go list -m all | wc -l) modules"