# FM24 Golang Development Guide

This document explains how to set up and use the development tools for this project.

## Quick Setup

Run the setup script to check dependencies and install packages:

```bash
./scripts/setup-dev.sh
```

## Tools Overview

### Frontend (JavaScript/Vue.js)
- **ESLint** - JavaScript/Vue linting with Vue 3 best practices
- **Prettier** - Code formatting for consistent style
- **Vitest** - Fast unit testing framework

### Backend (Go)
- **golangci-lint** - Comprehensive Go linting with multiple analyzers
- **Go testing** - Built-in Go test framework

## Development Commands

### Linting

```bash
# Frontend linting
npm run lint:check      # Check for ESLint issues
npm run lint            # Fix auto-fixable ESLint issues

# Go linting  
npm run lint:go         # Check for Go linting issues
npm run lint:go:fix     # Fix auto-fixable Go issues

# Check both
npm run lint:all        # Run all linting checks
```

### Code Formatting

```bash
npm run format:check    # Check if code is properly formatted
npm run format          # Format all code with Prettier
```

### Testing

```bash
# Frontend tests
npm run test            # Run tests in watch mode
npm run test:run        # Run tests once
npm run test:ui         # Run tests with UI
npm run test:coverage   # Run tests with coverage report

# Go tests
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

- **ESLint** (dbaeumer.vscode-eslint)
- **Prettier** (esbenp.prettier-vscode)
- **Vue Language Features (Volar)** (Vue.volar)
- **Go** (golang.go)

### Configuration Files

- `.eslintrc.js` → `eslint.config.js` - ESLint configuration (flat config format)
- `.prettierrc` - Prettier formatting rules
- `.prettierignore` - Files to exclude from formatting
- `.golangci.yml` - Go linting configuration
- `vitest.config.js` - Testing configuration

## Tool Configuration Details

### ESLint Rules
- Vue 3 recommended rules
- Strict about unused variables
- Consistent quote style (single quotes)
- No semicolons
- 2-space indentation

### Prettier Rules
- Single quotes
- No semicolons
- 2-space indentation
- 100 character line width
- No trailing commas

### Go Linting
- Multiple linters enabled (errcheck, gosimple, govet, etc.)
- Security checks (gosec)
- Code quality checks (gocritic, revive)
- Import organization (goimports)

## Pre-commit Hooks (Optional)

You can set up pre-commit hooks to automatically run checks:

```bash
# Install husky for Git hooks
npm install --save-dev husky

# Add pre-commit hook
npx husky add .husky/pre-commit "npm run check"
```

## Troubleshooting

### Common Issues

1. **ESLint errors after updating dependencies**
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
- Use the VS Code extensions for real-time feedback
- Run tests in watch mode during development (`npm run test`)

## Contributing

Before submitting a pull request:

1. Run `npm run check` to ensure all checks pass
2. Run `npm run fix` to auto-fix any issues
3. Make sure all tests pass
4. Follow the existing code style and patterns 