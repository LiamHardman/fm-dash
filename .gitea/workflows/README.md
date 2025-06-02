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

### 🏷️ `release.yaml` - Conventional Commit Release

**Triggers:** Push to `main` branch

**Purpose:** Automated semantic versioning and release creation based on conventional commits

**Steps:**
1. **Commit Analysis**
   - Analyzes commit messages using conventional commit format
   - Determines appropriate version bump (major/minor/patch)
   - Skips release if no relevant changes found

2. **Build & Test**
   - Builds the application
   - Runs test suites to ensure quality
   - Validates release readiness

3. **Release Creation**
   - Updates package.json version
   - Generates/updates CHANGELOG.md
   - Creates Git tag
   - Creates GitHub/Gitea release with assets

**Conventional Commit Format:**
```
type(scope): description

[optional body]

[optional footer]
```

**Release Rules:**
- `feat:` → Minor version bump (1.0.0 → 1.1.0)
- `fix:` → Patch version bump (1.0.0 → 1.0.1)
- `feat!:` or `BREAKING CHANGE:` → Major version bump (1.0.0 → 2.0.0)
- `docs:`, `style:`, `chore:`, `test:`, `ci:` → No version bump
- `refactor:`, `perf:` → Patch version bump

**Skip Release:**
- Add `[skip release]` or `[skip ci]` to commit message

**Key Features:**
- 🤖 Fully automated semantic versioning
- 📝 Auto-generated changelog with emoji categories
- 🏷️ Git tags and GitHub releases
- 📦 Includes built application assets
- 🚫 Smart skip conditions

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

## Conventional Commits Guide

To trigger releases, use conventional commit messages:

```bash
# Feature (minor version bump)
git commit -m "feat: add user authentication system"
git commit -m "feat(api): implement player statistics endpoint"

# Bug fix (patch version bump)
git commit -m "fix: resolve memory leak in data processing"
git commit -m "fix(ui): correct responsive layout on mobile"

# Breaking change (major version bump)
git commit -m "feat!: redesign API with new authentication"
git commit -m "feat: remove deprecated endpoints

BREAKING CHANGE: The /old-api endpoint has been removed"

# No release
git commit -m "docs: update installation instructions"
git commit -m "chore: update dependencies"
git commit -m "style: fix code formatting"

# Skip release entirely
git commit -m "feat: add new feature [skip release]"
```

## Pipeline Philosophy

- **Deployment First**: Never block deployments with code quality issues
- **Visibility**: Provide clear feedback on code quality status
- **Developer Friendly**: Include actionable commands for fixing issues
- **Separation of Concerns**: Quality checks run independently from deployment
- **Semantic Versioning**: Automated releases based on commit semantics

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

### Release Issues
If releases aren't being created:

1. **Check Commit Format:**
   ```bash
   # Ensure commits follow conventional format
   git log --oneline -5
   ```

2. **Verify Triggers:**
   - Releases only trigger on pushes to `main` branch
   - Check for `[skip release]` in commit messages

3. **Debug Release Process:**
   - Check workflow logs for semantic-release dry-run output
   - Verify GitHub token permissions for creating releases

### Pipeline Failures
- If deployment fails, check the Docker build and Kubernetes steps
- Code quality failures are informational only and won't cause pipeline failure
- Check the `code-quality.yaml` workflow for detailed quality feedback
- Release failures won't affect deployment pipeline (they run independently) 