# Football Manager Data Browser (FM-Dash)

<div align="center">

![FM-Dash Logo](https://img.shields.io/badge/FM--Dash-Football%20Manager%20Data%20Browser-blue?style=for-the-badge)

A powerful tool for parsing, analyzing, and visualizing Football Manager player data with a modern Vue.js + Quasar UI interface and robust Go backend.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue.js-3.0+-4FC08D?style=flat-square&logo=vue.js)](https://vuejs.org)
[![Quasar Version](https://img.shields.io/badge/Quasar-2.0+-1976D2?style=flat-square&logo=quasar)](https://quasar.dev)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

</div>

## 🚀 Features

- **🔄 Data Processing**: Advanced HTML parser for Football Manager exports
- **📊 Interactive Analytics**: Rich data visualization and player statistics
- **🔍 Smart Search**: Advanced filtering and search capabilities
- **📱 Responsive Design**: Modern UI that works on all devices
- **☁️ Cloud Ready**: Kubernetes-native deployment with S3 storage
- **📈 Observability**: Built-in OpenTelemetry metrics and tracing
- **🎮 Engaging UX**: Game-inspired loading screens and interactions

## 🏗️ Architecture

### Frontend (Vue.js + Quasar)
- **Framework**: Vue 3 with Composition API
- **UI Library**: Quasar Framework for Material Design components
- **Build Tool**: Vite for fast development and optimized builds
- **Testing**: Vitest for unit testing
- **Code Quality**: Biome for linting and formatting

### Backend (Go)
- **Language**: Go 1.21+ with modern concurrency patterns
- **HTTP Framework**: Native Go net/http with custom middleware
- **File Processing**: Advanced HTML parsing with golang.org/x/net/html
- **Storage**: S3-compatible object storage (MinIO/AWS S3)
- **Observability**: OpenTelemetry integration with SignOz
- **Testing**: Native Go testing with coverage reporting

### Deployment
- **Container**: Multi-stage Docker builds with Nginx + supervisord
- **Orchestration**: Kubernetes with production-ready manifests
- **Monitoring**: Integrated OpenTelemetry and health checks
- **Security**: Non-root containers with proper secret management

## 🛠️ Quick Start

### Prerequisites

- **Go** 1.21 or higher
- **Node.js** 18 or higher
- **Docker** (for containerized deployment)
- **Kubernetes** (for production deployment)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://git.liamhardman.com/liam/v2fmdash.git
   cd v2fmdash
   ```

2. **Run the setup script**
   ```bash
   ./scripts/setup-dev.sh
   ```
   This will check dependencies and install packages for both frontend and backend.

3. **Start the development environment**
   
   **Option A: Development Mode (Hot Reload)**
   ```bash
   # Terminal 1: Start Go backend
   go run main.go
   
   # Terminal 2: Start Vue frontend
   npm run dev
   ```
   
   **Option B: Production-like Mode**
   ```bash
   # Build and run with Docker
   docker build -t fm-dash .
   docker run -p 8080:8080 fm-dash
   ```

4. **Access the application**
   - Development: http://localhost:3000 (frontend) + http://localhost:8091 (API)
   - Production: http://localhost:8080

## 📖 Usage Guide

### Basic Workflow

1. **Export Data**: Export player data from Football Manager as HTML
2. **Upload File**: Use the web interface to upload your FM HTML file
3. **Analyze Data**: Explore players with advanced search, filtering, and sorting
4. **Visualize Insights**: View statistics and data visualizations

### Advanced Features

- **Interactive Upload**: Engaging loading screens with progress tracking
- **Smart Search**: Search by player attributes, positions, clubs, and more
- **Data Export**: Export filtered results in various formats
- **Team Analysis**: Compare players and build optimal team compositions

## 📚 Documentation

### Core Documentation
- **[API Documentation](docs/API.md)** - Complete REST API reference with examples
- **[Architecture Guide](docs/ARCHITECTURE.md)** - Detailed technical architecture and design decisions
- **[Configuration Guide](docs/CONFIGURATION.md)** - Environment variables, deployment, and configuration options
- **[Development Guide](DEVELOPMENT.md)** - Development setup, tools, and workflows

### Operational Documentation
- **[Troubleshooting Guide](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[Performance Guide](docs/PERFORMANCE.md)** - Optimization, monitoring, and benchmarking
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute to the project

### Additional Resources
- **[Roadmap](roadmap.md)** - Future features and development plans
- **[License](LICENSE)** - MIT License details

## 🚢 Deployment

### Development
See [DEVELOPMENT.md](DEVELOPMENT.md) for detailed development setup and tools.

### Production (Kubernetes)

1. **Build and push image**
   ```bash
   docker build -t your-registry/fm-dash:latest .
   docker push your-registry/fm-dash:latest
   ```

2. **Configure secrets**
   ```bash
   kubectl create secret generic v2fmdash-minio-secret \
     --from-literal=endpoint=your-s3-endpoint \
     --from-literal=access-key=your-access-key \
     --from-literal=secret-key=your-secret-key \
     --from-literal=use-ssl=true
   ```

3. **Deploy to Kubernetes**
   ```bash
   kubectl apply -f kube.yaml
   ```

For detailed deployment configurations, see the [Configuration Guide](docs/CONFIGURATION.md).

## 🧪 Testing

### Run All Tests
```bash
npm run test:all        # Frontend + Backend tests
```

### Individual Test Suites
```bash
# Frontend tests
npm run test            # Watch mode
npm run test:run        # Single run
npm run test:coverage   # With coverage

# Backend tests  
npm run test:go         # Go tests
npm run test:go:coverage # With coverage
```

## 🔧 Development Tools

### Code Quality
```bash
npm run lint:all        # Check all linting
npm run format          # Auto-fix formatting
npm run check           # Full quality check
```

### Pre-commit Hooks
Automated quality checks run on every commit:
- **Pre-commit**: Fast checks on staged files only
- **Pre-push**: Full test suite and linting

## 📁 Project Structure

```
.
├── src/                    # Vue.js frontend source
│   ├── components/         # Reusable Vue components
│   ├── pages/             # Page components and routes
│   └── utils/             # Frontend utilities
├── *.go                   # Go backend source files
├── docs/                  # Comprehensive documentation
│   ├── API.md             # API documentation
│   ├── ARCHITECTURE.md    # Technical architecture
│   ├── CONFIGURATION.md   # Configuration guide
│   ├── TROUBLESHOOTING.md # Troubleshooting guide
│   └── PERFORMANCE.md     # Performance optimization
├── kube.yaml              # Kubernetes deployment manifests
├── Dockerfile             # Multi-stage container build
├── DEVELOPMENT.md         # Detailed development guide
├── INTERACTIVE_UPLOAD_FEATURE.md # Feature documentation
└── scripts/               # Development and deployment scripts
```

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes**: Follow the coding standards and run tests
4. **Run quality checks**: `npm run check`
5. **Commit changes**: `git commit -m 'Add amazing feature'`
6. **Push to branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### Development Guidelines

- Follow the existing code style and patterns
- Write tests for new features
- Update documentation for significant changes
- See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines

## 📊 Performance & Monitoring

FM-Dash is designed for high performance with large datasets:

- **Efficient Processing**: Handles 50MB+ HTML files with streaming
- **Memory Optimized**: Processes thousands of players efficiently
- **Fast Search**: Indexed search across large datasets
- **Responsive UI**: Virtual scrolling for large data tables
- **Monitoring**: Built-in metrics and health checks

For optimization tips, see the [Performance Guide](docs/PERFORMANCE.md).

## 🔧 Troubleshooting

Common issues and solutions:

- **Upload Issues**: Check file format and size limits
- **Performance**: Optimize worker count and memory settings
- **CORS Errors**: Verify allowed origins configuration
- **S3 Connection**: Check endpoint and credentials

For detailed troubleshooting, see the [Troubleshooting Guide](docs/TROUBLESHOOTING.md).

## 📈 Metrics & Health

Monitor your FM-Dash deployment:

```bash
# Health check
curl http://localhost:8091/api/health

# Metrics (Prometheus format)
curl http://localhost:8091/metrics

# Performance profiling
curl http://localhost:8091/debug/pprof/heap
```

## 🏆 Features Highlights

- **🚀 Fast Processing**: Stream-based HTML parsing for large files
- **💾 Smart Caching**: Multi-level caching for optimal performance
- **🔍 Advanced Search**: Complex filtering with real-time results
- **📊 Rich Analytics**: Player statistics and data visualizations
- **🎨 Modern UI**: Beautiful, responsive interface with dark mode
- **☁️ Cloud Native**: Kubernetes-ready with horizontal scaling
- **🔒 Secure**: Input validation, CORS protection, rate limiting
- **📱 Progressive**: Works offline with service worker caching

---

**Built with ❤️ for Football Manager enthusiasts**

For questions, issues, or feature requests, please check the [documentation](docs/) or open an issue on the repository.
