# Tests Directory

This directory contains all the Go tests for the project, organized separately from the source code for better project structure.

## Structure

```
test/
└── api/
    ├── *_test.go          # Test files (copied from src/api/)
    ├── *.go               # Symlinked source files
    ├── go.mod             # Module definition
    ├── go.sum             # Module dependencies
    ├── public/            # Symlinked config files
    ├── services/          # Symlinked services directory
    ├── errors/            # Symlinked errors directory
    └── datasets/          # Test data storage
```

## Running Tests

### Run All Tests
```bash
cd test/api
go test -v ./...
```

### Run Specific Test
```bash
cd test/api
go test -v -run TestFunctionName
```

### Run Tests with Coverage
```bash
cd test/api
go test -v -cover ./...
```

### Run Benchmarks
```bash
cd test/api
go test -v -bench=.
```

## How It Works

- **Test Files**: All `*_test.go` files are copied to this directory to keep them separate from source code
- **Source Files**: Symbolic links to source files in `src/api/` allow tests to access the code being tested
- **Dependencies**: The same `go.mod` and `go.sum` files ensure consistent dependency management
- **Configuration**: Symlinks to `public/`, `services/`, and `errors/` directories provide access to necessary configuration and helper files

## Benefits

1. **Separation of Concerns**: Tests are organized separately from production code
2. **Clean Source Directory**: The `src/api/` directory only contains production code
3. **Easy Testing**: All tests can be run from a single location
4. **Maintains Functionality**: Tests can still access unexported functions and types since they're in the same package

## Notes

- Tests remain in the `main` package to access unexported functions
- Symlinks ensure tests have access to all necessary source files and configurations
- The directory structure supports both unit tests and benchmarks
- Test data and temporary files are isolated in the test directory 