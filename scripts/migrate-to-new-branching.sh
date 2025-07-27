#!/bin/bash

# Migration script for transitioning to new main/dev branching strategy
# This script helps set up the new branch structure and clean up old workflows

set -e

echo "🚀 Starting migration to new branching strategy..."
echo "=================================================="

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
        echo "❌ Migration cancelled"
        exit 1
    fi
fi

echo ""
echo "📋 Migration Steps:"
echo "=================="

# Step 1: Create dev branch from current main
echo "1️⃣  Creating dev branch..."
if git show-ref --verify --quiet refs/heads/dev; then
    echo "   ⚠️  dev branch already exists, switching to it..."
    git checkout dev
    git pull origin dev
else
    echo "   Creating new dev branch from current state..."
    git checkout -b dev
    git push -u origin dev
fi

# Step 2: Backup current main branch
echo ""
echo "2️⃣  Backing up current main branch..."
if git show-ref --verify --quiet refs/heads/main; then
    echo "   Renaming current main to main-backup..."
    git checkout main
    git branch -m main main-backup
    echo "   ✅ Current main branch backed up as main-backup"
else
    echo "   ⚠️  No main branch found, creating from current state..."
fi

# Step 3: Create new main branch
echo ""
echo "3️⃣  Creating new main branch..."
git checkout -b main
echo "   ✅ New main branch created"

# Step 4: Push new branches
echo ""
echo "4️⃣  Pushing new branches to remote..."
git push -u origin main
echo "   ✅ New main branch pushed to remote"

# Step 5: Update .gitignore if needed
echo ""
echo "5️⃣  Checking .gitignore..."
if ! grep -q "main-backup" .gitignore 2>/dev/null; then
    echo "   Adding main-backup to .gitignore..."
    echo "" >> .gitignore
    echo "# Backup branches from migration" >> .gitignore
    echo "main-backup" >> .gitignore
    git add .gitignore
    git commit -m "chore: add backup branches to gitignore"
fi

# Step 6: Create initial commit on main
echo ""
echo "6️⃣  Creating initial commit on main..."
if [ -z "$(git log --oneline -1)" ]; then
    echo "   Creating initial commit..."
    git commit --allow-empty -m "chore: initial commit for new main branch"
    git push origin main
fi

echo ""
echo "✅ Migration completed successfully!"
echo "=================================="
echo ""
echo "📋 Next steps:"
echo "=============="
echo ""
echo "1. Update GitHub repository settings:"
echo "   - Go to Settings > Branches"
echo "   - Change default branch to 'main'"
echo "   - Set up branch protection rules"
echo ""
echo "2. Update branch protection:"
echo "   - Protect 'main' branch (require PR reviews)"
echo "   - Protect 'dev' branch (require CI to pass)"
echo ""
echo "3. Clean up old workflows:"
echo "   - Archive or remove old release workflows"
echo "   - Keep the new ci.yml, release-please.yml, and deploy.yml"
echo ""
echo "4. Update team workflow:"
echo "   - Switch to working on 'dev' branch for development"
echo "   - Use conventional commits for automatic releases"
echo ""
echo "5. Test the new workflow:"
echo "   - Make a small change on dev branch"
echo "   - Use conventional commit message (e.g., 'feat: test new workflow')"
echo "   - Push to dev and verify release-please creates a PR to main"
echo ""
echo "📚 Documentation:"
echo "================"
echo "See docs/BRANCHING_STRATEGY.md for detailed information"
echo ""
echo "🎉 Migration complete! Happy coding! 🚀" 