# Contributing to FM-Dash

Welcome to the FM-Dash community! We're excited to have you contribute to this sophisticated Football Manager data analysis platform. This guide covers everything you need to know about contributing to our advanced features and maintaining our high-quality standards.

## 🌟 Project Overview

FM-Dash is a comprehensive platform featuring:
- **Advanced Player Analysis Tools** (Bargain Hunter, Wonderkids Discovery, Upgrade Finder)
- **Sophisticated Data Processing** with Go-based HTML parsing and web workers
- **Modern Vue.js Architecture** with Pinia stores and Quasar UI components
- **Enterprise-Grade Infrastructure** with OpenTelemetry observability and S3 storage
- **Team Management Features** including wishlist system and team logo integration

## 🚀 Getting Started

### Prerequisites

**Core Requirements**:
- **Go 1.24+** - Backend API with modern concurrency patterns
- **Node.js 18+** - Frontend development with modern JavaScript features
- **Git** - Version control with conventional commit standards

**Optional Tools** (for advanced development):
- **Docker** - Container development and testing
- **MinIO** - Local S3-compatible storage for testing
- **VS Code** - Recommended IDE with extensions
- **Thunder Client** - API testing within VS Code

### Development Environment Setup

1. **Clone and Initialize**
   ```bash
   git clone <repository-url>
   cd fm-dash
   ./scripts/setup-dev.sh
   ```

2. **Configure Environment**
   ```bash
   # Copy environment template
   cp .env.example .env
   
   # Configure for local development
   echo "ENVIRONMENT=development" >> .env
   echo "LOG_LEVEL=debug" >> .env
   echo "ENABLE_TRACING=true" >> .env
   ```

3. **Start Development Servers**
   ```bash
   # Terminal 1: Backend API with hot reload
   npm run serve
   
   # Terminal 2: Frontend with Vite HMR
   npm run dev
   
   # Terminal 3: Optional - Storage services
   npm run dev:storage-local
   ```

4. **Verify Setup**
   ```bash
   # Health check
   curl http://localhost:8091/api/health
   
   # Frontend access
   open http://localhost:3000
   ```

## 🎯 Contribution Areas

### Frontend Contributions

#### Player Analysis Tools Development

**Bargain Hunter Dialog** (`src/components/BargainHunterDialog.vue`):
- Value-for-money calculation algorithms
- Budget-based filtering with configurable ranges
- Statistical modeling for transfer recommendations
- Export functionality for scouting reports

**Wonderkids Discovery** (`src/components/WonderkidsDialog.vue`):
- Age-based talent identification
- Potential vs. current ability analysis
- Growth trajectory predictions
- Scout rating integration

**Upgrade Finder** (`src/components/UpgradeFinderDialog.vue`):
- Position-specific improvement suggestions
- Squad analysis and gap identification
- Comparative statistics against current players
- Transfer budget optimization strategies

**Contributing Guidelines**:
```javascript
// Example: Adding new analysis algorithm
export function calculatePlayerValue(player, marketFactors) {
  // 1. Implement algorithm with proper type checking
  if (!player?.attributes || !marketFactors?.inflation) {
    throw new Error('Invalid input parameters')
  }
  
  // 2. Add comprehensive unit tests
  // 3. Document calculation methodology
  // 4. Consider performance implications for large datasets
  
  return {
    estimatedValue: baseValue * adjustmentFactor,
    confidence: confidenceScore,
    factors: explanationFactors
  }
}
```

#### Wishlist System Enhancement

**Current Features**:
- LocalStorage persistence with automatic synchronization
- Cross-tab state management using BroadcastChannel API
- Export/import functionality with JSON format
- Integration with player comparison tools

**Contribution Opportunities**:
- Cloud synchronization with user accounts
- Advanced wishlist categorization and tagging
- Sharing wishlists with other users
- Wishlist-based market alerts and notifications

#### Team Logo Integration

**Current Implementation**:
- Fuzzy search algorithms for automatic logo matching
- Dynamic loading with fallback mechanisms
- Logo optimization and WebP format support
- Custom logo upload and management

