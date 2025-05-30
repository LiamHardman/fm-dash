# FM24 Golang Development Guide

This document explains how to set up and use the development tools for this project.

## Quick Setup

Run the setup script to check dependencies and install packages:

```bash
./scripts/setup-dev.sh
```

## Tools Overview

### Frontend (JavaScript/Vue.js)
- **Biome** - Fast, unified linting and formatting tool (replaces ESLint + Prettier)
- **Vitest** - Fast unit testing framework

### Backend (Go)
- **golangci-lint** - Comprehensive Go linting with multiple analyzers
- **Go testing** - Built-in Go test framework

## Development Commands

### Linting & Formatting

```bash
# Frontend linting and formatting (Biome)
npm run lint:check      # Check for linting and formatting issues
npm run lint            # Fix auto-fixable linting issues
npm run format:check    # Check formatting only
npm run format          # Fix formatting issues

# Go linting  
npm run lint:go         # Check for Go linting issues
npm run lint:go:fix     # Fix auto-fixable Go issues

# Check both
npm run lint:all        # Run all linting checks
```

### Testing

```bash
# Frontend tests
npm run test            # Run tests in watch mode
npm run test:run        # Run tests once
npm run test:ui         # Run tests with UI
npm run test:coverage   # Run tests with coverage report

# Go tests (now run from test/api directory)
npm run test:go         # Run Go tests
npm run test:go:verbose # Run Go tests with verbose output
npm run test:go:coverage # Run Go tests with coverage

# Run all tests
npm run test:all        # Run both frontend and backend tests
```

### Combined Commands

```bash
npm run check           # Run all checks (lint + format + test)
npm run fix             # Fix all auto-fixable issues
```

### Development Server

```bash
npm run dev             # Start Vite development server
npm run build           # Build for production
npm run preview         # Preview production build
```

## IDE Integration

### VS Code

Install these extensions for the best development experience:

- **Biome** (biomejs.biome) - Unified linting and formatting
- **Vue Language Features (Volar)** (Vue.volar)
- **Go** (golang.go)

### Configuration Files

- `biome.json` - Biome configuration for linting and formatting
- `.golangci.yml` - Go linting configuration
- `vitest.config.js` - Testing configuration

## Tool Configuration Details

### Biome Rules
- Vue 3 support with modern JavaScript
- Strict about unused variables and console statements
- Consistent quote style (single quotes)
- No semicolons (as needed)
- 2-space indentation
- 100 character line width
- Import organization enabled

### Go Linting
- Multiple linters enabled (errcheck, gosimple, govet, etc.)
- Security checks (gosec)
- Code quality checks (gocritic, revive)
- Import organization (goimports)

## Pre-commit Hooks

This project uses **Husky** and **lint-staged** for automated code quality checks:

### Hooks Configured

- **pre-commit**: Runs lint-staged for fast, staged-file-only checks
  - Formats and lints JavaScript/Vue files with Biome
  - Lints and fixes Go files with golangci-lint
  - Only processes staged files for optimal performance

- **pre-push**: Runs comprehensive checks before pushing
  - Full linting (frontend and backend)
  - Format checking
  - All tests (frontend and backend)

### Manual Hook Management

```bash
# Hooks are automatically installed via the "prepare" script
# But you can manually manage them:

# Reinstall hooks
npx husky install

# Skip hooks temporarily (not recommended)
git commit --no-verify
git push --no-verify

# Test hooks manually
npx lint-staged              # Test pre-commit
npm run check               # Test pre-push
```

### Performance Benefits

- **lint-staged**: Only checks files you've changed
- **Biome**: Faster than ESLint + Prettier
- **Selective testing**: Only runs relevant tests

## Troubleshooting

### Common Issues

1. **Biome errors after updating dependencies**
   ```bash
   npm run lint:check
   npm run lint  # Fix auto-fixable issues
   ```

2. **Go linting fails**
   ```bash
   # Make sure golangci-lint is installed
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

3. **Tests fail to run**
   ```bash
   # Clear cache and reinstall
   rm -rf node_modules package-lock.json
   npm install
   ```

### Performance Tips

- Use `npm run lint` (with auto-fix) instead of manually fixing issues
- Run `npm run check` before committing
- Use the VS Code Biome extension for real-time feedback
- Run tests in watch mode during development (`npm run test`)

## Contributing

Before submitting a pull request:

1. Run `npm run check` to ensure all checks pass
2. Run `npm run fix` to auto-fix any issues
3. Make sure all tests pass
4. Follow the existing code style and patterns

## CI/CD Integration

### Gitea Actions Pipelines

This project includes automated code quality checks in the CI/CD pipeline:

#### 🚀 Deployment Pipeline (`deploy.yaml`)
- **Triggers:** Push to `main` branch
- **Quick code quality check** (informational only)
- **Never blocks deployments** - provides visibility without disrupting releases
- Matches: `npm run lint:check`

#### 🔍 Code Quality Pipeline (`code-quality.yaml`)  
- **Triggers:** Push to `main`/`develop` or pull requests
- **Comprehensive checks:** Both frontend (Biome) and backend (Go) linting
- **Always succeeds** - provides detailed feedback without failing the pipeline
- Matches: `npm run check`

#### Key Features:
- 🔒 **Non-blocking**: Code quality issues never prevent deployments
- 📊 **Informational**: Clear feedback on code quality status
- 💡 **Actionable**: Includes commands to fix issues locally
- 🔄 **Comprehensive**: Covers both frontend and backend code

#### Local vs CI Commands:
```bash
# Local development (matches CI exactly)
npm run lint:check      # Quick check (deployment pipeline)
npm run check           # Full check (code quality pipeline)  
npm run fix             # Fix issues before committing
```

## Why Biome?

Biome offers several advantages over ESLint + Prettier:

- **Faster**: Single tool, written in Rust
- **Unified**: Linting and formatting in one configuration
- **Better Vue support**: Native Vue.js support
- **Simpler setup**: Less configuration complexity
- **Import organization**: Built-in import sorting 