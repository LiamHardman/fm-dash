# FM-Dash Development Guide

This comprehensive guide covers all aspects of developing, testing, and debugging FM-Dash, including advanced features and specialized development workflows.

## Quick Setup

Initialize your development environment with our automated setup:

```bash
./scripts/setup-dev.sh
```

This script validates dependencies, installs packages, and configures development tools for both frontend and backend components.

## Development Architecture Overview

### Frontend Stack (Vue.js Ecosystem)
- **Vue 3** with Composition API and `<script setup>` syntax
- **Quasar Framework** - Material Design component library with responsive utilities
- **Pinia Stores** - Modular state management (`playerStore`, `uiStore`, `wishlistStore`)
- **Vite** - Lightning-fast build tool with hot module replacement
- **Vitest** - Modern testing framework with component testing utilities
- **Biome** - Unified linting, formatting, and import organization
- **Web Workers** - Background processing for intensive calculations

### Backend Infrastructure (Go)
- **Go 1.24+** with modern concurrency patterns and generics
- **Native HTTP Stack** with custom middleware pipeline
- **OpenTelemetry** - Full observability with metrics, tracing, and logging
- **S3-Compatible Storage** - MinIO/AWS S3 with intelligent caching
- **golangci-lint** - Comprehensive static analysis with 40+ linters
- **Advanced Parser** - Stream-based HTML processing for large datasets

## Core Development Commands

### Quality Assurance Suite

```bash
# Comprehensive quality check
npm run check                    # Full pipeline: lint + format + test + coverage

# Individual quality checks
npm run lint:all                 # Frontend (Biome) + Backend (golangci-lint)
npm run lint:check               # Check-only mode (no auto-fixes)
npm run format                   # Auto-fix formatting issues
npm run format:check             # Verify formatting compliance

# Advanced linting
npm run lint:go:fix              # Auto-fix Go issues where possible
npm run lint:go                  # Go static analysis only
```

### Testing Infrastructure

```bash
# Frontend Testing Suite
npm run test                     # Watch mode with hot reloading
npm run test:run                 # Single run with exit
npm run test:ui                  # Interactive test UI dashboard
npm run test:coverage            # Coverage report with lcov output

# Backend Testing Suite
npm run test:go                  # Go unit and integration tests
npm run test:go:verbose          # Detailed test output with benchmarks
npm run test:go:coverage         # Coverage analysis with HTML report
npm run test:go:timeout          # Time-limited test execution

# Cross-Platform Testing
npm run test:all                 # Complete test suite (frontend + backend)
```

### Development Servers

```bash
# Development Mode (Hot Reload)
npm run dev                      # Vite dev server (port 3000)
npm run serve                    # Go API server (port 8091)

# Production Simulation
npm run build                    # Production build with optimization
npm run preview                  # Preview production build locally

# Advanced Development
npm run dev:debug                # Development with source maps
npm run dev:profile              # Performance profiling mode
```

## Advanced Development Features

### Specialized Component Development

**Player Analysis Tools Development**:
```bash
# Component-specific testing
npm run test:components          # Component isolation testing
npm run test:player-analysis     # Specialized analysis tool tests

# Development with mock data
npm run dev:mock                 # Use local test datasets
npm run dev:demo                 # Demo mode with sample data
```

**Wishlist System Development**:
- Persistent storage using LocalStorage API
- State synchronization across browser tabs
- Export/import functionality for user data migration
- Integration with player comparison tools

**Team Logo Integration**:
- Fuzzy search algorithms for logo matching
- Dynamic logo loading with fallback mechanisms
- Logo optimization and caching strategies
- Custom logo upload and management

### Web Workers Development

FM-Dash uses Web Workers for CPU-intensive calculations:

```javascript
// src/workers/playerCalculationWorker.js
// Handles FIFA rating calculations, percentile analysis, and statistical modeling
```

**Worker Development Workflow**:
```bash
# Worker-specific testing
npm run test:workers             # Web Worker unit tests
npm run dev:worker-debug         # Worker debugging mode

# Performance testing
npm run benchmark:workers        # Worker performance benchmarks
```

### State Management Development

**Pinia Store Architecture**:
- **playerStore.js** - Player data, filtering, and search state
- **uiStore.js** - UI state, modals, loading states, and user preferences
- **wishlistStore.js** - User wishlist management and persistence

**Store Development Best Practices**:
```javascript
// Example: Advanced store pattern with computed properties
export const usePlayerStore = defineStore('players', () => {
  // Reactive state
  const players = ref([])
  const filters = ref({})
  const searchTerm = ref('')

  // Advanced computed properties
  const wonderkids = computed(() => 
    players.value.filter(player => 
      player.age <= 21 && player.potential >= 80
    )
  )

  const bargainPlayers = computed(() =>
    players.value.filter(player =>
      calculateValueRatio(player) > 1.5
    )
  )

  // Async actions with error handling
  const fetchPlayers = async (params) => {
    try {
      loading.value = true
      const response = await playerService.getPlayers(params)
      players.value = response.data.players
    } catch (error) {
      handleApiError(error)
    } finally {
      loading.value = false
    }
  }

  return {
    // State
    players: readonly(players),
    filters: readonly(filters),
    
    // Computed
    wonderkids,
    bargainPlayers,
    
    // Actions
    fetchPlayers,
    updateFilters,
    searchPlayers
  }
})
```

