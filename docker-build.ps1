# Docker Build Script with BuildKit Optimizations (PowerShell)
# This script builds Docker images with BuildKit enabled for optimal performance

param(
    [Parameter(HelpMessage="Build type: frontend, backend, unified, or all")]
    [ValidateSet("frontend", "backend", "unified", "all")]
    [string]$Type = "all",
    
    [Parameter(HelpMessage="Image tag/version")]
    [string]$Version = "latest",
    
    [Parameter(HelpMessage="Push images to registry after build")]
    [switch]$Push,
    
    [Parameter(HelpMessage="Docker registry/repository")]
    [string]$Registry = "",
    
    [Parameter(HelpMessage="Clean build cache before building")]
    [switch]$Clean,
    
    [Parameter(HelpMessage="Show help")]
    [switch]$Help
)

# Enable BuildKit
$env:DOCKER_BUILDKIT = 1

# Function to print colored output
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Function to show usage
function Show-Usage {
    Write-Host @"
Docker Build Script with BuildKit Optimizations

USAGE:
    .\docker-build.ps1 [OPTIONS]

OPTIONS:
    -Type <TYPE>        Build type: frontend, backend, unified, or all (default: all)
    -Version <TAG>      Image tag/version (default: latest)
    -Push               Push images to registry after build
    -Registry <REPO>    Docker registry/repository (e.g., docker.io/username)
    -Clean              Clean build cache before building
    -Help               Show this help message

EXAMPLES:
    # Build all images with default settings
    .\docker-build.ps1

    # Build only frontend with specific tag
    .\docker-build.ps1 -Type frontend -Version v1.2.3

    # Build and push to registry
    .\docker-build.ps1 -Type all -Version v1.2.3 -Push -Registry docker.io/myuser

    # Clean cache and rebuild
    .\docker-build.ps1 -Clean -Type unified
"@
}

# Show help if requested
if ($Help) {
    Show-Usage
    exit 0
}

# Function to build an image
function Build-Image {
    param(
        [string]$Dockerfile,
        [string]$ImageName
    )
    
    $fullName = "${ImageName}:${Version}"
    
    if ($Registry) {
        $fullName = "${Registry}/${fullName}"
    }
    
    Write-Info "Building ${fullName}..."
    
    $startTime = Get-Date
    
    try {
        docker build -f $Dockerfile -t $fullName .
        
        if ($LASTEXITCODE -eq 0) {
            $endTime = Get-Date
            $duration = ($endTime - $startTime).TotalSeconds
            Write-Info "✓ Built ${fullName} in $([math]::Round($duration, 2))s"
            
            if ($Push) {
                Write-Info "Pushing ${fullName}..."
                docker push $fullName
                
                if ($LASTEXITCODE -eq 0) {
                    Write-Info "✓ Pushed ${fullName}"
                } else {
                    Write-Error "Failed to push ${fullName}"
                    return $false
                }
            }
            return $true
        } else {
            Write-Error "Failed to build ${fullName}"
            return $false
        }
    } catch {
        Write-Error "Exception while building ${fullName}: $_"
        return $false
    }
}

# Check if Docker is available
try {
    $null = docker version 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Docker is not running or not accessible"
        exit 1
    }
} catch {
    Write-Error "Docker is not installed or not in PATH"
    exit 1
}

# Get Docker version
try {
    $dockerVersion = docker version --format '{{.Server.Version}}' 2>$null
    if (-not $dockerVersion) {
        $dockerVersion = "unknown"
    }
} catch {
    $dockerVersion = "unknown"
}

Write-Info "Docker version: $dockerVersion"
Write-Info "BuildKit: ENABLED"

# Clean cache if requested
if ($Clean) {
    Write-Warning "Cleaning build cache..."
    docker builder prune -f
}

# Verify .dockerignore exists
if (-not (Test-Path ".dockerignore")) {
    Write-Warning ".dockerignore file not found! Builds may be slower."
}

Write-Info "Build configuration:"
Write-Host "  Type: $Type"
Write-Host "  Tag: $Version"
Write-Host "  Push: $Push"
Write-Host "  Registry: $(if ($Registry) { $Registry } else { 'none' })"
Write-Host ""

# Build images based on type
$success = $true

switch ($Type) {
    "frontend" {
        $success = Build-Image -Dockerfile "Dockerfile.frontend" -ImageName "fmdash-frontend"
    }
    "backend" {
        $success = Build-Image -Dockerfile "Dockerfile.backend" -ImageName "fmdash-backend"
    }
    "unified" {
        $success = Build-Image -Dockerfile "Dockerfile" -ImageName "fmdash"
    }
    "all" {
        $success = (Build-Image -Dockerfile "Dockerfile.frontend" -ImageName "fmdash-frontend") -and $success
        $success = (Build-Image -Dockerfile "Dockerfile.backend" -ImageName "fmdash-backend") -and $success
        $success = (Build-Image -Dockerfile "Dockerfile" -ImageName "fmdash") -and $success
    }
}

if ($success) {
    Write-Info "Build complete!"
} else {
    Write-Error "Build completed with errors"
    exit 1
}

# Show cache statistics
Write-Info "Build cache statistics:"
docker system df

exit 0



