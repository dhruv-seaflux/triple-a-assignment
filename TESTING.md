# Testing Documentation

## Overview

This document provides comprehensive information about the testing strategy, test cases, and procedures for the Internal Transfers Application. The test suite ensures code quality, functionality, and reliability across all components.

## Testing Strategy

### Test Pyramid Structure

Our testing approach follows the test pyramid methodology:

1. **Unit Tests** (70%) - Fast, isolated tests for individual components
2. **Integration Tests** (20%) - Tests for component interactions
3. **End-to-End Tests** (10%) - Complete workflow validation

### Test Categories

#### 1. Unit Tests
- **Models & Validation**: Test data structures and input validation
- **Repository Layer**: Test database operations with mocked dependencies
- **Business Logic**: Test core application logic in isolation

#### 2. Integration Tests
- **API Endpoints**: Test HTTP handlers with real request/response cycles
- **Database Integration**: Test actual database operations
- **Component Integration**: Test interactions between application layers

#### 3. Performance Tests
- **Load Testing**: Test system behavior under high load
- **Concurrency Testing**: Test concurrent transaction processing
- **Benchmark Tests**: Measure performance metrics

## Test Structure

```
tests/
├── unit/                   # Unit tests
│   ├── models_test.go     # Model validation tests
│   └── repository_test.go # Repository layer tests
├── integration/           # Integration tests
│   └── api_test.go       # API endpoint tests
├── fixtures/              # Test data and fixtures
│   └── test_data.go      # Sample data for testing
├── testutils/             # Test utilities and helpers
│   └── utils.go          # Common test functions
└── test_config.go        # Test configuration
```

## Test Cases

### 1. Account Creation Tests

#### Valid Scenarios
- ✅ Create account with positive balance
- ✅ Create account with zero balance
- ✅ Create account with high precision decimal (5 decimal places)
- ✅ Create account with maximum account ID
- ✅ Handle concurrent account creation

#### Invalid Scenarios
- ❌ Negative account ID
- ❌ Zero account ID
- ❌ Negative initial balance
- ❌ Invalid balance format (non-numeric)
- ❌ Empty balance field
- ❌ Duplicate account ID

#### Expected Responses
```json
// Success (201 Created)
// Empty response body

// Error (400 Bad Request)
{
  "error": "account_id must be a positive integer"
}

// Error (409 Conflict)
{
  "error": "account already exists"
}
```

### 2. Account Query Tests

#### Valid Scenarios
- ✅ Query existing account
- ✅ Verify exact JSON response format
- ✅ Handle large account IDs
- ✅ Return correct decimal precision

#### Invalid Scenarios
- ❌ Non-existent account ID
- ❌ Invalid account ID format (non-numeric)
- ❌ Negative account ID

#### Expected Responses
```json
// Success (200 OK)
{
  "account_id": 123,
  "balance": "1000.50000"
}

// Error (404 Not Found)
{
  "error": "account not found"
}

// Error (400 Bad Request)
{
  "error": "account_id must be a positive integer"
}
```

### 3. Transaction Processing Tests

#### Valid Scenarios
- ✅ Transfer between different accounts
- ✅ High precision amount transfers
- ✅ Multiple sequential transactions
- ✅ Concurrent transaction processing
- ✅ Large amount transfers
- ✅ Minimum amount transfers (0.00001)

#### Invalid Scenarios
- ❌ Same account transfer (source = destination)
- ❌ Insufficient balance
- ❌ Negative amounts
- ❌ Zero amounts
- ❌ Non-existent source account
- ❌ Non-existent destination account
- ❌ Invalid amount format

#### Expected Responses
```json
// Success (201 Created)
{
  "transaction_id": 42,
  "status": "completed",
  "message": "Transaction processed successfully"
}

// Error (400 Bad Request)
{
  "error": "insufficient balance"
}

// Error (404 Not Found)
{
  "error": "one or both accounts not found"
}
```

### 4. Data Integrity Tests

#### Atomicity Tests
- ✅ Transaction rollback on failure
- ✅ Partial failure handling
- ✅ Database constraint enforcement

#### Consistency Tests
- ✅ Balance calculations
- ✅ Audit trail creation
- ✅ Concurrent access handling

#### Precision Tests
- ✅ Decimal arithmetic accuracy
- ✅ Rounding behavior
- ✅ High precision preservation

### 5. Error Handling Tests

#### Input Validation
- ✅ Malformed JSON requests
- ✅ Missing required fields
- ✅ Invalid data types
- ✅ Out-of-range values

#### System Errors
- ✅ Database connection failures
- ✅ Transaction timeout handling
- ✅ Memory limitations
- ✅ Network interruptions

## Running Tests

### Prerequisites

1. **Go Environment**: Go 1.21 or higher
2. **Test Database**: PostgreSQL instance for integration tests
3. **Dependencies**: All test dependencies installed

```bash
# Install test dependencies
go mod download
```

### Unit Tests

Unit tests use mocked dependencies and run quickly without external services.

```bash
# Run all unit tests
go test ./tests/unit/... -v

# Run specific test file
go test ./tests/unit/models_test.go -v

# Run with coverage
go test ./tests/unit/... -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests require a running PostgreSQL database.

```bash
# Set up test database environment
export TEST_DATABASE_URL="postgres://user:password@localhost:5432/internal_transfers_test?sslmode=disable"
export RUN_INTEGRATION_TESTS=true

# Run integration tests
go test ./tests/integration/... -v

# Run with database setup
go test ./tests/integration/... -v -tags=integration
```

### All Tests

```bash
# Run complete test suite
go test ./... -v

