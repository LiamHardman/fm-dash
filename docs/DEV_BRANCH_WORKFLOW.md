# Dev Branch Workflow

## Overview

The `dev` branch serves as the primary development branch where all development work happens. When you push to `dev`, it automatically:

1. **Runs code quality checks** (linting, formatting, tests)
2. **Builds the application** (frontend + backend)
3. **Creates Docker images** tagged with `dev-{SHA}`
4. **Creates GitHub releases** for development builds
5. **Creates release PRs** to main when conventional commits are detected

## Dev Branch Setup

### Creating the Dev Branch

Run the setup script to create a clean dev branch with squashed commits:

```bash
./scripts/create-dev-branch.sh
```

This script will:
- Create a temporary branch
- Squash all current commits into a single commit
- Create a new `dev` branch from the squashed commit
- Push the dev branch to remote

### What Gets Created

After running the script, you'll have:
- ✅ **Dev branch** with single squashed commit
- ✅ **Automatic build pipeline** on dev pushes
- ✅ **Container/package releases** tagged with `dev-{SHA}`
- ✅ **Ready for development workflow**

## Development Workflow

### Daily Development

1. **Switch to dev branch**
   ```bash
   git checkout dev
   git pull origin dev
   ```

2. **Create feature branch** (optional)
   ```bash
   git checkout -b feature/your-feature-name
   # Make changes...
   git add .
   git commit -m "feat: your new feature"
   git push origin feature/your-feature-name
   # Create PR to dev
   ```

3. **Or work directly on dev**
   ```bash
   git checkout dev
   # Make changes...
   git add .
   git commit -m "feat: your new feature"
   git push origin dev
   ```

### What Happens on Dev Push

When you push to `dev`, the following happens automatically:

#### 1. Code Quality Checks
- **Frontend linting** (Biome)
- **Backend linting** (golangci-lint)
- **Formatting checks**
- **Build verification**

#### 2. Build Process
- **Frontend build** (Vite)
- **Backend build** (Go)
- **Artifact creation** (archives, binaries)

#### 3. Docker Images
- **Unified image**: `your-username/fm-dash:dev` and `your-username/fm-dash:{SHA}`
- **Frontend image**: `your-username/fm-dash-frontend:dev` and `your-username/fm-dash-frontend:{SHA}`
- **Backend image**: `your-username/fm-dash-backend:dev` and `your-username/fm-dash-backend:{SHA}`

#### 4. No GitHub Releases
- **No releases created**: Dev builds don't create GitHub releases
- **Build artifacts**: Available in GitHub Actions artifacts
- **Docker images**: Available in Docker Hub with dev and SHA tags

#### 5. Release PR Creation
- **If conventional commits detected**: Creates PR to `main`
- **Version bump**: Based on commit types (feat, fix, etc.)
- **Changelog**: Automatic generation from commits

## Tagging Strategy

### Dev Builds
- **Docker tags**: `dev` and `{short-sha}` (e.g., `dev` and `a1b2c3d`)
- **Unified**: `your-username/fm-dash:dev` and `your-username/fm-dash:a1b2c3d`
- **Frontend**: `your-username/fm-dash-frontend:dev` and `your-username/fm-dash-frontend:a1b2c3d`
- **Backend**: `your-username/fm-dash-backend:dev` and `your-username/fm-dash-backend:a1b2c3d`
- **No GitHub releases**: Dev builds don't create releases

### Production Releases
- **Format**: `v{major}.{minor}.{patch}` (e.g., `v1.2.3`)
- **Docker**: `your-username/fm-dash:v1.2.3`
- **GitHub Release**: `v1.2.3`

## Conventional Commits

Use conventional commits to trigger automatic release PRs:

### Commit Types
- **`feat:`** - New features (triggers minor version bump)
- **`fix:`** - Bug fixes (triggers patch version bump)
- **`BREAKING CHANGE:`** - Breaking changes (triggers major version bump)
- **`docs:`** - Documentation changes
- **`style:`** - Code style changes
- **`refactor:`** - Code refactoring
- **`test:`** - Adding or updating tests
- **`chore:`** - Maintenance tasks

### Examples
```bash
# New feature - creates release PR with minor version bump
git commit -m "feat: add player comparison feature"

# Bug fix - creates release PR with patch version bump
git commit -m "fix: resolve team matching algorithm issue"

# Breaking change - creates release PR with major version bump
git commit -m "feat!: change API response format

BREAKING CHANGE: API now returns JSON instead of XML"

# Documentation - no release PR created
git commit -m "docs: update installation instructions"
```

## Release Process

### Development Cycle
1. **Work on dev** → Make changes and push to dev
2. **Automatic builds** → Docker images and artifacts created
3. **Release PR** → If conventional commits, PR created to main
4. **Review and merge** → Merge PR to main when ready
5. **Production release** → GitHub release created automatically

### Example Timeline
```
Day 1: Push feat: new player cards → dev build + release PR created
Day 2: Push fix: alignment issue → dev build + updated release PR
Day 3: Push docs: update API → dev build (no release PR)
Day 4: Merge release PR to main → Production release v1.2.0 created
```

## GitHub Actions Workflows

### Dev Build & Release (`dev-build.yml`)
- **Triggers**: Push to `dev`
- **Actions**:
  - Code quality checks
  - Build application (frontend, backend, unified)
  - Create Docker images with `dev` and `{SHA}` tags
  - Upload build artifacts (no GitHub releases)

### Create Release PR (`create-release-pr.yml`)
- **Triggers**: Push to `dev`
- **Actions**:
  - Analyze conventional commits
  - Create PR to main with version bump
  - Update changelog

### Release Please (`release-please.yml`)
- **Triggers**: Push to `main`
- **Actions**:
  - Create GitHub release
  - Tag the release
  - Update changelog

## Benefits

### For Development
- **Immediate feedback**: Every push creates a build
- **Easy testing**: Docker images available for testing
- **No manual releases**: Automatic dev builds
- **Clear separation**: Dev vs production releases

### For Production
- **Clean history**: Only meaningful releases in main
- **Automated process**: No manual version management
- **Proper tagging**: Clear distinction between dev and production
- **Rollback capability**: Easy to revert to previous versions

## Troubleshooting

### Dev Build Not Created
- Check GitHub Actions for workflow failures
- Verify you're pushing to `dev` branch
- Check for linting/formatting issues

### Release PR Not Created
- Verify commit messages follow conventional format
- Check release-please workflow logs
- Ensure conventional commits are present

### Docker Images Not Pushed
- Verify Docker Hub credentials in secrets
- Check Docker build logs
- Ensure Docker Hub repository exists

## Next Steps

1. **Run the setup script**: `./scripts/create-dev-branch.sh`
2. **Test the workflow**: Make a small change and push to dev
3. **Verify builds**: Check GitHub Actions and Docker Hub
4. **Test release PR**: Use conventional commits to trigger PRs
5. **Update team**: Share the new workflow with your team

This workflow gives you the best of both worlds: rapid development with immediate feedback, and clean, automated production releases! 