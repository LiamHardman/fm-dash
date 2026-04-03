#!/bin/bash

# Docker Build Script with BuildKit Optimizations
# This script builds Docker images with BuildKit enabled for optimal performance

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Enable BuildKit
export DOCKER_BUILDKIT=1

# Default values
IMAGE_TAG="latest"
BUILD_TYPE="all"
PUSH=false
REGISTRY=""

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build Docker images with BuildKit optimizations enabled.

OPTIONS:
    -t, --type TYPE        Build type: frontend, backend, unified, or all (default: all)
    -v, --version TAG      Image tag/version (default: latest)
    -p, --push             Push images to registry after build
    -r, --registry REPO    Docker registry/repository (e.g., docker.io/username)
    -c, --clean            Clean build cache before building
    -h, --help             Show this help message

EXAMPLES:
    # Build all images with default settings
    $0

    # Build only frontend with specific tag
    $0 --type frontend --version v1.2.3

    # Build and push to registry
    $0 --type all --version v1.2.3 --push --registry docker.io/myuser

    # Clean cache and rebuild
    $0 --clean --type unified

EOF
}

# Function to build an image
build_image() {
    local dockerfile=$1
    local image_name=$2
    local full_name="${image_name}:${IMAGE_TAG}"
    
    if [ -n "$REGISTRY" ]; then
        full_name="${REGISTRY}/${full_name}"
    fi
    
    print_info "Building ${full_name}..."
    
    local start_time=$(date +%s)
    
    if docker build -f "$dockerfile" -t "$full_name" .; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        print_info "✓ Built ${full_name} in ${duration}s"
        
        if [ "$PUSH" = true ]; then
            print_info "Pushing ${full_name}..."
            if docker push "$full_name"; then
                print_info "✓ Pushed ${full_name}"
            else
                print_error "Failed to push ${full_name}"
                return 1
            fi
        fi
    else
        print_error "Failed to build ${full_name}"
        return 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            BUILD_TYPE="$2"
            shift 2
            ;;
        -v|--version)
            IMAGE_TAG="$2"
            shift 2
            ;;
        -p|--push)
            PUSH=true
            shift
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -c|--clean)
            print_warning "Cleaning build cache..."
            docker builder prune -f
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Validate build type
case $BUILD_TYPE in
    frontend|backend|unified|all)
        ;;
    *)
        print_error "Invalid build type: $BUILD_TYPE"
        usage
        exit 1
        ;;
esac

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed or not in PATH"
    exit 1
fi

# Check Docker version and BuildKit support
DOCKER_VERSION=$(docker version --format '{{.Server.Version}}' 2>/dev/null || echo "unknown")
print_info "Docker version: $DOCKER_VERSION"
print_info "BuildKit: ENABLED"

# Verify .dockerignore exists
if [ ! -f ".dockerignore" ]; then
    print_warning ".dockerignore file not found! Builds may be slower."
fi

print_info "Build configuration:"
echo "  Type: $BUILD_TYPE"
echo "  Tag: $IMAGE_TAG"
echo "  Push: $PUSH"
echo "  Registry: ${REGISTRY:-none}"
echo ""

# Build images based on type
case $BUILD_TYPE in
    frontend)
        build_image "Dockerfile.frontend" "fmdash-frontend"
        ;;
    backend)
        build_image "Dockerfile.backend" "fmdash-backend"
        ;;
    unified)
        build_image "Dockerfile" "fmdash"
        ;;
    all)
        build_image "Dockerfile.frontend" "fmdash-frontend"
        build_image "Dockerfile.backend" "fmdash-backend"
        build_image "Dockerfile" "fmdash"
        ;;
esac

print_info "Build complete!"

# Show cache statistics
print_info "Build cache statistics:"
docker system df --format "table {{.Type}}\t{{.TotalCount}}\t{{.Size}}\t{{.Reclaimable}}"



