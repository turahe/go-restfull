# Testing Strategy and Setup

This document outlines the comprehensive testing strategy for the Go RESTful API project, including unit tests, integration tests, and end-to-end tests.

## Overview

The testing strategy follows a multi-layered approach to ensure code quality, reliability, and maintainability:

1. **Unit Tests** - Test individual functions and methods in isolation
2. **Integration Tests** - Test interactions between components
3. **End-to-End Tests** - Test complete user workflows
4. **Performance Tests** - Test system performance under load

## Test Structure

### Directory Organization

```
├── internal/
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── user_test.go
│   │   │   ├── role_test.go
│   │   │   └── ...
│   │   └── services/
│   │       ├── password_service_test.go
│   │       └── ...
│   ├── infrastructure/
│   │   └── adapters/
│   │       ├── user_repository_test.go
│   │       └── ...
│   ├── interfaces/
│   │   └── http/
│   │       └── controllers/
│   │           ├── user_controller_test.go
│   │           └── ...
│   └── testutils/
│       └── testutils.go
├── config/
│   └── config.testing.yaml
└── docs/
    └── TESTING_STRATEGY.md
```

## Test Categories

### 1. Unit Tests

Unit tests focus on testing individual functions, methods, and small components in isolation.

#### Domain Entities Tests
- **Purpose**: Test business logic and validation rules
- **Location**: `internal/domain/entities/*_test.go`
- **Examples**:
  - User entity validation
  - Role entity business rules
  - Post entity methods

#### Domain Services Tests
- **Purpose**: Test business logic services
- **Location**: `internal/domain/services/*_test.go`
- **Examples**:
  - Password service hashing and validation
  - Email service formatting
  - RBAC service authorization logic

#### Repository Tests
- **Purpose**: Test data access layer
- **Location**: `internal/infrastructure/adapters/*_test.go`
- **Examples**:
  - User repository CRUD operations
  - Post repository queries
  - Role repository assignments

### 2. Integration Tests

Integration tests verify that different components work together correctly.

#### Controller Tests
- **Purpose**: Test HTTP handlers and request/response processing
- **Location**: `internal/interfaces/http/controllers/*_test.go`
- **Examples**:
  - User registration and login
  - Post creation and retrieval
  - Role assignment and management

#### Service Integration Tests
- **Purpose**: Test service layer interactions
- **Location**: `internal/application/services/*_test.go`
- **Examples**:
  - User service with repository
  - Auth service with JWT
  - RBAC service with Casbin

### 3. End-to-End Tests

E2E tests verify complete user workflows from HTTP request to database persistence.

#### API Tests
- **Purpose**: Test complete API endpoints
- **Location**: `tests/e2e/*_test.go`
- **Examples**:
  - User registration → login → profile update
  - Post creation → publishing → commenting
  - Role assignment → permission checking

## Test Utilities

### TestSetup Structure

The `TestSetup` struct provides a centralized way to manage test dependencies:

```go
type TestSetup struct {
    DB          *pgxpool.Pool
    RedisClient redis.Cmdable
    Container   *container.Container
    Cleanup     func()
}
```

### Key Test Utilities

#### Database Setup
```go
func SetupTestDatabase(t *testing.T) *pgxpool.Pool
```
- Creates test database connection
- Runs migrations
- Provides cleanup functionality

#### Redis Setup
```go
func SetupTestRedis(t *testing.T) redis.Cmdable
```
- Creates test Redis connection
- Uses separate database for isolation
- Provides cleanup functionality

#### Container Setup
```go
func SetupTestContainer(t *testing.T) *TestSetup
```
- Initializes all dependencies
- Creates service container
- Provides comprehensive cleanup

#### Test Data Creation
```go
func CreateTestUser(t *testing.T, db *pgxpool.Pool, username, email, password string) string
func CreateTestRole(t *testing.T, db *pgxpool.Pool, name, slug, description string) string
func AssignUserRole(t *testing.T, db *pgxpool.Pool, userID, roleID string)
```

#### Response Assertions
```go
func AssertJSONResponse(t *testing.T, body []byte, expectedCode int, expectedMessage string)
```

## Test Configuration

### Testing Configuration File

The `config.testing.yaml` file provides test-specific settings:

