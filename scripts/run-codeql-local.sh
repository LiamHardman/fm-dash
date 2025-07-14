#!/bin/bash

set -e

echo "🔒 Running CodeQL Security Analysis Locally"
echo "============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create CodeQL databases directory
CODEQL_DB_DIR="./codeql-databases"
mkdir -p "$CODEQL_DB_DIR"

# Check if CodeQL is installed
if ! command -v codeql &> /dev/null; then
    echo -e "${RED}❌ CodeQL CLI not found. Please install it first:${NC}"
    echo "   brew install codeql"
    exit 1
fi

echo -e "${BLUE}📋 CodeQL Setup:${NC}"
echo "   • Languages: Go, JavaScript/TypeScript"
echo "   • Queries: security-and-quality"
echo "   • Config: .github/codeql/codeql-config.yml"
echo ""

# Function to create and analyze Go database
analyze_go() {
    echo -e "${YELLOW}🔍 Analyzing Go Backend...${NC}"
    
    # Ensure Go dependencies are available
    echo "   📦 Installing Go dependencies..."
    cd src/api && go mod tidy && cd ../..
    
    # Create database
    echo "   🗄️  Creating CodeQL database for Go..."
    codeql database create \
        "$CODEQL_DB_DIR/go-database" \
        --language=go \
        --source-root=src/api \
        --overwrite \
        --working-dir=src/api \
        --command="go build ./..."
    
    # Run analysis
    echo "   🔬 Running CodeQL analysis for Go..."
    codeql database analyze \
        "$CODEQL_DB_DIR/go-database" \
        --format=sarif-latest \
        --output="$CODEQL_DB_DIR/go-results.sarif" \
        --sarif-category="go" \
        --download \
        codeql/go-queries:codeql-suites/go-security-and-quality.qls
    
    echo -e "${GREEN}   ✅ Go analysis complete${NC}"
}

# Function to create and analyze JavaScript database  
analyze_javascript() {
    echo -e "${YELLOW}🔍 Analyzing JavaScript/TypeScript Frontend...${NC}"
    
    # Ensure Node dependencies are available
    echo "   📦 Installing Node.js dependencies..."
    npm ci --legacy-peer-deps
    
    # Create database
    echo "   🗄️  Creating CodeQL database for JavaScript..."
    codeql database create \
        "$CODEQL_DB_DIR/javascript-database" \
        --language=javascript \
        --source-root=. \
        --overwrite
    
    # Run analysis
    echo "   🔬 Running CodeQL analysis for JavaScript..."
    codeql database analyze \
        "$CODEQL_DB_DIR/javascript-database" \
        --format=sarif-latest \
        --output="$CODEQL_DB_DIR/javascript-results.sarif" \
        --sarif-category="javascript" \
        --download \
        codeql/javascript-queries:codeql-suites/javascript-security-and-quality.qls
    
    echo -e "${GREEN}   ✅ JavaScript analysis complete${NC}"
}

# Function to display results summary
show_results() {
    echo ""
    echo -e "${BLUE}📊 Analysis Results:${NC}"
    echo "   📁 Results directory: $CODEQL_DB_DIR/"
    echo "   📄 Go results: $CODEQL_DB_DIR/go-results.sarif"
    echo "   📄 JavaScript results: $CODEQL_DB_DIR/javascript-results.sarif"
    echo ""
    
    # Count issues if jq is available
    if command -v jq &> /dev/null; then
        echo -e "${YELLOW}🔍 Issue Summary:${NC}"
        
        if [ -f "$CODEQL_DB_DIR/go-results.sarif" ]; then
            GO_ISSUES=$(jq '.runs[0].results | length' "$CODEQL_DB_DIR/go-results.sarif" 2>/dev/null || echo "0")
            echo "   Go: $GO_ISSUES issues found"
        fi
        
        if [ -f "$CODEQL_DB_DIR/javascript-results.sarif" ]; then
            JS_ISSUES=$(jq '.runs[0].results | length' "$CODEQL_DB_DIR/javascript-results.sarif" 2>/dev/null || echo "0")
            echo "   JavaScript: $JS_ISSUES issues found"
        fi
    fi
    
    echo ""
    echo -e "${BLUE}💡 To view detailed results:${NC}"
    echo "   • Open SARIF files in VS Code with CodeQL extension"
    echo "   • Upload to GitHub Security tab (manual upload)"
    echo "   • Use: codeql database analyze --format=csv for CSV output"
}

# Main execution
echo -e "${YELLOW}🚀 Starting CodeQL analysis...${NC}"

# Analyze both languages
analyze_go
analyze_javascript

# Show results
show_results

echo ""
echo -e "${GREEN}🎉 CodeQL analysis complete!${NC}" 