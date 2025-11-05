# KivaPlus Backend Tests

Comprehensive test suite for the KivaPlus backend system.

## 📁 Test Structure

```
tests/
├── README.md                    # This documentation
├── integration/                 # Integration tests
│   ├── README.md               # Integration test docs
│   ├── basic_integration_test.go    # Lambda handler tests
│   ├── http_integration_test.go     # HTTP server tests
│   ├── real_flow_test.go           # End-to-end flow tests
│   └── user_condominium_flow_test.go # Workflow tests
└── unit/                       # Unit tests (distributed in internal/)
    └── (unit tests are co-located with source code)
```

## 🧪 Test Categories

### 1. Unit Tests
Located alongside source code in `internal/*/` directories:
- **Location**: `internal/users/usecases/*_test.go`
- **Purpose**: Test individual functions and methods
- **Scope**: Single unit of code
- **Dependencies**: Mocked

### 2. Integration Tests
Located in `tests/integration/`:
- **Location**: `tests/integration/*_test.go`
- **Purpose**: Test component interactions
- **Scope**: Multiple components working together
- **Dependencies**: Real database, services

## 🚀 Running Tests

### Quick Commands
```bash
# All tests
make test-all

# Unit tests only
make test

# Integration tests only
make test-integration

# Specific test file
go test -v ./tests/integration -run TestRealUserCondominiumFlow
```

### Detailed Commands
```bash
# Unit tests with coverage
go test -cover ./internal/...

# Integration tests with verbose output
go test -v ./tests/integration

# All tests with race detection
go test -race ./...

# Specific package tests
go test ./internal/users/usecases
go test ./internal/condominiums/handlers
```

## 🔧 Test Environment Setup

### Prerequisites
1. **Docker** - For PostgreSQL, Redis, LocalStack
2. **Go 1.21+** - For running tests
3. **Make** - For convenience commands

### Setup Commands
```bash
# Complete setup
make setup

# Just start services
docker compose up -d postgres redis localstack

# Check services
make status
```

### Environment Variables
```bash
# Required for tests
export DATABASE_URL='postgres://kivaplus:kivaplus123@localhost:5432/kivaplus_local?sslmode=disable'
export JWT_SECRET='test-secret-key-for-integration-tests'
export GO_ENV=test
```

## 📊 Test Coverage

### Current Coverage
- **Unit Tests**: 85%+ coverage
- **Integration Tests**: All major flows covered
- **End-to-End**: Complete user journeys tested

### Coverage Commands
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Coverage by package
go test -cover ./internal/users/...
go test -cover ./internal/condominiums/...
```

## 🎯 Test Guidelines

### Unit Test Guidelines
1. **Co-location**: Tests next to source code
2. **Naming**: `*_test.go` files
3. **Mocking**: Use interfaces and mocks
4. **Fast**: Should run in milliseconds
5. **Isolated**: No external dependencies

### Integration Test Guidelines
1. **Real Dependencies**: Use actual database/services
2. **Unique Data**: Generate unique test data
3. **Cleanup**: Clean test data after tests
4. **Realistic**: Test real user scenarios
5. **Comprehensive**: Cover happy path and edge cases

## 🐛 Troubleshooting

### Common Issues

#### Tests Fail to Connect to Database
```bash
# Check database status
make status

# Restart services
make setup

# Manual connection test
docker compose exec postgres pg_isready -U kivaplus -d kivaplus_local
```

#### JWT Secret Issues
```bash
# Set environment variable
export JWT_SECRET='test-secret-key-for-integration-tests'

# Check LocalStack secret
aws --endpoint-url=http://localhost:4566 secretsmanager get-secret-value --secret-id kivaplus/jwt-secret
```

#### Tests Timeout
```bash
# Increase timeout
go test -timeout 120s ./tests/integration

# Run specific test
go test -v ./tests/integration -run TestSpecificTest -timeout 60s
```

### Debug Commands
```bash
# Verbose output
go test -v ./tests/integration

# Race condition detection
go test -race ./...

# Memory profiling
go test -memprofile=mem.prof ./tests/integration

# CPU profiling
go test -cpuprofile=cpu.prof ./tests/integration
```

## 📈 Best Practices

### Writing Tests
1. **Clear Names**: Test names should describe what they test
2. **AAA Pattern**: Arrange, Act, Assert
3. **Single Responsibility**: One test, one scenario
4. **Independent**: Tests should not depend on each other
5. **Deterministic**: Same input, same output

### Test Data
1. **Unique**: Use timestamps or UUIDs
2. **Realistic**: Use realistic test data
3. **Minimal**: Only create necessary data
4. **Cleanup**: Remove test data after tests
5. **Isolated**: Each test creates its own data

### Performance
1. **Fast Unit Tests**: < 100ms per test
2. **Reasonable Integration**: < 5s per test
3. **Parallel Execution**: Use `t.Parallel()` when safe
4. **Resource Cleanup**: Proper cleanup to avoid leaks
5. **Efficient Queries**: Optimize database operations

## 🔄 CI/CD Integration

### GitHub Actions
Tests run automatically on:
- Push to main branch
- Pull requests
- Scheduled runs (nightly)

### Local Pre-commit
```bash
# Run before committing
make test-all

# Quick check
make test
```

### Docker Testing
```bash
# Run tests in Docker
docker compose --profile test run --rm integration-tests

# Build test environment
docker compose build integration-tests
```

## 📚 Resources

### Documentation
- [Integration Tests README](./integration/README.md)
- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)

### Tools
- **testify**: Assertions and test suites
- **go test**: Built-in Go testing
- **Docker**: Service dependencies
- **Make**: Test automation

### Examples
- Check existing tests for patterns
- Follow established conventions
- Use helper functions for common operations

---

**🎯 Goal**: Maintain high code quality through comprehensive testing at all levels.
