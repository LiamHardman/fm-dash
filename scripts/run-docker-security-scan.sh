#!/bin/bash

set -e

echo "🐳 Docker Security Scanning (GitHub Actions equivalent)"
echo "======================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
RESULTS_DIR="./docker-security-results"
IMAGE_NAME="ghcr.io/liamhardman/fm-dash:latest"
LOCAL_IMAGE_TAG="fm-dash:security-test"

# Create results directory
mkdir -p "$RESULTS_DIR"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install Trivy
install_trivy() {
    echo -e "${YELLOW}📦 Installing Trivy vulnerability scanner...${NC}"
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command_exists brew; then
            brew install trivy
        else
            echo -e "${RED}❌ Homebrew not found. Please install Homebrew first.${NC}"
            exit 1
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
    else
        echo -e "${RED}❌ Unsupported OS: $OSTYPE${NC}"
        exit 1
    fi
}

# Function to install Hadolint (Dockerfile linter)
install_hadolint() {
    echo -e "${YELLOW}📦 Installing Hadolint Dockerfile linter...${NC}"
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        if command_exists brew; then
            brew install hadolint
        else
            echo -e "${YELLOW}⚠️ Homebrew not found, downloading hadolint directly...${NC}"
            curl -L -o /usr/local/bin/hadolint https://github.com/hadolint/hadolint/releases/latest/download/hadolint-Darwin-x86_64
            chmod +x /usr/local/bin/hadolint
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        curl -L -o /usr/local/bin/hadolint https://github.com/hadolint/hadolint/releases/latest/download/hadolint-Linux-x86_64
        chmod +x /usr/local/bin/hadolint
    fi
}

# Check and install required tools
echo -e "${BLUE}🔧 Checking required tools...${NC}"

if ! command_exists trivy; then
    install_trivy
fi

if ! command_exists hadolint; then
    install_hadolint
fi

if ! command_exists docker; then
    echo -e "${RED}❌ Docker not found. Please install Docker first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All required tools installed${NC}"
echo ""

# Function to scan Dockerfile with Hadolint
scan_dockerfile() {
    local dockerfile=$1
    local output_file="$RESULTS_DIR/hadolint-$(basename $dockerfile).json"
    
    echo -e "${YELLOW}🔍 Scanning $dockerfile with Hadolint...${NC}"
    
    if [ -f "$dockerfile" ]; then
        hadolint --format json "$dockerfile" > "$output_file" 2>/dev/null || true
        
        # Count issues
        local issues=0
        if [ -f "$output_file" ] && [ -s "$output_file" ]; then
            issues=$(jq '. | length' "$output_file" 2>/dev/null || echo "0")
        fi
        
        echo "   📄 $dockerfile: $issues Dockerfile issues found"
        echo "   📁 Results: $output_file"
    else
        echo -e "${YELLOW}   ⚠️ $dockerfile not found, skipping...${NC}"
    fi
}

# Function to build image locally for scanning
build_local_image() {
    local dockerfile=$1
    local tag=$2
    
    echo -e "${YELLOW}🏗️ Building local image for security scanning...${NC}"
    echo "   📄 Using: $dockerfile"
    echo "   🏷️ Tag: $tag"
    
    if docker build -f "$dockerfile" -t "$tag" . >/dev/null 2>&1; then
        echo -e "${GREEN}   ✅ Build successful${NC}"
        return 0
    else
        echo -e "${RED}   ❌ Build failed${NC}"
        return 1
    fi
}

# Function to scan image with Trivy (GitHub Actions equivalent)
scan_image_trivy() {
    local image=$1
    local scan_type=$2  # "local" or "remote"
    
    echo -e "${YELLOW}🔍 Running Trivy vulnerability scan ($scan_type image)...${NC}"
    echo "   🎯 Target: $image"
    
    local base_filename
    if [ "$scan_type" = "remote" ]; then
        base_filename="trivy-remote"
    else
        base_filename="trivy-local"
    fi
    
    # SARIF output (matches GitHub Actions)
    echo "   📋 Generating SARIF report (GitHub Actions format)..."
    trivy image --format sarif --output "$RESULTS_DIR/${base_filename}-results.sarif" "$image" 2>/dev/null || true
    
    # JSON output for detailed analysis
    echo "   📋 Generating JSON report..."
    trivy image --format json --output "$RESULTS_DIR/${base_filename}-results.json" "$image" 2>/dev/null || true
    
    # Table output for console viewing
    echo "   📋 Generating human-readable report..."
    trivy image --format table --output "$RESULTS_DIR/${base_filename}-results.txt" "$image" 2>/dev/null || true
    
    # Count vulnerabilities
    local critical=0 high=0 medium=0 low=0 unknown=0
    if [ -f "$RESULTS_DIR/${base_filename}-results.json" ]; then
        critical=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "CRITICAL")] | length' "$RESULTS_DIR/${base_filename}-results.json" 2>/dev/null || echo "0")
        high=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "HIGH")] | length' "$RESULTS_DIR/${base_filename}-results.json" 2>/dev/null || echo "0")
        medium=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "MEDIUM")] | length' "$RESULTS_DIR/${base_filename}-results.json" 2>/dev/null || echo "0")
        low=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "LOW")] | length' "$RESULTS_DIR/${base_filename}-results.json" 2>/dev/null || echo "0")
        unknown=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "UNKNOWN")] | length' "$RESULTS_DIR/${base_filename}-results.json" 2>/dev/null || echo "0")
    fi
    
    echo "   🎯 Vulnerabilities found:"
    echo "      • Critical: $critical"
    echo "      • High: $high"
    echo "      • Medium: $medium"
    echo "      • Low: $low"
    echo "      • Unknown: $unknown"
    
    return 0
}