## Backend Development Deep Dive

### Advanced Go Development

**OpenTelemetry Integration**:
```bash
# Observability development
go run -tags otel main.go        # Run with full telemetry
go run -tags debug main.go       # Debug mode with verbose logging

# Metrics and tracing
curl http://localhost:8091/metrics          # Prometheus metrics
curl http://localhost:8091/debug/pprof/heap # Memory profiling
```

**Performance Optimization**:
```go
// Example: Optimized concurrent processing
func ProcessPlayersInBatches(players []RawPlayer) []ProcessedPlayer {
    const batchSize = 1000
    workers := runtime.NumCPU()
    
    jobs := make(chan []RawPlayer, workers)
    results := make(chan []ProcessedPlayer, workers)
    
    // Worker pool pattern for CPU-intensive operations
    for i := 0; i < workers; i++ {
        go playerWorker(jobs, results)
    }
    
    // Distribute work in batches
    go func() {
        defer close(jobs)
        for i := 0; i < len(players); i += batchSize {
            end := min(i+batchSize, len(players))
            jobs <- players[i:end]
        }
    }()
    
    return collectResults(results, len(players))
}
```

### Database and Storage Development

**S3-Compatible Storage Integration**:
- Intelligent caching with TTL management
- Automatic data compression and optimization
- Multi-region replication support
- Backup and recovery procedures

**Development Storage Setup**:
```bash
# Local MinIO setup for development
docker run -p 9000:9000 -p 9001:9001 minio/minio server /data --console-address ":9001"

# Storage testing commands
npm run test:storage             # Storage integration tests
npm run dev:storage-local        # Use local MinIO instance
```

## IDE Integration and Tooling

### VS Code Configuration

**Essential Extensions**:
- **Biome** (biomejs.biome) - Unified linting and formatting
- **Vue Language Features (Volar)** (Vue.volar) - Vue 3 support
- **Go** (golang.go) - Go language support with debugging
- **Thunder Client** - API testing directly in VS Code
- **GitLens** - Advanced Git integration
- **Error Lens** - Inline error visualization

**Workspace Settings** (.vscode/settings.json):
```json
{
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.biome": true,
    "source.organizeImports.biome": true
  },
  "go.lintTool": "golangci-lint",
  "go.testFlags": ["-v", "-race"],
  "vue.server.hybridMode": true
}
```

### Advanced Debugging

**Frontend Debugging**:
```bash
# Debug modes
npm run dev:debug                # Source maps + Vue devtools
npm run test:debug               # Debug failing tests
npm run build:analyze            # Bundle analysis

# Performance debugging
npm run dev:profile              # Performance profiling
npm run lighthouse              # Lighthouse audit
```

**Backend Debugging**:
```bash
# Debug builds
go build -gcflags="-N -l" main.go       # Debug symbols
go run -race main.go                    # Race condition detection

# Profiling
go tool pprof http://localhost:8091/debug/pprof/profile
go tool pprof http://localhost:8091/debug/pprof/heap
```

## Advanced Testing Strategies

### Component Testing

**Player Analysis Component Testing**:
```javascript
// Example: Testing specialized dialogs
import { mount } from '@vue/test-utils'
import { createTestingPinia } from '@pinia/testing'
import BargainHunterDialog from '@/components/BargainHunterDialog.vue'

describe('BargainHunterDialog', () => {
  test('calculates value ratios correctly', async () => {
    const wrapper = mount(BargainHunterDialog, {
      global: {
        plugins: [createTestingPinia()]
      }
    })

    const mockPlayers = [
      { name: 'Test Player', transferValue: 1000000, overall: 85 }
    ]

    await wrapper.vm.analyzeValueRatios(mockPlayers)
    expect(wrapper.vm.bargainPlayers).toHaveLength(1)
  })
})
```

### Backend Testing

**Integration Testing with Real Data**:
```go
func TestHTMLParsingWithRealData(t *testing.T) {
    // Test with actual FM HTML exports
    testCases := []struct {
        filename string
        expectedPlayers int
        expectedLeagues int
    }{
        {"testdata/premier_league.html", 500, 1},
        {"testdata/world_data.html", 10000, 50},
    }

    for _, tc := range testCases {
        t.Run(tc.filename, func(t *testing.T) {
            data := loadTestFile(tc.filename)
            result, err := ParseHTML(data)
            
            assert.NoError(t, err)
            assert.Len(t, result.Players, tc.expectedPlayers)
            assert.Len(t, result.Leagues, tc.expectedLeagues)
        })
    }
}
```

## Performance Optimization Development

### Frontend Performance

**Bundle Optimization**:
```bash
# Bundle analysis
npm run build:analyze            # Webpack bundle analyzer
npm run lighthouse              # Performance audit

# Optimization commands
npm run optimize:images          # Image optimization
npm run optimize:fonts           # Font optimization
```

