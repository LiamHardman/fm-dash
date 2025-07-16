# Makefile for protobuf code generation and build tasks

# Variables
PROTO_DIR = proto
PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)
GENERATED_FILES = $(PROTO_FILES:.proto=.pb.go)

# Default target
.PHONY: all
all: generate-proto build

# Generate protobuf Go code
.PHONY: generate-proto
generate-proto: $(GENERATED_FILES)

# Rule to generate .pb.go files from .proto files
%.pb.go: %.proto
	@echo "Generating Go code from $<..."
	protoc --go_out=. --go_opt=paths=source_relative $<

# Clean generated files
.PHONY: clean-proto
clean-proto:
	@echo "Cleaning generated protobuf files..."
	rm -f $(PROTO_DIR)/*.pb.go

# Build the application
.PHONY: build
build:
	@echo "Building application..."
	go build -o bin/api .

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Verify protobuf setup
.PHONY: verify-proto
verify-proto:
	@echo "Verifying protobuf setup..."
	@which protoc > /dev/null || (echo "Error: protoc not found. Please install Protocol Buffers compiler." && exit 1)
	@which protoc-gen-go > /dev/null || (echo "Error: protoc-gen-go not found. Please install with: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" && exit 1)
	@echo "Protobuf setup verified successfully!"

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all              - Generate protobuf code and build application"
	@echo "  generate-proto   - Generate Go code from protobuf files"
	@echo "  clean-proto      - Remove generated protobuf files"
	@echo "  build            - Build the application"
	@echo "  test             - Run tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  deps             - Install and tidy dependencies"
	@echo "  verify-proto     - Verify protobuf toolchain is installed"
	@echo "  help             - Show this help message"

# Ensure bin directory exists
bin:
	mkdir -p bin