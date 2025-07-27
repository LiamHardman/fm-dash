# Release Structure Migration Summary

## What's Been Set Up

I've restructured your GitHub repository to use a cleaner, more maintainable release strategy with automated releases using [release-please](https://github.com/google-github-actions/release-please-action).

## New Structure Overview

### Branch Strategy
- **`main`** - Production branch (stable, released code)
- **`dev`** - Development branch (latest development changes)
- **Feature branches** - `feature/*`, `bugfix/*`, `hotfix/*`

### Automated Release Process
1. **Development** → Work on `dev` branch with conventional commits
2. **Release Creation** → release-please creates PRs to `main` with version bumps
3. **Release** → Merging to `main` creates GitHub releases automatically

## Files Created/Modified

### New Files
- `.github/release-please.yml` - release-please configuration
- `.github/workflows/release-please.yml` - Automated release workflow
- `.github/workflows/dev-build.yml` - Dev branch build and release workflow
- `.github/workflows/create-release-pr.yml` - Release PR creation workflow
- `.github/workflows/ci.yml` - Simplified CI pipeline
- `docs/BRANCHING_STRATEGY.md` - Complete documentation
- `docs/DEV_BRANCH_WORKFLOW.md` - Dev branch workflow documentation
- `scripts/create-dev-branch.sh` - Dev branch creation script

### Modified Files
- `.github/workflows/deploy.yml` - Updated for new branch structure
- `package.json` - Removed semantic-release dependencies

## Key Benefits

### Before (Current State)
- Many small releases (1.5.1, 1.5.2, 1.5.3, etc.)
- Complex release pipeline with multiple workflows
- Manual version management
- Frequent releases cluttering the release history

### After (New State)
- **Cleaner History**: Fewer, more meaningful releases
- **Automated Process**: No manual version management
- **Better Documentation**: Automatic changelog generation
- **Consistent Workflow**: Standardized commit messages
- **Reduced Complexity**: Simplified CI/CD pipeline
- **Dev Builds**: Docker images tagged with `dev` and commit SHA
- **No Dev Releases**: Dev builds don't clutter GitHub releases

## Migration Steps

### 1. Create Dev Branch
```bash
./scripts/create-dev-branch.sh
```

This script will:
- Create a temporary branch
- Squash all current commits into a single commit
- Create a new `dev` branch from the squashed commit
- Push the dev branch to remote

### 2. Update GitHub Settings
- Go to repository Settings → Branches
- Change default branch to `main`
- Set up branch protection rules:
  - `main`: Require PR reviews, require CI to pass
  - `dev`: Require CI to pass

### 3. Clean Up Old Workflows
Archive or remove these old workflows:
- `release.yml` (replaced by release-please.yml)
- `gh-release-pipeline.yml` (replaced by ci.yml + release-please.yml)

Keep these new workflows:
- `ci.yml` - Code quality and testing
- `release-please.yml` - Automated releases
- `deploy.yml` - Production deployment

### 4. Update Team Workflow
- Switch to working on `dev` branch for development
- Use conventional commits for automatic releases:
  ```bash
  git commit -m "feat: add new player cards"
  git commit -m "fix: resolve team matching issue"
  git commit -m "docs: update API documentation"
  ```

## Conventional Commits

Use these commit types to trigger automatic releases:

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat:` | New features | Minor (1.0.0 → 1.1.0) |
| `fix:` | Bug fixes | Patch (1.0.0 → 1.0.1) |
| `BREAKING CHANGE:` | Breaking changes | Major (1.0.0 → 2.0.0) |
| `docs:` | Documentation | None |
| `style:` | Code style | None |
| `refactor:` | Code refactoring | None |
| `test:` | Tests | None |
| `chore:` | Maintenance | None |

## Example Workflow

### Development
```bash
# Create feature branch
git checkout dev
git pull origin dev
git checkout -b feature/new-player-cards

# Make changes and commit
git add .
git commit -m "feat: add new player cards component"
git commit -m "fix: resolve alignment issues in player cards"

# Push and create PR
git push origin feature/new-player-cards
# Create PR on GitHub: feature/new-player-cards → dev
```

### Release Process
1. **Merge to dev** → release-please analyzes commits
2. **Release PR created** → PR to main with version bump and changelog
3. **Review and merge** → Merge PR to main
4. **GitHub release created** → Automatic release with tag

## Testing the New Workflow

### Test Release Process
1. Make a small change on `dev` branch
2. Use conventional commit: `git commit -m "feat: test new release workflow"`
3. Push to `dev`: `git push origin dev`
4. Check GitHub Actions → release-please should create a PR to `main`
5. Review and merge the PR
6. Verify GitHub release is created automatically

## Troubleshooting

### Release Not Created
- Check commit messages follow conventional format
- Verify release-please workflow is enabled
- Check GitHub Actions logs for errors

### Wrong Version Bump
- Review commit messages for proper conventional format
- Check release-please configuration
- Use manual version override if needed: `git commit -m "chore: release 2.0.0"`

### CI Failures
- Fix linting issues locally first
- Run `npm run check` to verify all checks pass
- Ensure tests are passing before pushing

## Documentation

- **Complete Guide**: `docs/BRANCHING_STRATEGY.md`
- **Migration Script**: `scripts/migrate-to-new-branching.sh`
- **Release Config**: `.github/release-please.yml`

## Next Steps

1. **Run the migration script** to set up new branches
2. **Update GitHub settings** for branch protection
3. **Test the workflow** with a small change
4. **Clean up old workflows** once new system is working
5. **Update team documentation** and processes

This new structure will give you a much cleaner release history and automated version management, reducing the overhead of manual releases while maintaining proper version control and documentation. 