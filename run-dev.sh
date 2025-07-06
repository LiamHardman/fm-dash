#!/bin/bash

# Development runner script for v2fmdash
# This script provides easy ways to run the application with different configurations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Show usage
show_usage() {
    echo "v2fmdash Development Runner"
    echo ""
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  dev              Run with development config (OTEL disabled)"
    echo "  prod             Run with production config"
    echo "  otel             Run with development config + OTEL enabled"
    echo "  custom <file>    Run with custom config file"
    echo "  docker           Build and run in Docker"
    echo "  help             Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 dev                                    # Development mode"
    echo "  $0 otel                                   # Development with OTEL"
    echo "  $0 prod                                   # Production config"
    echo "  $0 custom ./my-config.yaml               # Custom config"
    echo "  $0 docker                                 # Docker container"
    echo ""
}

# Check if we're in the right directory
check_environment() {
    if [[ ! -f "config/backend-config.yaml" ]]; then
        print_error "backend-config.yaml not found. Please run from project root."
        exit 1
    fi
    
    if [[ ! -d "src/api" ]]; then
        print_error "src/api directory not found. Please run from project root."
        exit 1
    fi
}

# Run with development config
run_dev() {
    print_info "Starting v2fmdash in development mode..."
    print_info "Using development-config.yaml (OTEL disabled for faster startup)"
    
    cd src/api
    CONFIG_FILE_PATH=../../config/development-config.yaml go run .
}

# Run with production config
run_prod() {
    print_info "Starting v2fmdash in production mode..."
    print_info "Using backend-config.yaml (OTEL enabled)"
    
    cd src/api
    CONFIG_FILE_PATH=../../config/backend-config.yaml go run .
}

# Run with OTEL enabled
run_otel() {
    print_info "Starting v2fmdash in development mode with OTEL enabled..."
    print_warning "Make sure you have an OTEL collector running at localhost:4317"
    
    cd src/api
    CONFIG_FILE_PATH=../../config/development-config.yaml \
    OTEL_ENABLED=true \
    OTEL_TRACES_ENABLED=true \
    OTEL_METRICS_ENABLED=true \
    go run .
}

# Run with custom config
run_custom() {
    local config_file="$1"
    
    if [[ -z "$config_file" ]]; then
        print_error "Please specify a config file path"
        echo "Usage: $0 custom <config-file>"
        exit 1
    fi
    
    if [[ ! -f "$config_file" ]]; then
        print_error "Config file not found: $config_file"
        exit 1
    fi
    
    print_info "Starting v2fmdash with custom config: $config_file"
    
    cd src/api
    CONFIG_FILE_PATH="../../$config_file" go run .
}

# Run in Docker
run_docker() {
    print_info "Building and running v2fmdash in Docker..."
    
    # Build the image
    print_info "Building Docker image..."
    docker build -f Dockerfile.backend -t v2fmdash-backend .
    
    # Run the container
    print_info "Starting container on port 8091..."
    docker run -p 8091:8091 --rm v2fmdash-backend
}

# Main script logic
main() {
    local command="${1:-help}"
    
    case "$command" in
        "dev")
            check_environment
            run_dev
            ;;
        "prod")
            check_environment
            run_prod
            ;;
        "otel")
            check_environment
            run_otel
            ;;
        "custom")
            check_environment
            run_custom "$2"
            ;;
        "docker")
            check_environment
            run_docker
            ;;
        "help"|"--help"|"-h")
            show_usage
            ;;
        *)
            print_error "Unknown command: $command"
            echo ""
            show_usage
            exit 1
            ;;
    esac
}

# Run the main function with all arguments
main "$@" 