**Enhancement Areas**:
- Improved fuzzy matching accuracy
- Logo database expansion and crowdsourcing
- Performance optimization for large logo collections
- SVG logo support with dynamic theming

### Backend Contributions

#### HTML Parser Enhancement

**Core Parser** (`src/api/parsing.go`):
- Stream-based processing for memory efficiency
- Robust error handling for malformed HTML
- Support for multiple FM export formats
- Performance optimization for large datasets

**Contribution Guidelines**:
```go
// Example: Adding new parser functionality
func ParsePlayerAttributes(node *html.Node) (*PlayerAttributes, error) {
    // 1. Follow Go best practices and idioms
    if node == nil {
        return nil, errors.New("nil node provided")
    }
    
    // 2. Add comprehensive error handling
    attrs := &PlayerAttributes{}
    
    // 3. Implement OpenTelemetry tracing
    span := trace.SpanFromContext(ctx).StartSpan("parse_player_attributes")
    defer span.End()
    
    // 4. Add performance metrics
    start := time.Now()
    defer func() {
        metrics.RecordParsingDuration(time.Since(start))
    }()
    
    return attrs, nil
}
```

#### Observability and Monitoring

**OpenTelemetry Integration**:
- Distributed tracing across request lifecycle
- Custom metrics for business logic monitoring
- Structured logging with contextual information
- Performance profiling and bottleneck identification

**Enhancement Opportunities**:
- Custom dashboard development for SignOz
- Alert configuration for performance degradation
- Log aggregation and analysis automation
- Real-time performance monitoring

#### S3 Storage Optimization

**Current Features**:
- Intelligent caching with configurable TTL
- Automatic data compression and optimization
- Multi-region replication support
- Backup and recovery procedures

**Contribution Areas**:
- Advanced caching strategies (Redis integration)
- Data deduplication for efficient storage
- Encryption at rest and in transit
- Lifecycle management for old datasets

### Web Workers Development

**Player Calculation Worker** (`src/workers/playerCalculationWorker.js`):
- FIFA-style rating calculations
- Percentile analysis across large datasets
- Statistical modeling for player comparisons
- Background processing without UI blocking

**Development Standards**:
```javascript
// Example: Web Worker message handling
self.onmessage = function(e) {
  const { type, data, requestId } = e.data
  
  try {
    let result
    switch (type) {
      case 'CALCULATE_FIFA_RATINGS':
        result = calculateFIFARatings(data.players)
        break
      case 'ANALYZE_PERCENTILES':
        result = analyzePercentiles(data.players, data.league)
        break
      default:
        throw new Error(`Unknown calculation type: ${type}`)
    }
    
    // Post result back to main thread
    self.postMessage({
      requestId,
      success: true,
      result
    })
  } catch (error) {
    // Error handling with detailed context
    self.postMessage({
      requestId,
      success: false,
      error: error.message,
      stack: error.stack
    })
  }
}
```

## 📝 Contribution Process

### 1. Issue Creation and Planning

**Before Starting Work**:
- Search existing issues to avoid duplication
- Create detailed issue with use cases and acceptance criteria
- Discuss approach with maintainers for complex features
- Consider breaking large features into smaller, manageable PRs

**Issue Templates**:
- **Feature Request**: New functionality with detailed specifications
- **Bug Report**: Reproducible issues with environment details
- **Performance**: Performance optimization opportunities
- **Documentation**: Improvements to guides and API docs

### 2. Development Workflow

**Branch Strategy**:
```bash
# Feature development
git checkout -b feature/player-analysis-enhancement

# Bug fixes
git checkout -b fix/wishlist-persistence-bug

# Performance improvements
git checkout -b perf/optimize-large-dataset-parsing
```

**Development Best Practices**:
1. **Write Tests First**: Implement comprehensive test coverage
2. **Follow Code Standards**: Use automated linting and formatting
3. **Add Documentation**: Update relevant docs for user-facing changes
4. **Performance Consideration**: Benchmark critical paths
5. **Security Review**: Follow secure coding practices

