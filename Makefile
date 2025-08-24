# StaticLang Compiler Makefile

.PHONY: all build test clean install fmt vet lint run-example help

# Configuration
BINARY_NAME=staticlang
MAIN_PATH=./cmd/staticlang
BUILD_DIR=./build
COVERAGE_DIR=./coverage
RUNTIME_DIR=./runtime

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# C compiler parameters
CC=clang
CFLAGS=-O2 -Wall -Wextra

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev') -X main.BuildDate=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

# Default target
all: fmt vet test build

# Build the compiler
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build the runtime library
build-runtime:
	@echo "Building runtime library..."
	@mkdir -p $(BUILD_DIR)
	$(CC) $(CFLAGS) -c -o $(BUILD_DIR)/builtin.o $(RUNTIME_DIR)/builtin.c

# Build both compiler and runtime
build-with-runtime: build build-runtime

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated at $(COVERAGE_DIR)/coverage.html"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Vet code
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Run golangci-lint (if installed)
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Generate parser (if using Goyacc)
generate-parser:
	@echo "Generating parser with Goyacc..."
	@if command -v goyacc >/dev/null 2>&1; then \
		cd grammar && goyacc -v y.output -o parser.go -l staticlang.y; \
	else \
		echo "goyacc not found. Install it with:"; \
		echo "  go install golang.org/x/tools/cmd/goyacc@latest"; \
	fi

# Install the compiler
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)
	rm -f examples/hello examples/hello.ll

# Run example compilation
run-example: build-with-runtime
	@echo "Running example compilation..."
	@mkdir -p examples
	@echo 'func main() -> int { print("Hello from StaticLang!"); print("Answer:", 42); return 42; }' > examples/hello.sl
	$(BUILD_DIR)/$(BINARY_NAME) -i examples/hello.sl -o examples/hello.ll -v
	@echo "Compiling LLVM IR to executable..."
	$(CC) examples/hello.ll $(BUILD_DIR)/builtin.o -o examples/hello
	@echo "Running compiled program..."
	./examples/hello

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	$(GOMOD) download
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goyacc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Show available targets
help:
	@echo "Available targets:"
	@echo "  all           - Format, vet, test, and build"
	@echo "  build         - Build the compiler for current platform"
	@echo "  build-runtime - Build the runtime library (builtin.c)"
	@echo "  build-with-runtime - Build both compiler and runtime"
	@echo "  build-all     - Build for all supported platforms"
	@echo "  build-linux   - Build for Linux"
	@echo "  build-darwin  - Build for macOS"
	@echo "  build-windows - Build for Windows"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  bench         - Run benchmarks"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run golangci-lint"
	@echo "  deps          - Install dependencies"
	@echo "  generate-parser - Generate parser from grammar"
	@echo "  install       - Install compiler to GOPATH/bin"
	@echo "  clean         - Clean build artifacts"
	@echo "  run-example   - Build and run example compilation"
	@echo "  dev-setup     - Set up development environment"
	@echo "  help          - Show this help message"

# Docker support
docker-build:
	@echo "Building Docker image..."
	docker build -t staticlang:latest .

docker-run:
	@echo "Running in Docker..."
	docker run --rm -v $(PWD):/workspace staticlang:latest

# Release targets
tag:
	@echo "Current version: $(shell git describe --tags --always)"
	@echo "Create a new tag with: git tag -a v1.0.0 -m 'Release v1.0.0'"

release: clean fmt vet test build-all
	@echo "Release build complete!"
	@echo "Binaries available in $(BUILD_DIR)/"

# Debug build
debug:
	@echo "Building debug version..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME)-debug $(MAIN_PATH)
