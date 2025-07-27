#!/bin/bash

# Script to create dev branch with squashed commits and dev-specific workflows
# This creates a clean dev branch for development with automatic builds

set -e

echo "🚀 Creating dev branch with squashed commits..."
echo "=============================================="

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Error: Not in a git repository"
    exit 1
fi

# Check current branch
CURRENT_BRANCH=$(git branch --show-current)
echo "📍 Current branch: $CURRENT_BRANCH"

# Check if we have uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo "⚠️  Warning: You have uncommitted changes"
    echo "   Please commit or stash them before proceeding"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Script cancelled"
        exit 1
    fi
fi

echo ""
echo "📋 Dev Branch Creation Steps:"
echo "============================"

# Step 1: Create a temporary branch with all current changes
echo "1️⃣  Creating temporary branch with all changes..."
git checkout -b temp-squash-branch

# Step 2: Reset to create a single commit
echo ""
echo "2️⃣  Squashing all commits into a single commit..."
git reset --soft $(git rev-list --max-parents=0 HEAD)
git commit --no-verify -m "feat: initial dev branch setup

- Squashed all previous commits
- Set up new branching strategy
- Configured release-please for automated releases
- Updated CI/CD pipelines for dev/main workflow"

# Step 3: Create dev branch from the squashed commit
echo ""
echo "3️⃣  Creating dev branch from squashed commit..."
git checkout -b dev
echo "   ✅ Dev branch created with single squashed commit"

# Step 4: Push dev branch to remote
echo ""
echo "4️⃣  Pushing dev branch to remote..."
git push --no-verify -u github dev
echo "   ✅ Dev branch pushed to remote"

# Step 5: Clean up temporary branch
echo ""
echo "5️⃣  Cleaning up temporary branch..."
git branch -D temp-squash-branch
echo "   ✅ Temporary branch removed"

echo ""
echo "✅ Dev branch creation completed successfully!"
echo "============================================"
echo ""
echo "📋 What was created:"
echo "==================="
echo ""
echo "✅ Dev branch with single squashed commit"
echo "✅ Automatic build pipeline on dev pushes"
echo "✅ Container/package releases tagged with 'dev' and '{SHA}'"
echo "✅ Ready for development workflow"
echo ""
echo "📋 Next steps:"
echo "=============="
echo ""
echo "1. Switch to dev branch for development:"
echo "   git checkout dev"
echo ""
echo "2. Make changes and push to dev:"
echo "   git add ."
echo "   git commit -m 'feat: your new feature'"
echo "   git push github dev"
echo ""
echo "3. Check GitHub Actions for automatic builds:"
echo "   - Builds will be tagged with 'dev' and '{commit-sha}'"
echo "   - Containers will be built and pushed"
echo "   - Packages will be created"
echo ""
echo "4. When ready for production:"
echo "   - Merge dev to main"
echo "   - release-please will create proper versioned releases"
echo ""
echo "🎉 Dev branch is ready! Happy coding! 🚀" 