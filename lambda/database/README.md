# Database Package

This package provides database clients for the KivaPlus backend, supporting both PostgreSQL and DynamoDB through a unified interface.

## Overview

The package implements the `UserStore` interface that provides user management operations:

```go
type UserStore interface {
    DoesUserExist(username string) (bool, error)
    InsertUser(user types.User) error
    GetUser(username string) (types.User, error)
}
```

## Clients

### PostgreSQL Client (`postgres.go`)

The PostgreSQL client connects to an RDS PostgreSQL instance and uses the `usuario` table for user management.

**Features:**
- Full CRUD operations for users
- Soft delete support (using `deleted_at` field)
- User activation/deactivation
- Last access tracking
- Connection pooling and error handling

**Database Schema:**
- Uses `email` field as username
- Stores password hash in `senha_hash` field
- Supports soft deletes with `deleted_at` timestamp
- Tracks user activity with `ultimo_acesso` field

### DynamoDB Client (`database.go`)

The DynamoDB client provides the same interface but stores data in AWS DynamoDB.

**Features:**
- Simple key-value storage
- AWS SDK integration
- Automatic configuration from environment

## Factory Function

The `NewUserStore()` function automatically selects the appropriate client based on environment variables:

```go
func NewUserStore() (UserStore, error)
```

- If `DATABASE_URL` is set → Returns PostgreSQL client
- Otherwise → Returns DynamoDB client

## Usage

### Basic Usage

```go
// Create a client (auto-selects based on environment)
store, err := database.NewUserStore()
if err != nil {
    log.Fatal(err)
}

// Check if user exists
exists, err := store.DoesUserExist("user@example.com")

// Create and insert user
registerUser := types.RegisterUser{
    Username: "user@example.com",
    Password: "password123",
}
user, err := types.NewUser(registerUser)
if err != nil {
    log.Fatal(err)
}

err = store.InsertUser(user)
if err != nil {
    log.Fatal(err)
}

// Get user
user, err = store.GetUser("user@example.com")
if err != nil {
    log.Fatal(err)
}

// Validate password
isValid := types.ValidatePassword(user.PasswordHash, "password123")
```

### PostgreSQL-Specific Operations

```go
// Create PostgreSQL client directly
client, err := database.NewPostgresClient()
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Update last access
err = client.UpdateLastAccess("user@example.com")

// Deactivate user
err = client.DeactivateUser("user@example.com")

// Soft delete user
err = client.DeleteUser("user@example.com")
```

## Environment Variables

### PostgreSQL
- `DATABASE_URL`: Full PostgreSQL connection string
  - Format: `postgresql://user:password@host:port/database?sslmode=require`

### DynamoDB
- `USER_TABLE_NAME`: DynamoDB table name for users
- AWS credentials (via AWS SDK configuration)

## Testing

### Running Tests

```bash
# Run all PostgreSQL tests
make test-postgres

# Run specific test
make test-postgres-specific

# Or use go test directly
source ./scripts/get-db-credentials.sh
go test -v ./lambda/database -run TestPostgres
```

### Test Requirements

- PostgreSQL database connection (via `DATABASE_URL`)
- SSH tunnel or direct database access
- Test database with proper schema (run migrations first)

### Test Coverage

The test suite covers:
- ✅ User creation and retrieval
- ✅ Password validation
- ✅ User existence checks
- ✅ Last access updates
- ✅ User deactivation
- ✅ Soft deletion
- ✅ Error handling
- ✅ Connection management
- ✅ Performance benchmarks

## Performance

Benchmark results on Apple M1 Max:
- `DoesUserExist`: ~289ms per operation
- `GetUser`: ~291ms per operation
- `UpdateLastAccess`: ~290ms per operation

*Note: Times include network latency to RDS via SSH tunnel*

## Error Handling

All methods return descriptive errors:
- Connection errors
- User not found errors
- Validation errors
- Database constraint violations

## Security

- Passwords are hashed using bcrypt
- SQL injection protection via prepared statements
- Connection string validation
- Soft deletes preserve audit trail

## Migration Support

The PostgreSQL client works with the migration system in `internal/database/migrations.go`. Ensure migrations are applied before using the client:

```bash
source ./scripts/get-db-credentials.sh
./scripts/run-migrations.sh up
```
