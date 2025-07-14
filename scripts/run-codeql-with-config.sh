#!/bin/bash

set -e

echo "🔒 Running CodeQL Security Analysis (with custom config)"
echo "========================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create CodeQL databases directory
CODEQL_DB_DIR="./codeql-databases"
CONFIG_FILE=".github/codeql/codeql-config.yml"
mkdir -p "$CODEQL_DB_DIR"

# Check if CodeQL is installed
if ! command -v codeql &> /dev/null; then
    echo -e "${RED}❌ CodeQL CLI not found. Please install it first:${NC}"
    echo "   brew install codeql"
    exit 1
fi

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}❌ CodeQL config file not found: $CONFIG_FILE${NC}"
    echo "   Falling back to default configuration..."
    CONFIG_FILE=""
fi

echo -e "${BLUE}📋 CodeQL Setup:${NC}"
echo "   • Languages: Go, JavaScript/TypeScript"
echo "   • Queries: security-extended, security-and-quality"
echo "   • Config: $CONFIG_FILE"
echo "   • Filters: Custom false positive filters applied"
echo ""

# Function to create and analyze with config
analyze_with_config() {
    local language=$1
    local build_command=$2
    
    echo -e "${YELLOW}🔍 Analyzing $language with custom config...${NC}"
    
    # Setup dependencies based on language
    if [ "$language" == "go" ]; then
        echo "   📦 Installing Go dependencies..."
        cd src/api && go mod tidy && cd ../..
    elif [ "$language" == "javascript" ]; then
        echo "   📦 Installing Node.js dependencies..."
        npm ci --legacy-peer-deps
    fi
    
    # Create database with config
    echo "   🗄️  Creating CodeQL database for $language..."
    
    if [ -n "$CONFIG_FILE" ]; then
        # Use custom config
        if [ -n "$build_command" ]; then
            if [ "$language" == "go" ]; then
                # Go-specific database creation
                codeql database create \
                    "$CODEQL_DB_DIR/$language-database" \
                    --language="$language" \
                    --source-root=src/api \
                    --overwrite \
                    --working-dir=src/api \
                    --command="$build_command" \
                    --codescanning-config="$CONFIG_FILE"
            else
                codeql database create \
                    "$CODEQL_DB_DIR/$language-database" \
                    --language="$language" \
                    --source-root=. \
                    --overwrite \
                    --command="$build_command" \
                    --codescanning-config="$CONFIG_FILE"
            fi
        else
            codeql database create \
                "$CODEQL_DB_DIR/$language-database" \
                --language="$language" \
                --source-root=. \
                --overwrite \
                --codescanning-config="$CONFIG_FILE"
        fi
    else
        # Fallback without config
        if [ -n "$build_command" ]; then
            if [ "$language" == "go" ]; then
                # Go-specific database creation
                codeql database create \
                    "$CODEQL_DB_DIR/$language-database" \
                    --language="$language" \
                    --source-root=src/api \
                    --overwrite \
                    --working-dir=src/api \
                    --command="$build_command"
            else
                codeql database create \
                    "$CODEQL_DB_DIR/$language-database" \
                    --language="$language" \
                    --source-root=. \
                    --overwrite \
                    --command="$build_command"
            fi
        else
            codeql database create \
                "$CODEQL_DB_DIR/$language-database" \
                --language="$language" \
                --source-root=. \
                --overwrite
        fi
    fi
    
    # Run analysis with custom queries
    echo "   🔬 Running CodeQL analysis for $language..."
    
    # Use the same query packs as defined in the config
    local query_packs=""
    if [ "$language" == "go" ]; then
        query_packs="codeql/go-queries:codeql-suites/go-security-extended.qls codeql/go-queries:codeql-suites/go-security-and-quality.qls"
    elif [ "$language" == "javascript" ]; then
        query_packs="codeql/javascript-queries:codeql-suites/javascript-security-extended.qls codeql/javascript-queries:codeql-suites/javascript-security-and-quality.qls"
    fi
    
    codeql database analyze \
        "$CODEQL_DB_DIR/$language-database" \
        --format=sarif-latest \
        --output="$CODEQL_DB_DIR/$language-results.sarif" \
        --sarif-category="$language" \
        --download \
        $query_packs
    
    echo -e "${GREEN}   ✅ $language analysis complete${NC}"
}

# Function to display detailed results
show_detailed_results() {
    echo ""
    echo -e "${BLUE}📊 Detailed Analysis Results:${NC}"
    echo "   📁 Results directory: $CODEQL_DB_DIR/"
    
    # Show SARIF files
    for file in "$CODEQL_DB_DIR"/*.sarif; do
        if [ -f "$file" ]; then
            echo "   📄 $(basename "$file")"
        fi
    done
    
    echo ""
    
    # Detailed issue analysis if jq is available
    if command -v jq &> /dev/null; then
        echo -e "${YELLOW}🔍 Security Issue Breakdown:${NC}"
        
        for sarif_file in "$CODEQL_DB_DIR"/*.sarif; do
            if [ -f "$sarif_file" ]; then
                local filename=$(basename "$sarif_file" .sarif)
                local total_issues=$(jq '.runs[0].results | length' "$sarif_file" 2>/dev/null || echo "0")
                
                echo "   ${filename^}: $total_issues total issues"
                
                # Break down by severity if available
                local high=$(jq '.runs[0].results | map(select(.level == "error" or .properties.severity == "high")) | length' "$sarif_file" 2>/dev/null || echo "0")
                local medium=$(jq '.runs[0].results | map(select(.level == "warning" or .properties.severity == "medium")) | length' "$sarif_file" 2>/dev/null || echo "0")
                local low=$(jq '.runs[0].results | map(select(.level == "note" or .properties.severity == "low")) | length' "$sarif_file" 2>/dev/null || echo "0")
                
                echo "     • High severity: $high"
                echo "     • Medium severity: $medium"  
                echo "     • Low severity: $low"
                echo ""
            fi
        done
    else
        echo -e "${YELLOW}💡 Install jq for detailed issue breakdown: brew install jq${NC}"
    fi
    
    echo -e "${BLUE}📖 How to view results:${NC}"
    echo "   1. VS Code: Install CodeQL extension, open SARIF files"
    echo "   2. GitHub: Upload SARIF files to Security > Code scanning alerts"
    echo "   3. Command line: Use 'codeql database analyze --format=csv' for CSV output"
    echo "   4. Web viewer: Use 'codeql database serve <database>' for web interface"
}

# Main execution
echo -e "${YELLOW}🚀 Starting CodeQL analysis with custom configuration...${NC}"

# Analyze Go with build command
analyze_with_config "go" "go build ./..."

# Analyze JavaScript (no build command needed)
analyze_with_config "javascript" ""

# Show detailed results
show_detailed_results

echo ""
echo -e "${GREEN}🎉 CodeQL analysis with custom config complete!${NC}"
echo -e "${BLUE}🔍 This analysis mirrors your GitHub Actions CodeQL setup exactly.${NC}" 