**Runtime Performance**:
- Virtual scrolling for large player datasets
- Lazy loading of player images and team logos
- Web Workers for intensive calculations
- Service Worker caching strategies

### Backend Performance

**Go Performance Optimization**:
```bash
# Benchmarking
go test -bench=. -benchmem       # Performance benchmarks
go test -race ./...              # Race condition detection

# Profiling
go tool pprof cpu.prof           # CPU profiling
go tool pprof mem.prof           # Memory profiling
```

## CI/CD Integration

### Automated Quality Gates

**Pre-commit Hooks** (Husky + lint-staged):
- **Pre-commit**: Fast, staged-file-only checks
  - Biome formatting and linting for frontend files
  - golangci-lint for Go files
  - Optimized for speed with parallel processing

- **Pre-push**: Comprehensive validation
  - Full test suite execution
  - Code coverage requirements
  - Build verification

**Pipeline Configuration**:
```yaml
# .github/workflows/quality.yml
name: Code Quality
on: [push, pull_request]

jobs:
  frontend:
    runs-on: ubuntu-latest
    steps:
      - name: Check code quality
        run: npm run check
      
      - name: Test components
        run: npm run test:coverage
        
  backend:
    runs-on: ubuntu-latest
    steps:
      - name: Run Go tests
        run: go test -race -cover ./...
        
      - name: Security scan
        run: gosec ./...
```

### Manual Testing and QA

**Manual Testing Workflows**:
```bash
# UI testing
npm run test:e2e                 # End-to-end testing
npm run test:visual              # Visual regression testing

# API testing
npm run test:api                 # API endpoint testing
npm run test:load               # Load testing
```

## Troubleshooting Development Issues

### Common Development Problems

**Frontend Issues**:
1. **Biome conflicts after dependency updates**:
   ```bash
   rm -rf node_modules package-lock.json
   npm install
   npm run lint:check
   ```

2. **Vue composition issues**:
   ```bash
   npm run dev:debug               # Enable Vue devtools
   npm run test:components         # Isolated component testing
   ```

3. **State management debugging**:
   ```bash
   # Enable Pinia devtools in development
   npm run dev:pinia-debug
   ```

**Backend Issues**:
1. **Go module issues**:
   ```bash
   go mod tidy                     # Clean dependencies
   go mod download                 # Re-download modules
   ```

2. **Memory issues during large file processing**:
   ```bash
   GOMEMLIMIT=4GiB go run main.go  # Set memory limit
   go tool pprof mem.prof          # Memory profiling
   ```

3. **OpenTelemetry configuration**:
   ```bash
   # Disable telemetry for debugging
   OTEL_SDK_DISABLED=true go run main.go
   ```

### Performance Debugging

**Frontend Performance Issues**:
```bash
# Component performance
npm run dev:profile              # Performance profiling
npm run test:performance         # Performance regression tests

# Bundle size debugging
npm run analyze:bundle           # Bundle composition analysis
npm run analyze:deps             # Dependency analysis
```

**Backend Performance Issues**:
```bash
# CPU profiling
go tool pprof -http=:8080 http://localhost:8091/debug/pprof/profile

# Memory profiling
go tool pprof -http=:8080 http://localhost:8091/debug/pprof/heap

# Goroutine analysis
go tool pprof -http=:8080 http://localhost:8091/debug/pprof/goroutine
```

## Advanced Development Workflows

### Feature Development Workflow

1. **Branch Creation**:
   ```bash
   git checkout -b feature/player-analysis-enhancement
   ```

2. **Development with Testing**:
   ```bash
   npm run test                    # Watch mode during development
   npm run dev                     # Hot reload development
   ```

3. **Quality Validation**:
   ```bash
   npm run check                   # Full quality pipeline
   npm run test:all                # Complete test suite
   ```

4. **Performance Validation**:
   ```bash
   npm run benchmark               # Performance benchmarks
   npm run lighthouse              # Performance audit
   ```

### Contributing to Core Features

**Player Analysis Tools**:
- Implement statistical algorithms in web workers
- Create reusable calculation functions
- Add comprehensive unit tests for edge cases
- Document calculation methodologies

**UI Component Development**:
- Follow Quasar design system principles
- Implement responsive design patterns
- Add accessibility features (ARIA labels, keyboard navigation)
- Create comprehensive component documentation

**Backend Feature Development**:
- Follow Go best practices and idioms
- Implement comprehensive error handling
- Add OpenTelemetry instrumentation
- Write integration tests with real data

## Development Resources

### Documentation Development

**Local Documentation Server**:
```bash
npm run docs:dev                 # Live documentation server
npm run docs:build               # Build static documentation
```

### Community and Support

- **Code Reviews**: All changes require peer review
- **Documentation**: Update docs for user-facing changes
- **Testing**: Maintain test coverage above 80%
- **Performance**: Benchmark critical paths
- **Security**: Follow secure coding practices

---

*This development guide is continuously updated. For the latest information, check the [changelog](CHANGELOG.md) and [project roadmap](roadmap.md).* 