#!/bin/bash

# FM24 Golang Development Setup Script
echo "🏈 FM24 Golang Development Environment Setup"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "[${GREEN}✓${NC}] $2"
    else
        echo -e "[${RED}✗${NC}] $2"
    fi
}

echo
echo "📋 Checking Development Dependencies:"
echo "======================================"

# Check Node.js
if command_exists node; then
    NODE_VERSION=$(node --version)
    print_status 0 "Node.js $NODE_VERSION"
else
    print_status 1 "Node.js not found"
    echo -e "${YELLOW}Please install Node.js from https://nodejs.org/${NC}"
fi

# Check npm
if command_exists npm; then
    NPM_VERSION=$(npm --version)
    print_status 0 "npm $NPM_VERSION"
else
    print_status 1 "npm not found"
fi

# Check Go
if command_exists go; then
    GO_VERSION=$(go version | cut -d' ' -f3)
    print_status 0 "Go $GO_VERSION"
else
    print_status 1 "Go not found"
    echo -e "${YELLOW}Please install Go from https://golang.org/dl/${NC}"
fi

# Check golangci-lint
if [ -f "$HOME/go/bin/golangci-lint" ]; then
    GOLANGCI_VERSION=$($HOME/go/bin/golangci-lint version | cut -d' ' -f4)
    print_status 0 "golangci-lint $GOLANGCI_VERSION"
else
    print_status 1 "golangci-lint not found"
    echo -e "${YELLOW}Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest${NC}"
fi

echo
echo "📦 Installing Node.js Dependencies:"
echo "===================================="
if [ -f "package.json" ]; then
    npm install
    if [ $? -eq 0 ]; then
        print_status 0 "Node.js dependencies installed"
    else
        print_status 1 "Failed to install Node.js dependencies"
    fi
else
    print_status 1 "package.json not found"
fi

echo
echo "🧹 Available Development Commands:"
echo "=================================="
echo -e "${BLUE}Frontend (JavaScript/Vue):${NC}"
echo "  npm run lint:check    - Check Biome linting and formatting issues"
echo "  npm run lint          - Fix Biome auto-fixable issues"
echo "  npm run format:check  - Check Biome formatting"
echo "  npm run format        - Fix Biome formatting"
echo "  npm run test          - Run tests in watch mode"
echo "  npm run test:run      - Run tests once"
echo ""
echo -e "${BLUE}Backend (Go):${NC}"
echo "  npm run lint:go       - Check Go linting issues"
echo "  npm run lint:go:fix   - Fix Go linting issues"
echo "  npm run test:go       - Run Go tests (from test/api directory)"
echo "  npm run test:go:coverage - Run Go tests with coverage"
echo ""
echo -e "${BLUE}Combined:${NC}"
echo "  npm run check         - Run all checks (lint + format + test)"
echo "  npm run fix           - Fix all auto-fixable issues"
echo ""
echo -e "${BLUE}Development:${NC}"
echo "  npm run dev           - Start development server"
echo "  npm run build         - Build for production"

echo
echo -e "${GREEN}✅ Setup complete! You can now use the development commands listed above.${NC}" 