```yaml
# Testing Configuration
env: "testing"
port: 8001
host: "localhost"

# Database Configuration for Testing
database:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "password"
  name: "webapi_test"
  sslmode: "disable"

# Redis Configuration for Testing
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 1

# JWT Configuration
jwt:
  secret: "test-secret-key-for-testing-only"
  expires_in: "24h"
```

### Environment Variables

Set these environment variables for testing:

```bash
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_NAME=webapi_test
export TEST_REDIS_HOST=localhost
export TEST_REDIS_PORT=6379
```

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Test Categories
```bash
# Unit tests only
go test ./internal/domain/...

# Integration tests only
go test ./internal/interfaces/...

# E2E tests only
go test ./tests/e2e/...
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Run Tests with Race Detection
```bash
go test -race ./...
```

## Test Best Practices

### 1. Test Naming Convention
- Use descriptive test names that explain the scenario
- Follow the pattern: `Test[FunctionName]_[Scenario]`
- Example: `TestUser_UpdateUser_ValidData`, `TestUser_UpdateUser_InvalidEmail`

### 2. Test Structure (AAA Pattern)
```go
func TestExample(t *testing.T) {
    // Arrange - Set up test data and dependencies
    user := CreateTestUser(t, db, "testuser", "test@example.com", "password123")
    
    // Act - Execute the function being tested
    result, err := service.UpdateUser(user.ID, "newemail@example.com")
    
    // Assert - Verify the results
    assert.NoError(t, err)
    assert.Equal(t, "newemail@example.com", result.Email)
}
```

### 3. Test Data Management
- Use test utilities to create consistent test data
- Clean up test data after each test
- Use unique identifiers to avoid conflicts
- Use transactions for database tests when possible

### 4. Mocking and Stubbing
- Mock external dependencies (databases, APIs)
- Use interfaces for testability
- Create test doubles for complex dependencies

### 5. Error Testing
- Test both success and failure scenarios
- Test edge cases and boundary conditions
- Test error messages and error codes

## Continuous Integration

### GitHub Actions Workflow

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: password
          POSTGRES_DB: webapi_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:6
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.24
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out
```

## Coverage Targets

### Minimum Coverage Requirements
- **Unit Tests**: 80% coverage
- **Integration Tests**: 70% coverage
- **Overall Coverage**: 75% coverage

### Coverage Exclusions
- Generated code
- Main function
- Configuration files
- Test files themselves

## Performance Testing

### Load Testing
- Use tools like Apache Bench (ab) or wrk
- Test API endpoints under load
- Monitor response times and error rates

### Benchmark Tests
```go
func BenchmarkUserService_CreateUser(b *testing.B) {
    setup := SetupTestContainer(b)
    defer setup.Cleanup()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        user, _ := setup.Container.UserService.CreateUser(context.Background(), "user"+strconv.Itoa(i), "user"+strconv.Itoa(i)+"@example.com", "password123")
        _ = user
    }
}
```

## Security Testing

### Authentication Tests
- Test JWT token validation
- Test password hashing and verification
- Test session management

### Authorization Tests
- Test RBAC permissions
- Test role-based access control
- Test API endpoint protection

### Input Validation Tests
- Test SQL injection prevention
- Test XSS prevention
- Test input sanitization

## Monitoring and Reporting

### Test Metrics
- Test execution time
- Coverage percentage
- Pass/fail rates
- Performance benchmarks

### Test Reports
- Generate HTML coverage reports
- Export test results to CI/CD
- Track test trends over time

## Troubleshooting

### Common Issues

#### Database Connection Issues
- Ensure PostgreSQL is running
- Check connection credentials
- Verify database exists

#### Redis Connection Issues
- Ensure Redis is running
- Check connection credentials
- Verify Redis database is accessible

#### Test Data Conflicts
- Use unique identifiers
- Clean up test data properly
- Use test transactions

#### Race Conditions
- Run tests with `-race` flag
- Use proper synchronization
- Avoid shared state in tests

## Conclusion

This comprehensive testing strategy ensures:

1. **Code Quality**: High test coverage and quality
2. **Reliability**: Robust error handling and edge case testing
3. **Maintainability**: Well-structured and documented tests
4. **Performance**: Performance testing and optimization
5. **Security**: Security-focused testing practices

By following this strategy, the application will be more reliable, maintainable, and secure. 