### 3. Code Quality Standards

**Automated Quality Checks**:
```bash
# Full quality pipeline
npm run check                    # Lint + format + test + coverage

# Individual checks
npm run lint:all                 # Frontend (Biome) + Backend (golangci-lint)
npm run test:all                 # Complete test suite
npm run format                   # Auto-fix formatting issues
```

**Pre-commit Requirements**:
- All automated quality checks must pass
- Test coverage must be maintained above 80%
- No security vulnerabilities in dependencies
- Performance regression tests must pass

### 4. Testing Requirements

#### Frontend Testing Standards

**Component Testing**:
```javascript
// Example: Player analysis component test
import { mount } from '@vue/test-utils'
import { createTestingPinia } from '@pinia/testing'
import WonderkidsDialog from '@/components/WonderkidsDialog.vue'

describe('WonderkidsDialog', () => {
  let wrapper
  
  beforeEach(() => {
    wrapper = mount(WonderkidsDialog, {
      global: {
        plugins: [createTestingPinia({
          initialState: {
            players: {
              players: mockPlayersData
            }
          }
        })]
      }
    })
  })
  
  test('identifies wonderkids correctly', async () => {
    await wrapper.vm.analyzeWonderkids()
    
    const wonderkids = wrapper.vm.identifiedWonderkids
    expect(wonderkids).toHaveLength(3)
    expect(wonderkids[0].age).toBeLessThanOrEqual(21)
    expect(wonderkids[0].potential).toBeGreaterThanOrEqual(80)
  })
  
  test('filters by position correctly', async () => {
    await wrapper.setData({ selectedPosition: 'ST' })
    await wrapper.vm.analyzeWonderkids()
    
    const strikers = wrapper.vm.identifiedWonderkids
    expect(strikers.every(p => p.positions.includes('ST'))).toBe(true)
  })
})
```

**Store Testing**:
```javascript
// Example: Wishlist store testing
import { setActivePinia, createPinia } from 'pinia'
import { useWishlistStore } from '@/stores/wishlistStore'

describe('WishlistStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })
  
  test('adds player to wishlist', () => {
    const store = useWishlistStore()
    const player = { id: 1, name: 'Test Player' }
    
    store.addToWishlist(player)
    
    expect(store.wishlistPlayers).toHaveLength(1)
    expect(store.isInWishlist(player.id)).toBe(true)
  })
  
  test('persists wishlist to localStorage', () => {
    const store = useWishlistStore()
    const player = { id: 1, name: 'Test Player' }
    
    store.addToWishlist(player)
    
    const stored = JSON.parse(localStorage.getItem('fm-dash-wishlist'))
    expect(stored).toContain(player.id)
  })
})
```

#### Backend Testing Standards

**Unit Testing**:
```go
func TestPlayerValueCalculation(t *testing.T) {
    testCases := []struct {
        name          string
        player        Player
        marketFactors MarketFactors
        expected      float64
    }{
        {
            name: "High potential young player",
            player: Player{
                Age:       19,
                Overall:   75,
                Potential: 90,
                Position:  "ST",
            },
            marketFactors: MarketFactors{
                PositionInflation: 1.2,
                AgeBonus:         0.8,
            },
            expected: 25000000, // Expected value in euros
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := CalculatePlayerValue(tc.player, tc.marketFactors)
            assert.InDelta(t, tc.expected, result, 1000000) // 1M tolerance
        })
    }
}
```

**Integration Testing**:
```go
func TestHTMLParsingIntegration(t *testing.T) {
    // Test with real FM HTML export data
    testData := loadTestFile("testdata/full_export.html")
    
    result, err := ParseHTMLExport(testData)
    
    require.NoError(t, err)
    assert.Greater(t, len(result.Players), 1000)
    assert.Greater(t, len(result.Teams), 50)
    
    // Verify data quality
    for _, player := range result.Players {
        assert.NotEmpty(t, player.Name)
        assert.Greater(t, player.Age, 15)
        assert.Less(t, player.Age, 50)
    }
}
```

