# Integration Tests

This directory contains comprehensive integration tests for the KivaPlus backend.

## 📁 Test Files Overview

### Core Integration Tests
- **`real_flow_test.go`** - Complete end-to-end flow with real handlers
- **`user_condominium_flow_test.go`** - User-condominium workflow tests
- **`basic_integration_test.go`** - Basic Lambda handler tests
- **`http_integration_test.go`** - HTTP server integration tests

## 🧪 Test Categories

### 1. End-to-End Flow Tests (`TestRealUserCondominiumFlow`)
Tests the complete user journey:
1. **User Registration** - POST `/register`
2. **User Login** - POST `/login`
3. **Profile Completion** - PUT `/profile`
4. **Condominium Creation** - POST `/condominios`
5. **Condominium Retrieval** - GET `/condominios/:id`

### 2. Lambda Handler Tests (`TestBasicUserRegistration`)
Tests Lambda functions directly:
- Direct handler invocation
- AWS Lambda event simulation
- Database integration

### 3. HTTP Server Tests (`TestHTTPUserRegistrationAndLogin`)
Tests via HTTP server:
- Real HTTP requests
- Authentication flow
- Profile access

### 4. Validation Tests (`TestFlowWithValidation`)
Tests error handling:
- Invalid registration attempts
- Invalid login attempts
- Authorization checks

## 🚀 Running Tests

### Quick Commands
```bash
# Run all integration tests
make test

# Run specific test file
go test -v ./tests/integration -run TestRealUserCondominiumFlow

# Run with timeout
go test -v ./tests/integration -timeout 60s

# Run HTTP tests (requires local server)
make run-local  # In another terminal
go test -v ./tests/integration -run TestHTTP
```

### Detailed Setup
```bash
# 1. Setup test environment
make setup

# 2. Run all tests
make test

# 3. Run specific test categories
go test -v ./tests/integration -run TestReal      # End-to-end tests
go test -v ./tests/integration -run TestBasic     # Lambda tests
go test -v ./tests/integration -run TestHTTP      # HTTP tests
```

## 🔧 Prerequisites

### Required Services
1. **PostgreSQL** - Database (Docker recommended)
2. **Redis** - Cache (Docker recommended)
3. **LocalStack** - AWS services simulation

### Environment Setup
```bash
# Start all services
make setup

# Or manually
docker compose up -d postgres redis localstack
```

### Environment Variables
```bash
# Test database
export DATABASE_URL='postgres://kivaplus:kivaplus123@localhost:5432/kivaplus_local?sslmode=disable'

# JWT secret
export JWT_SECRET='test-secret-key-for-integration-tests'

# Test environment
export GO_ENV=test
```

## 📊 Test Data Management

### Unique Test Data
All tests create unique data to avoid conflicts:
- Users: `user+timestamp@test.com`
- Condominiums: `Test Condominium + timestamp`
- Documents: Generated with unique numbers

### Data Cleanup
```bash
# Clean test data (optional)
make integration-clean

# Or manually
docker compose exec postgres psql -U kivaplus -d kivaplus_local -c "
DELETE FROM condominiums WHERE name LIKE '%test%';
DELETE FROM users WHERE email LIKE '%test%';
"
```

## 🏗️ Test Architecture

### Test Structure
```
tests/integration/
├── README.md                    # This documentation
├── basic_integration_test.go    # Lambda handler tests
├── http_integration_test.go     # HTTP server tests
├── real_flow_test.go           # End-to-end flow tests
└── user_condominium_flow_test.go # User-condominium workflow
```

### Test Dependencies
```go
// Core testing
github.com/stretchr/testify/assert
github.com/stretchr/testify/require

// AWS Lambda
github.com/aws/aws-lambda-go/events

// Application
internal/shared/database
internal/users/handlers
internal/condominiums/handlers
```

## ✅ What Gets Tested

### API Endpoints
- ✅ POST `/register` - User registration
- ✅ POST `/login` - User authentication
- ✅ POST `/refresh` - Token refresh
- ✅ GET `/profile` - Profile retrieval
- ✅ PUT `/profile` - Profile completion
- ✅ GET `/condominios` - List condominiums
- ✅ POST `/condominios` - Create condominium
- ✅ GET `/condominios/:id` - Get condominium

### Business Logic
- ✅ User registration with validation
- ✅ JWT token generation and validation
- ✅ Profile completion workflow
- ✅ Permission-based access control
- ✅ Condominium creation with address
- ✅ Cross-entity relationships

### Technical Aspects
- ✅ Database persistence
- ✅ Transaction integrity
- ✅ Error handling and validation
- ✅ Authentication and authorization
- ✅ Lambda function execution
- ✅ HTTP request/response handling

## 🐛 Troubleshooting

### Common Issues

#### Database Connection Failed
```bash
# Check database status
make status

# Restart database
docker compose restart postgres

# Check connection manually
docker compose exec postgres pg_isready -U kivaplus -d kivaplus_local
```

#### Tests Timeout
```bash
# Increase timeout
go test -v ./tests/integration -timeout 120s

# Run specific test
go test -v ./tests/integration -run TestRealUserCondominiumFlow -timeout 60s
```

#### HTTP Tests Fail
```bash
# Start local server first
make run-local

# Then run HTTP tests
go test -v ./tests/integration -run TestHTTP
```

#### JWT Secret Issues
```bash
# Set JWT secret
export JWT_SECRET='test-secret-key-for-integration-tests'

# Or check LocalStack secrets
aws --endpoint-url=http://localhost:4566 secretsmanager get-secret-value --secret-id kivaplus/jwt-secret
```

### Debug Commands
```bash
# Verbose test output
go test -v ./tests/integration

# Test with race detection
go test -race ./tests/integration

# Test with coverage
go test -cover ./tests/integration

# Test specific function
go test -v ./tests/integration -run TestRealUserCondominiumFlow
```

## 🔄 CI/CD Integration

### GitHub Actions Example
```yaml
name: Integration Tests
on: [push, pull_request]

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.21'

      - name: Start services
        run: make setup

      - name: Run integration tests
        run: make test
        env:
          DATABASE_URL: postgres://kivaplus:kivaplus123@localhost:5432/kivaplus_local?sslmode=disable
          JWT_SECRET: test-secret-key-for-ci
```

### Docker-based Testing
```bash
# Run tests in Docker
docker compose --profile test run --rm integration-tests

# Build test image
docker compose build integration-tests
```

## 📈 Best Practices

### Test Design
1. **Isolation** - Each test creates unique data
2. **Cleanup** - Test data is identifiable and removable
3. **Realistic** - Use real handlers and use cases
4. **Fast** - Optimized for quick execution
5. **Reliable** - Handle timing and async operations

### Code Quality
1. **Assertions** - Use meaningful assertions
2. **Error Handling** - Proper error checking
3. **Documentation** - Clear test descriptions
4. **Maintainability** - Easy to update and extend

### Performance
1. **Parallel Execution** - Tests can run in parallel
2. **Resource Management** - Proper cleanup
3. **Timeout Handling** - Reasonable timeouts
4. **Database Optimization** - Efficient queries

## 📞 Support

- **Issues**: Report test failures as GitHub issues
- **Documentation**: This README and inline comments
- **Debugging**: Use verbose mode and debug commands
- **Help**: Check troubleshooting section first

---

**🎯 Goal**: Ensure KivaPlus backend works correctly in all scenarios through comprehensive integration testing.
