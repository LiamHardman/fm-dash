#!/bin/bash

set -e

echo "🔒 Trivy Security Scan (GitHub Actions equivalent)"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration - matches your GitHub Actions exactly
IMAGE_NAME="ghcr.io/liamhardman/fm-dash:latest"
RESULTS_DIR="./docker-security-results"
SARIF_OUTPUT="$RESULTS_DIR/trivy-results.sarif"

# Create results directory
mkdir -p "$RESULTS_DIR"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install Trivy if needed
if ! command_exists trivy; then
    echo -e "${YELLOW}📦 Installing Trivy...${NC}"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install trivy
    else
        curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
    fi
fi

echo -e "${BLUE}🎯 Target Image: $IMAGE_NAME${NC}"
echo -e "${BLUE}📁 Results: $SARIF_OUTPUT${NC}"
echo ""

# Build local image if remote not accessible
LOCAL_IMAGE_TAG="fm-dash:trivy-scan"
echo -e "${YELLOW}🏗️ Building local image for scanning...${NC}"

if docker build -t "$LOCAL_IMAGE_TAG" . >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Local build successful${NC}"
    SCAN_TARGET="$LOCAL_IMAGE_TAG"
    SCAN_TYPE="local"
else
    echo -e "${RED}❌ Local build failed${NC}"
    
    # Try to pull remote image
    echo -e "${YELLOW}🔄 Attempting to pull remote image...${NC}"
    if docker pull "$IMAGE_NAME" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Remote image pulled successfully${NC}"
        SCAN_TARGET="$IMAGE_NAME"
        SCAN_TYPE="remote"
    else
        echo -e "${RED}❌ Cannot access remote image${NC}"
        echo -e "${YELLOW}💡 Make sure you're logged in: docker login ghcr.io${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${YELLOW}🔍 Running Trivy vulnerability scan ($SCAN_TYPE image)...${NC}"

# Run Trivy scan - EXACT same command as GitHub Actions
trivy image --format 'sarif' --output "$SARIF_OUTPUT" "$SCAN_TARGET"

# Additional formats for local analysis
echo -e "${YELLOW}📋 Generating additional reports...${NC}"
trivy image --format json --output "$RESULTS_DIR/trivy-results.json" "$SCAN_TARGET" 2>/dev/null || true
trivy image --format table --output "$RESULTS_DIR/trivy-results.txt" "$SCAN_TARGET" 2>/dev/null || true

# Count vulnerabilities
if command_exists jq && [ -f "$RESULTS_DIR/trivy-results.json" ]; then
    echo ""
    echo -e "${BLUE}📊 Vulnerability Summary:${NC}"
    critical=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "CRITICAL")] | length' "$RESULTS_DIR/trivy-results.json" 2>/dev/null || echo "0")
    high=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "HIGH")] | length' "$RESULTS_DIR/trivy-results.json" 2>/dev/null || echo "0")
    medium=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "MEDIUM")] | length' "$RESULTS_DIR/trivy-results.json" 2>/dev/null || echo "0")
    low=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "LOW")] | length' "$RESULTS_DIR/trivy-results.json" 2>/dev/null || echo "0")
    
    echo "   🔴 Critical: $critical"
    echo "   🟠 High: $high"
    echo "   🟡 Medium: $medium"
    echo "   🟢 Low: $low"
fi

# Cleanup
if [ "$SCAN_TYPE" = "local" ]; then
    docker rmi "$LOCAL_IMAGE_TAG" >/dev/null 2>&1 || true
elif [ "$SCAN_TYPE" = "remote" ]; then
    docker rmi "$IMAGE_NAME" >/dev/null 2>&1 || true
fi

echo ""
echo -e "${GREEN}✅ Trivy scan complete!${NC}"
echo -e "${BLUE}📄 SARIF results: $SARIF_OUTPUT${NC}"
echo ""
echo -e "${YELLOW}💡 To upload to GitHub Security tab:${NC}"
echo "   1. Go to your repo's Security tab"
echo "   2. Click 'Code scanning alerts'"  
echo "   3. Click 'Upload SARIF file'"
echo "   4. Upload: $SARIF_OUTPUT"
echo ""
echo -e "${YELLOW}📖 To view locally:${NC}"
echo "   • Human readable: cat $RESULTS_DIR/trivy-results.txt"
echo "   • JSON analysis: jq . $RESULTS_DIR/trivy-results.json" 