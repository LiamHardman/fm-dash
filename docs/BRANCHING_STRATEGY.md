# Branching Strategy & Release Process

## Overview

This repository uses a simplified branching strategy with automated releases using [release-please](https://github.com/google-github-actions/release-please-action).

## Branch Structure

### Main Branches

- **`main`** - Production branch, contains stable, released code
- **`dev`** - Development branch, contains the latest development changes

### Feature Branches

- **`feature/*`** - Feature development branches (e.g., `feature/new-player-cards`)
- **`bugfix/*`** - Bug fix branches (e.g., `bugfix/team-matching-issue`)
- **`hotfix/*`** - Critical production fixes (e.g., `hotfix/security-patch`)

## Workflow

### Development Process

1. **Create feature branch from `dev`**
   ```bash
   git checkout dev
   git pull origin dev
   git checkout -b feature/your-feature-name
   ```

2. **Make changes and commit with conventional commits**
   ```bash
   git commit -m "feat: add new player cards component"
   git commit -m "fix: resolve team matching issue"
   git commit -m "docs: update API documentation"
   ```

3. **Push and create PR to `dev`**
   ```bash
   git push origin feature/your-feature-name
   # Create PR on GitHub from feature/your-feature-name to dev
   ```

4. **Merge to `dev`**
   - PR is reviewed and merged to `dev`
   - CI runs automatically on `dev` branch

### Release Process

1. **Automatic Release Creation**
   - When changes are pushed to `dev`, release-please analyzes commits
   - If conventional commits are found, it creates a release PR to `main`
   - The PR includes:
     - Version bump in `package.json`
     - Updated `CHANGELOG.md`
     - Release notes

2. **Release PR Review**
   - Review the generated release PR
   - Ensure all changes are correct
   - Merge the PR to `main`

3. **Automatic Release**
   - When PR is merged to `main`, release-please:
     - Creates a GitHub release
     - Tags the release
     - Updates the changelog

## Conventional Commits

Use conventional commit messages to trigger automatic releases:

### Commit Types

- **`feat:`** - New features (triggers minor version bump)
- **`fix:`** - Bug fixes (triggers patch version bump)
- **`BREAKING CHANGE:`** - Breaking changes (triggers major version bump)
- **`docs:`** - Documentation changes
- **`style:`** - Code style changes (formatting, etc.)
- **`refactor:`** - Code refactoring
- **`test:`** - Adding or updating tests
- **`chore:`** - Maintenance tasks

### Examples

```bash
# New feature - triggers minor version bump
git commit -m "feat: add player comparison feature"

# Bug fix - triggers patch version bump
git commit -m "fix: resolve team matching algorithm issue"

# Breaking change - triggers major version bump
git commit -m "feat!: change API response format

BREAKING CHANGE: API now returns JSON instead of XML"

# Documentation update - no version bump
git commit -m "docs: update installation instructions"
```

## Version Management

### Version Bumps

- **Patch (1.0.0 → 1.0.1)**: Bug fixes
- **Minor (1.0.0 → 1.1.0)**: New features (backward compatible)
- **Major (1.0.0 → 2.0.0)**: Breaking changes

### Manual Version Override

To force a specific version:

```bash
git commit -m "chore: release 2.0.0"
```

## GitHub Actions

### CI Pipeline (`ci.yml`)
- Runs on `dev` and `main` branches
- Runs on all PRs
- Includes:
  - Frontend linting (Biome)
  - Backend linting (golangci-lint)
  - Tests (frontend and backend)
  - Build verification

### Release Pipeline (`release-please.yml`)
- Runs on pushes to `dev`
- Creates release PRs to `main`
- Handles version bumping and changelog generation

### Deploy Pipeline (`deploy.yml`)
- Runs on pushes to `main`
- Deploys to production
- Only runs after successful CI

## Migration from Old Structure

### Current State
- You have many small releases (1.5.1, 1.5.2, etc.)
- Complex release pipeline with multiple workflows
- Manual release management

### New State
- Clean `main`/`dev` branch structure
- Automated releases with release-please
- Conventional commits for version management
- Simplified CI/CD pipeline

### Migration Steps

1. **Create new branches**
   ```bash
   # Create dev branch from current main
   git checkout -b dev
   git push origin dev
   
   # Rename current main to main-old (backup)
   git branch -m main main-old
   
   # Create new main from current stable state
   git checkout -b main
   git push origin main
   ```

2. **Update default branch**
   - Go to GitHub repository settings
   - Change default branch to `main`

3. **Update branch protection**
   - Protect `main` branch (require PR reviews)
   - Protect `dev` branch (require CI to pass)

4. **Clean up old workflows**
   - Archive or remove old release workflows
   - Keep the new `ci.yml`, `release-please.yml`, and `deploy.yml`

## Benefits

1. **Cleaner History**: Fewer, more meaningful releases
2. **Automated Process**: No manual version management
3. **Better Documentation**: Automatic changelog generation
4. **Consistent Workflow**: Standardized commit messages
5. **Reduced Complexity**: Simplified CI/CD pipeline

## Troubleshooting

### Release Not Created
- Check commit messages follow conventional format
- Verify release-please workflow is enabled
- Check GitHub Actions logs for errors

### Wrong Version Bump
- Review commit messages for proper conventional format
- Check release-please configuration
- Use manual version override if needed

### CI Failures
- Fix linting issues locally first
- Run `npm run check` to verify all checks pass
- Ensure tests are passing before pushing 