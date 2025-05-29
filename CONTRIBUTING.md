# Contributing to Football Manager Player Parser

Thank you for your interest in contributing to this project! We welcome contributions from the community.

## Getting Started

### Prerequisites

- **Go 1.21+** for the backend API
- **Node.js 18+** for the frontend development
- **Git** for version control

### Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd fm24-golang
   ```

2. **Install dependencies**
   ```bash
   # Backend dependencies
   go mod download
   
   # Frontend dependencies
   npm install
   ```

3. **Set up environment variables**
   ```bash
   # Copy and configure environment file
   cp .env.example .env
   # Edit .env with your local configuration
   ```

4. **Run the development servers**
   ```bash
   # Terminal 1: Start backend API
   go run src/api/main.go
   
   # Terminal 2: Start frontend dev server
   npm run dev
   ```

## How to Contribute

### Reporting Issues

- Use the GitHub issue tracker to report bugs
- Include detailed reproduction steps
- Provide system information (OS, Go version, Node.js version)
- Include relevant logs or error messages

### Submitting Changes

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Follow the existing code style
   - Add tests for new functionality
   - Update documentation as needed

4. **Test your changes**
   ```bash
   # Run Go tests
   go test ./...
   
   # Run frontend tests (if available)
   npm test
   
   # Build to ensure no build errors
   npm run build
   ```

5. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

6. **Push and create a pull request**
   ```bash
   git push origin feature/your-feature-name
   ```

### Code Style Guidelines

#### Go Code
- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Follow Go best practices and idioms

#### JavaScript/Vue Code
- Use consistent indentation (2 spaces)
- Follow Vue.js style guide
- Use meaningful component and variable names
- Add JSDoc comments for complex functions

#### General Guidelines
- Keep functions small and focused
- Write descriptive commit messages
- Include tests for new features
- Update documentation for user-facing changes

## Project Structure

```
fm24-golang/
├── src/
│   ├── api/           # Go backend API
│   ├── components/    # Vue.js components
│   ├── composables/   # Vue composition functions
│   ├── pages/         # Application pages
│   ├── services/      # Frontend services
│   ├── stores/        # Pinia state stores
│   └── utils/         # Utility functions
├── dist/              # Production build output
├── node_modules/      # Node.js dependencies
└── docs/              # Documentation
```

## Development Tips

### Backend Development
- Use structured logging with `slog`
- Follow the existing error handling patterns
- Add OpenTelemetry tracing for new endpoints
- Use environment variables for configuration

### Frontend Development
- Use Quasar components when available
- Leverage Vue Composition API
- Use Pinia for state management
- Implement proper error handling

### Testing
- Write unit tests for new functions
- Test edge cases and error conditions
- Ensure UI components render correctly
- Test API endpoints with various inputs

## Release Process

1. Update version numbers in relevant files
2. Update CHANGELOG.md with new features and fixes
3. Create a release tag
4. Build and test the release
5. Publish release notes

## Getting Help

- Check existing issues and documentation
- Ask questions in GitHub Discussions
- Join our community chat (if available)

## Code of Conduct

Please be respectful and constructive in all interactions. We aim to create a welcoming environment for contributors of all backgrounds and experience levels.

---

Thank you for contributing to Football Manager Player Parser!