# Function to scan for secrets
scan_image_secrets() {
    local image=$1
    local scan_type=$2
    
    echo -e "${YELLOW}🔍 Scanning for secrets and sensitive data...${NC}"
    
    local base_filename
    if [ "$scan_type" = "remote" ]; then
        base_filename="trivy-secrets-remote"
    else
        base_filename="trivy-secrets-local"
    fi
    
    trivy image --scanners secret --format json --output "$RESULTS_DIR/${base_filename}-results.json" "$image" 2>/dev/null || true
    
    local secrets=0
    if [ -f "$RESULTS_DIR/${base_filename}-results.json" ]; then
        secrets=$(jq '[.Results[]?.Secrets[]?] | length' "$RESULTS_DIR/${base_filename}-results.json" 2>/dev/null || echo "0")
    fi
    
    echo "   🔐 Secrets found: $secrets"
}

# Function to display summary
show_summary() {
    echo ""
    echo -e "${BLUE}📊 Docker Security Scan Summary${NC}"
    echo "================================="
    echo ""
    echo -e "${YELLOW}📁 Results directory: $RESULTS_DIR/${NC}"
    echo ""
    
    echo -e "${BLUE}📄 Generated Reports:${NC}"
    for file in "$RESULTS_DIR"/*; do
        if [ -f "$file" ]; then
            echo "   • $(basename "$file")"
        fi
    done
    
    echo ""
    echo -e "${BLUE}🔍 How to view results:${NC}"
    echo "   1. SARIF (GitHub Actions format): Upload to GitHub Security tab"
    echo "   2. JSON: Open with jq for detailed analysis"
    echo "   3. TXT: Human-readable console output"
    echo ""
    echo -e "${BLUE}💡 Next steps:${NC}"
    echo "   • Review Dockerfile issues: hadolint reports"
    echo "   • Address critical/high vulnerabilities in Trivy reports"
    echo "   • Check secrets scan for exposed credentials"
    echo "   • Upload SARIF files to GitHub for centralized tracking"
    echo ""
    echo -e "${GREEN}🎉 Docker security scan complete!${NC}"
}

# Main execution
echo -e "${YELLOW}🚀 Starting Docker security analysis...${NC}"
echo ""

# Step 1: Dockerfile linting
echo -e "${BLUE}📋 Step 1: Dockerfile Security Linting${NC}"
scan_dockerfile "Dockerfile"
scan_dockerfile "Dockerfile.backend" 
scan_dockerfile "Dockerfile.frontend"
echo ""

# Step 2: Build local image for testing
echo -e "${BLUE}📋 Step 2: Building Local Image${NC}"
if build_local_image "Dockerfile" "$LOCAL_IMAGE_TAG"; then
    LOCAL_BUILD_SUCCESS=true
else
    LOCAL_BUILD_SUCCESS=false
fi
echo ""

# Step 3: Scan local image (if built successfully)
if [ "$LOCAL_BUILD_SUCCESS" = true ]; then
    echo -e "${BLUE}📋 Step 3: Local Image Security Scan${NC}"
    scan_image_trivy "$LOCAL_IMAGE_TAG" "local"
    scan_image_secrets "$LOCAL_IMAGE_TAG" "local"
    echo ""
else
    echo -e "${YELLOW}⚠️ Skipping local image scan due to build failure${NC}"
    echo ""
fi

# Step 4: Scan remote image (GitHub Actions equivalent)
echo -e "${BLUE}📋 Step 4: Remote Image Security Scan (GitHub Actions equivalent)${NC}"
echo "🔗 Attempting to scan: $IMAGE_NAME"

# Check if remote image exists and is accessible
if docker pull "$IMAGE_NAME" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Remote image accessible${NC}"
    scan_image_trivy "$IMAGE_NAME" "remote"
    scan_image_secrets "$IMAGE_NAME" "remote"
    
    # Clean up pulled image
    docker rmi "$IMAGE_NAME" >/dev/null 2>&1 || true
else
    echo -e "${YELLOW}⚠️ Remote image not accessible (might be private or not exist)${NC}"
    echo "   💡 This is normal for private repositories or unpublished images"
fi

# Clean up local test image
if [ "$LOCAL_BUILD_SUCCESS" = true ]; then
    docker rmi "$LOCAL_IMAGE_TAG" >/dev/null 2>&1 || true
fi

# Show summary
show_summary 