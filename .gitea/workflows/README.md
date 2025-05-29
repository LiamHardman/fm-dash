# Gitea Actions Workflows

This directory contains the CI/CD workflows for the FM24 Golang project.

## Workflows

### 🚀 `deploy.yaml` - Deployment Pipeline

**Triggers:** Push to `main` branch

**Purpose:** Builds and deploys the application to Kubernetes

**Steps:**
1. **Quick Code Quality Check** (informational only)
   - Runs Biome linting on frontend code
   - Does NOT block deployment if issues are found
   - Provides visibility into code quality

2. **Build & Deploy**
   - Builds Docker image with unified Nginx + Go backend
   - Pushes to container registry
   - Deploys to Kubernetes cluster

**Key Features:**
- 🔒 Code quality checks are informational only
- 🚀 Deployment always proceeds regardless of linting results
- 🔄 Handles both new deployments and rolling updates

### 🔍 `code-quality.yaml` - Code Quality Checks

**Triggers:** 
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Purpose:** Comprehensive code quality analysis (informational only)

**Jobs:**

#### Frontend Code Quality (Biome)
- Runs Biome linting checks
- Checks code formatting
- Provides suggestions for fixes

#### Backend Code Quality (Go)
- Runs golangci-lint with comprehensive rules
- Executes Go test suite
- Validates Go code quality standards

#### Code Quality Summary
- Aggregates results from both frontend and backend
- Provides helpful commands for local development
- Shows overall status

**Key Features:**
- 🛡️ `continue-on-error: true` ensures workflow always succeeds
- 📊 Provides detailed feedback without blocking development
- 💡 Includes helpful commands for local fixes

## Local Development

Use these commands locally to match the CI checks:

```bash
# Quick checks (matches deployment pipeline)
npm run lint:check

# Comprehensive checks (matches code quality pipeline)  
npm run check           # All checks (frontend + backend + tests)
npm run fix             # Fix auto-fixable issues

# Individual checks
npm run lint:check      # Biome linting
npm run format:check    # Biome formatting  
npm run lint:go         # Go linting
npm run test:all        # All tests
```

## Pipeline Philosophy

- **Deployment First**: Never block deployments with code quality issues
- **Visibility**: Provide clear feedback on code quality status
- **Developer Friendly**: Include actionable commands for fixing issues
- **Separation of Concerns**: Quality checks run independently from deployment

## Troubleshooting

### Code Quality Issues
If you see code quality warnings:

1. **Local Fixes:**
   ```bash
   npm run fix           # Auto-fix most issues
   npm run lint:go:fix   # Fix Go issues
   ```

2. **Check Specific Issues:**
   ```bash
   npm run lint:check    # See detailed Biome issues
   npm run lint:go       # See detailed Go issues
   ```

3. **Deployment Impact:**
   - Code quality issues do NOT block deployments
   - Focus on fixing issues in the next development cycle

### Pipeline Failures
- If deployment fails, check the Docker build and Kubernetes steps
- Code quality failures are informational only and won't cause pipeline failure
- Check the `code-quality.yaml` workflow for detailed quality feedback 