### 5. Documentation Requirements

**Code Documentation**:
- JSDoc comments for complex functions
- Go documentation following standard conventions
- API endpoint documentation with examples
- Architecture decision records (ADRs) for significant changes

**User Documentation**:
- Feature guides with screenshots
- API usage examples
- Configuration references
- Troubleshooting guides

## 🔍 Code Review Process

### Review Criteria

**Functionality**:
- Feature works as specified
- Edge cases are handled appropriately
- Error handling is comprehensive
- Performance is acceptable

**Code Quality**:
- Follows established patterns and conventions
- Is readable and well-documented
- Has appropriate test coverage
- Doesn't introduce security vulnerabilities

**Architecture**:
- Fits well with existing architecture
- Doesn't introduce unnecessary complexity
- Follows separation of concerns
- Is scalable and maintainable

### Review Guidelines for Contributors

**When Requesting Review**:
1. Provide clear description of changes
2. Include testing strategy and results
3. Document any breaking changes
4. Add screenshots for UI changes
5. Reference related issues and discussions

**Responding to Review Feedback**:
- Address all feedback promptly
- Ask for clarification when needed
- Update tests and documentation as required
- Request re-review after making changes

## 🚀 Release Process

### Version Management

We follow semantic versioning (semver):
- **Major** (x.0.0): Breaking changes
- **Minor** (0.x.0): New features, backward compatible
- **Patch** (0.0.x): Bug fixes, backward compatible

### Changelog Maintenance

**Automated Changelog**:
- Generated from conventional commits
- Categorized by type (feat, fix, perf, docs)
- Includes breaking change notifications
- Links to related issues and PRs

**Manual Updates**:
- Significant architectural changes
- Performance improvement summaries
- Migration guides for breaking changes
- Feature usage examples

## 🤝 Community Guidelines

### Communication Standards

- **Respectful**: Treat all community members with respect
- **Constructive**: Provide helpful feedback and suggestions
- **Inclusive**: Welcome contributors of all experience levels
- **Collaborative**: Work together to improve the project

### Getting Help

**Technical Support**:
- Check existing documentation and troubleshooting guides
- Search GitHub issues for similar problems
- Ask questions in GitHub Discussions
- Join community chat channels

**Contribution Support**:
- Start with good first issues labeled `good-first-issue`
- Ask for mentorship on complex features
- Participate in code review discussions
- Contribute to documentation improvements

## 📊 Performance Considerations

### Frontend Performance

**Optimization Strategies**:
- Virtual scrolling for large player datasets
- Lazy loading of images and logos
- Web Workers for CPU-intensive calculations
- Efficient state management with Pinia

**Measurement Tools**:
```bash
npm run lighthouse              # Performance audit
npm run bundle:analyze          # Bundle size analysis
npm run test:performance        # Performance regression tests
```

### Backend Performance

**Optimization Techniques**:
- Concurrent processing with goroutines
- Memory-efficient streaming parsers
- Intelligent caching strategies
- Database query optimization

**Profiling Tools**:
```bash
go tool pprof -http=:8080 http://localhost:8091/debug/pprof/profile
go test -bench=. -benchmem ./...
```

## 🔒 Security Guidelines

### Secure Coding Practices

**Input Validation**:
- Sanitize all user inputs
- Validate file uploads strictly
- Use parameterized queries
- Implement rate limiting

**Data Protection**:
- Encrypt sensitive data at rest
- Use HTTPS for all communications
- Implement proper session management
- Follow OWASP security guidelines

### Security Review Process

All contributions undergo security review:
- Automated dependency scanning
- Static analysis security testing
- Manual review of security-sensitive code
- Penetration testing for major releases

---

## 🎉 Recognition

We value all contributions to FM-Dash! Contributors are recognized through:
- Contributor credits in release notes
- GitHub contributor badges
- Community highlights
- Invitation to contributor meetings

Thank you for helping make FM-Dash the best Football Manager analysis platform! 🚀

---

*For questions about contributing, please check our [documentation](docs/) or open a discussion on GitHub.*