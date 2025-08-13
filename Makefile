.PHONY: build run test clean help deps lint fmt vet

# Variables
BINARY_NAME=cobra-template
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=.

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	go build -ldflags="-w -s" -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

# Build for development (with debug info)
build-dev:
	@echo "Building $(BINARY_NAME) for development..."
	@mkdir -p bin
	go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Development build complete: $(BINARY_PATH)"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_PATH) server

# Run in development mode
run-dev:
	@echo "Running $(BINARY_NAME) in development mode..."
	go run $(MAIN_PATH) server --verbose

# Run with custom port
run-port:
	@echo "Running $(BINARY_NAME) on port 3000..."
	go run $(MAIN_PATH) server --port 3000

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean
	@echo "Clean complete"

# Lint the code
lint:
	@echo "Running linter..."
	golangci-lint run

# Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet the code
vet:
	@echo "Running go vet..."
	go vet ./...

# Install linting tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060
	@echo "Documentation server running on http://localhost:6060"

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .
	@echo "Docker image built: $(BINARY_NAME):latest"

# Docker run
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME):latest

# Check for vulnerabilities
security:
	@echo "Checking for vulnerabilities..."
	govulncheck ./...

# Display help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application (optimized)"
	@echo "  build-dev     - Build the application (with debug info)"
	@echo "  run           - Build and run the application"
	@echo "  run-dev       - Run the application in development mode"
	@echo "  run-port      - Run the application on port 3000"
	@echo "  deps          - Install dependencies"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  install-tools - Install development tools"
	@echo "  docs          - Generate and serve documentation"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Build and run Docker container"
	@echo "  security      - Check for security vulnerabilities"
	@echo "  help          - Display this help message"
