# FM-Dash

<div align="center">

![FM-Dash Logo](https://img.shields.io/badge/FM--Dash-Football%20Manager%20Data%20Analysis-blue?style=for-the-badge)

**A comprehensive platform for analyzing Football Manager player data with enterprise-grade performance and modern web architecture**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue.js-3.0+-4FC08D?style=flat-square&logo=vue.js)](https://vuejs.org)
[![Quasar Version](https://img.shields.io/badge/Quasar-2.0+-1976D2?style=flat-square&logo=quasar)](https://quasar.dev)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

[![CI/CD](https://github.com/LiamHardman/fm-dash/actions/workflows/code-quality.yml/badge.svg)](https://github.com/LiamHardman/fm-dash/actions/workflows/code-quality.yml)
[![Release](https://github.com/LiamHardman/fm-dash/actions/workflows/release.yml/badge.svg)](https://github.com/LiamHardman/fm-dash/actions/workflows/release.yml)
[![Docker](https://github.com/LiamHardman/fm-dash/actions/workflows/deploy.yml/badge.svg)](https://github.com/LiamHardman/fm-dash/actions/workflows/deploy.yml)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/LiamHardman/fm-dash?sort=semver&style=flat-square)](https://github.com/LiamHardman/fm-dash/releases)
[![GitHub issues](https://img.shields.io/github/issues/LiamHardman/fm-dash?style=flat-square)](https://github.com/LiamHardman/fm-dash/issues)
[![GitHub stars](https://img.shields.io/github/stars/LiamHardman/fm-dash?style=flat-square)](https://github.com/LiamHardman/fm-dash/stargazers)

</div>

## Overview

FM-Dash is a sophisticated Football Manager data analysis platform that transforms HTML export data into a one-stop shop for finding your best next signings. Built with a modern Vue.js frontend and high-performance Go backend, it provides scouts, managers, and analysts with powerful tools for player evaluation, team building, and strategic planning.

## Core Features

### 🔍 **Advanced Player Analysis**
- **Specialized Search Tools**: Bargain Hunter, Wonderkids Discovery, Free Agents and Upgrade Finder dialogs
- **Multi-dimensional Filtering**: Position, nationality, league, club, age, and attribute-based filters
- **Performance Metrics**: FIFA-style ratings, percentile rankings, and custom calculation algorithms
- **Detailed Player Profiles**: Comprehensive attribute analysis with visual representations
- **Pitch Visualization**: Interactive formation display with player positioning

### 📊 **Intelligent Data Processing**
- **High-Performance Parser**: Stream-based HTML processing capable of handling 50MB+ files
- **Web Workers**: Background processing for calculations without UI blocking
- **Scalable Ratings**: Toggle between raw and scaled attribute systems
- **Smart Caching**: Multi-level caching with configurable retention policies
- **Nationality Processing**: Advanced nationality detection and filtering algorithms

### 🎯 **Team Management Tools**
- **Wishlist System**: Personal player tracking with persistent storage
- **Team Logo Integration**: Automatic logo matching with fuzzy search algorithms
- **Formation Analysis**: Interactive pitch display with tactical insights
- **League Overviews**: Comprehensive league and team statistics
- **Player Photography**: Support for player face images with unique ID matching

### ⚡ **Performance-Optimized Frontend**
- **Bundle Optimization**: Advanced code splitting with 40%+ bundle size reduction (in progress)
- **Third-Party Library Optimization**: Tree-shakeable imports and dynamic loading (in progress)
- **Virtual Scrolling**: Smooth 60fps performance with 10,000+ player datasets
- **Memory Management**: Object pooling and LRU caching for efficient memory usage
- **Mobile Optimization**: Touch-optimized interactions with battery efficiency
- **Image Loading**: Modern WebP/AVIF formats with lazy loading and progressive enhancement
- **Core Web Vitals**: Comprehensive performance monitoring and optimization

### 🔧 **Enterprise-Ready Infrastructure**
- **OpenTelemetry Integration**: Comprehensive metrics, tracing, and logging with SignOz compatibility
- **S3-Compatible Storage**: Scalable object storage with MinIO/AWS S3 support
- **Kubernetes Deployment**: Production-ready orchestration with health checks
- **Performance Monitoring**: Built-in profiling, metrics collection, and system diagnostics
- **Security Features**: Input validation, CORS protection, and rate limiting

## Technical Architecture

### Frontend Stack
- **Framework**: Vue 3 with Composition API and `<script setup>` syntax
- **UI Framework**: Quasar Framework with Material Design components
- **State Management**: Pinia with modular store architecture (`playerStore`, `uiStore`, `wishlistStore`)
- **Build System**: Vite with optimized hot module replacement
- **Testing**: Vitest with comprehensive component and unit tests
- **Code Quality**: Biome for unified linting and formatting

### Backend Infrastructure
- **Runtime**: Go 1.24+ with modern concurrency patterns
- **HTTP Layer**: Native Go `net/http` with custom middleware stack
- **File Processing**: Advanced HTML parsing using `golang.org/x/net/html`
- **Storage**: S3-compatible object storage with intelligent caching
- **Observability**: Full OpenTelemetry implementation with metrics, traces, and logs
- **Performance**: Optimized for large datasets with memory-efficient processing

### Deployment & Operations
- **Containerization**: Multi-stage Docker builds with Nginx + supervisord
- **Orchestration**: Kubernetes manifests with production-ready configurations
- **Monitoring**: Integrated health checks, metrics endpoints, and performance profiling
- **CI/CD**: Automated testing, code quality checks, and deployment pipelines


## Credits
- sortitoutsi, and Footygamer for providing the club + face graphics in the publicly-hosted version of FM-Dash, along with some very useful general advice!

## Quick Start

### Prerequisites

Ensure you have the following installed:
- **Go** 1.24 or higher
- **Node.js** 18 or higher  
- **Docker** (for containerized deployment)
- **Kubernetes** (for production deployment)

### Development Setup

1. **Clone and Initialize**
   ```bash
   git clone https://github.com/LiamHardman/fm-dash.git
   cd fm-dash
   ./scripts/setup-dev.sh
   ```

2. **Development Mode** (with hot reload)
   ```bash
   # Terminal 1: Backend API
   npm run serve
   
   # Terminal 2: Frontend Development Server  
   npm run dev
   ```
   - Frontend: http://localhost:3000
   - API: http://localhost:8091

3. **Production Mode** (containerized)
   ```bash
   docker build -t fm-dash .
   docker run -p 8080:8080 fm-dash
   ```
   - Application: http://localhost:8080

## Feature Deep Dive

### Player Analysis Suite

**Bargain Hunter**: Identifies undervalued players based on market value vs. attributes ratio
- Configurable budget ranges and position filters
- Value-for-money calculations with statistical modeling
- Export capabilities for transfer planning

**Wonderkids Discovery**: Advanced young talent identification system
- Age-based filtering with potential analysis
- Growth trajectory predictions
- Scouting report generation

**Upgrade Finder**: Team improvement recommendations
- Position-specific upgrade suggestions  
- Comparative analysis against current squad
- Transfer budget optimization

### Data Visualization

**Interactive Pitch Display**: 
- Formation-based player positioning
- Real-time tactical analysis
- Player role optimization suggestions

**Performance Analytics**:
- Percentile rankings across leagues
- Attribute distribution charts
- Historical performance tracking

### Advanced Filtering System

Multi-dimensional search capabilities:
- **Positional**: Primary, secondary, and natural positions
- **Geographic**: Nationality with continent grouping
- **Temporal**: Age ranges with career stage analysis
- **Financial**: Market value and wage filtering
- **Performance**: Attribute-based complex queries

## Configuration & Deployment

### Environment Configuration

Key configuration options:

```bash
# Image Storage Configuration
IMAGE_API_URL=https://sortitoutsi.b-cdn.net/uploads  # External CDN for player faces and team logos
FACES_DIR=./faces                                    # Local directory fallback for faces
LOGOS_DIR=./logos                                    # Local directory fallback for logos

# Storage Configuration
MINIO_ENDPOINT=your-s3-endpoint
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-secret-key
MINIO_USE_SSL=true

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=http://your-otlp-endpoint:4317
ENABLE_TRACING=true
ENABLE_METRICS=true

# Performance Tuning
MAX_WORKERS=8
CACHE_TTL=3600
MAX_UPLOAD_SIZE=100MB
```

### Production Deployment

1. **Build and Push Container**
   ```bash
   docker build -t your-registry/fm-dash:latest .
   docker push your-registry/fm-dash:latest
   ```

2. **Configure Kubernetes Secrets**
   ```bash
   kubectl create secret generic fm-dash-config \
     --from-literal=minio-endpoint=your-endpoint \
     --from-literal=minio-access-key=your-key \
     --from-literal=minio-secret-key=your-secret
   ```

3. **Deploy to Kubernetes**
   ```bash
   kubectl apply -f kube.yaml
   ```

## Development Workflow

### Code Quality Standards

```bash
# Full quality check suite
npm run check                    # Lint + format + test all

# Individual quality checks  
npm run lint:all                 # Frontend (Biome) + Backend (golangci-lint)
npm run format                   # Auto-fix formatting issues
npm run test:all                 # Frontend (Vitest) + Backend (Go test)
```


## Contributing

We welcome contributions from the community. Please follow these guidelines:

### Development Standards
- Follow existing code patterns and architectural decisions
- Write comprehensive tests for new features
- Update documentation for significant changes
- Ensure all quality checks pass before submitting

### Contribution Process
1. Fork the repository and create a feature branch
2. Implement changes with appropriate tests
3. Run `npm run check` to validate code quality
4. Submit a pull request with detailed description

For detailed guidelines, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Documentation

### Complete Documentation Suite
- **[API Reference](docs/API.md)** - Complete REST API documentation
- **[Architecture Guide](docs/ARCHITECTURE.md)** - Technical architecture and design decisions  
- **[Configuration Guide](docs/CONFIGURATION.md)** - Environment setup and deployment options
- **[Development Guide](DEVELOPMENT.md)** - Development tools and workflows
- **[Performance Guide](docs/PERFORMANCE.md)** - Optimization and scaling strategies
- **[Troubleshooting Guide](docs/TROUBLESHOOTING.md)** - Common issues and solutions

### Additional Resources
- **[Project Roadmap](roadmap.md)** - Future development plans
- **[Changelog](CHANGELOG.md)** - Version history and release notes
- **[Contributing Guidelines](CONTRIBUTING.md)** - How to contribute to the project


## Support

For questions, feature requests, or technical support:
- Review the comprehensive [documentation](docs/)
- Check the [troubleshooting guide](docs/TROUBLESHOOTING.md)
- Open an issue on the repository
- Consult the [changelog](CHANGELOG.md) for recent updates