# Run tests with race detection
go test ./... -race -v

# Run tests with timeout
go test ./... -timeout=30s -v
```

### Test Environment Variables

```bash
# Database Configuration
TEST_DATABASE_URL="postgres://user:password@localhost:5432/internal_transfers_test?sslmode=disable"
TEST_DB_HOST="localhost"
TEST_DB_PORT="5432"
TEST_DB_USER="user"
TEST_DB_PASSWORD="password"
TEST_DB_NAME="internal_transfers_test"

# Test Control
RUN_INTEGRATION_TESTS="true"
TEST_ENV="unit|integration|e2e|performance"
```

## Performance Testing

### Load Test Scenarios

1. **High Volume Account Creation**
   - 1,000 accounts created concurrently
   - Measure response times and error rates

2. **Concurrent Transaction Processing**
   - 100 concurrent transactions
   - Verify data consistency and atomicity

3. **Stress Testing**
   - Gradually increase load until failure
   - Identify system limitations

### Benchmark Tests

```bash
# Run benchmark tests
go test -bench=. ./tests/... -benchmem

# Run specific benchmarks
go test -bench=BenchmarkAccountCreation ./tests/integration/... -v
go test -bench=BenchmarkTransactionProcessing ./tests/integration/... -v
```

### Performance Metrics

- **Response Time**: < 100ms for 95th percentile
- **Throughput**: > 1,000 transactions per second
- **Concurrency**: Support 100+ concurrent users
- **Error Rate**: < 0.1% under normal load

## Test Data Management

### Test Fixtures

Test data is managed through fixtures in `tests/fixtures/test_data.go`:

- **TestAccounts**: Sample account data
- **TestTransactions**: Sample transaction data
- **ValidRequests**: Valid API request examples
- **InvalidRequests**: Invalid API request examples
- **EdgeCases**: Boundary condition test data

### Database Setup

```sql
-- Create test database
CREATE DATABASE internal_transfers_test;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE internal_transfers_test TO user;
```

### Data Cleanup

Tests automatically clean up data after execution:
- Unit tests use mocked data (no cleanup needed)
- Integration tests clean database tables after each test
- Test database can be reset with cleanup scripts

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
        image: postgres:15
        env:
          POSTGRES_PASSWORD: password
          POSTGRES_DB: internal_transfers_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - name: Run tests
        run: |
          go test ./tests/unit/... -v
          go test ./tests/integration/... -v
        env:
          TEST_DATABASE_URL: postgres://postgres:password@localhost:5432/internal_transfers_test?sslmode=disable
```

### Pre-commit Hooks

```bash
#!/bin/sh
# Run tests before commit
go test ./tests/unit/... || exit 1
go fmt ./... || exit 1
go vet ./... || exit 1
```

## Test Coverage

### Coverage Targets

- **Overall Coverage**: > 85%
- **Unit Test Coverage**: > 90%
- **Integration Test Coverage**: > 70%
- **Critical Path Coverage**: 100%

### Coverage Report

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# View coverage by package
go tool cover -func=coverage.out
```

### Coverage Analysis

Critical areas requiring 100% coverage:
- Financial calculation logic
- Transaction processing
- Input validation
- Error handling paths

## Test Best Practices

### 1. Test Naming

```go
// Good: Descriptive test names
func TestAccountCreation_WithValidData_ShouldSucceed(t *testing.T)
func TestTransaction_WithInsufficientBalance_ShouldFail(t *testing.T)

// Bad: Generic test names
func TestAccount(t *testing.T)
func TestTransfer(t *testing.T)
```

### 2. Test Structure

Follow the AAA pattern:
- **Arrange**: Set up test data and dependencies
- **Act**: Execute the code under test
- **Assert**: Verify the results

```go
func TestAccountCreation(t *testing.T) {
    // Arrange
    repo := setupTestRepository(t)
    accountID := 123
    balance := decimal.NewFromFloat(1000.00)
    
    // Act
    account, err := repo.CreateAccount(accountID, balance)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, accountID, account.AccountID)
    assert.True(t, balance.Equal(account.Balance))
}
```

### 3. Test Data

- Use fixtures for consistent test data
- Generate random data for property-based testing
- Test boundary conditions and edge cases

### 4. Error Testing

```go
// Test specific error types
assert.ErrorIs(t, err, repository.ErrAccountNotFound)

// Test error messages
assert.Contains(t, err.Error(), "insufficient balance")
```

### 5. Cleanup

```go
func TestWithCleanup(t *testing.T) {
    db := setupTestDatabase(t)
    defer cleanupTestDatabase(t, db)
    
    // Test implementation
}
```

## Troubleshooting

### Common Issues

1. **Database Connection Failures**
   - Verify PostgreSQL is running
   - Check connection string format
   - Ensure test database exists

2. **Test Flakiness**
   - Use deterministic test data
   - Clean up properly between tests
   - Avoid time-dependent assertions

3. **Mock Failures**
   - Verify mock expectations
   - Check argument matchers
   - Ensure proper cleanup

### Debug Commands

```bash
# Run single test with verbose output
go test -run TestSpecificTest ./tests/unit/... -v

# Run tests with race detection
go test ./... -race

# Profile test execution
go test ./... -cpuprofile=cpu.prof -memprofile=mem.prof
```

## Conclusion

This comprehensive test suite ensures the reliability, correctness, and performance of the Internal Transfers Application. Regular execution of these tests helps maintain code quality and catch regressions early in the development process.

The test documentation should be updated as new features are added and test cases are expanded. All team members are encouraged to contribute to the test suite and maintain high testing standards.