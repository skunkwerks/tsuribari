.PHONY: build test clean fmt vet lint install deps run help dist

# Binary name
BINARY_NAME=tsuribari
BUILD_DIR=bin

# Go parameters
GOCMD?=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Main package path
MAIN_PATH=./cmd/server

# Default target
# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) -v -race ./...

# Run specific test
test-specific:
	@echo "Running specific test (use TEST=TestName)..."
	$(GOTEST) -v -run $(TEST) ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -v -bench=. ./...

# Run tests in short mode
test-short:
	@echo "Running tests in short mode..."
	$(GOTEST) -v -short ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w .
	@command -v goimports >/dev/null 2>&1 && goimports -w . || echo "goimports not found, skipping import formatting"

# Vet code
vet:
	@echo "Vetting code..."
	$(GOVET) ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not found, skipping linting"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Install tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) golang.org/x/tools/cmd/goimports@latest
	@command -v golangci-lint >/dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run the application in development mode
dev:
	@echo "Running in development mode..."
	$(GOCMD) run $(MAIN_PATH)

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Cross-platform builds complete"

# Check code quality
check: fmt vet lint test

# Help target
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  lint          - Lint code (requires golangci-lint)"
	@echo "  deps          - Install dependencies"
	@echo "  install-tools - Install development tools"
	@echo "  run           - Build and run the application"
	@echo "  dev           - Run in development mode"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  check         - Run all code quality checks"
	@echo "  help          - Show this help message"
