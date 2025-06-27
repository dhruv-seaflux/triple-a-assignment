# Internal Transfers API Specification

## Overview

The Internal Transfers API is a robust HTTP-based service for managing financial transfers between accounts. Built with Go and PostgreSQL, it provides secure, atomic operations with high-precision decimal arithmetic for financial accuracy.

### Key Features
- **Account Management**: Create and query accounts with integer-based IDs
- **Financial Transactions**: Atomic money transfers between accounts
- **Audit Trail**: Complete transaction logging for compliance
- **Data Integrity**: ACID-compliant operations with automatic rollback
- **Decimal Precision**: High-precision arithmetic (15 digits, 5 decimal places)

### API Characteristics
- **Protocol**: HTTP/1.1, REST-compliant
- **Data Format**: JSON request/response bodies
- **Authentication**: None (internal service)
- **Base URL**: `http://localhost:8080` (development)
- **Content-Type**: `application/json`

## Base Information

| Attribute | Value |
|-----------|-------|
| **API Version** | 1.0.0 |
| **Protocol** | HTTP/1.1 |
| **Data Format** | JSON |
| **Character Encoding** | UTF-8 |
| **Date Format** | ISO 8601 (YYYY-MM-DDTHH:mm:ss.sssZ) |
| **Decimal Format** | String representation for precision |

## Business Rules

### Account Management
1. **Account IDs**: Must be positive integers (1, 2, 3, ...)
2. **Uniqueness**: Each account ID must be unique across the system
3. **Initial Balance**: Cannot be negative, supports up to 5 decimal places
4. **Immutability**: Accounts cannot be deleted once created
5. **Balance Precision**: Stored as DECIMAL(15,5) in database

### Transaction Processing
1. **Amount Validation**: Must be positive (> 0)
2. **Account Verification**: Both source and destination accounts must exist
3. **Self-Transfer Prevention**: Source and destination must be different
4. **Balance Verification**: Source account must have sufficient funds
5. **Atomicity**: All operations succeed together or fail together
6. **Audit Logging**: Every transaction creates a permanent audit record

### Data Precision
1. **Decimal Arithmetic**: No floating-point calculations
2. **String Format**: All monetary amounts as strings to preserve precision
3. **Precision Limit**: Up to 5 decimal places supported
4. **Range Limit**: Maximum 15 total digits (10 integer + 5 decimal)

## API Endpoints

### 1. Create Account

Creates a new account with specified ID and initial balance.

```http
POST /accounts
Content-Type: application/json
```

**Request Body:**
```json
{
  "account_id": 123,
  "initial_balance": "1000.50000"
}
```

**Request Schema:**
| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `account_id` | integer | Yes | > 0, unique | Positive integer account identifier |
| `initial_balance` | string | Yes | ≥ 0, decimal | Initial balance as decimal string |

**Response Codes:**
- `201 Created` - Account created successfully (empty response)
- `400 Bad Request` - Invalid input data
- `409 Conflict` - Account ID already exists
- `500 Internal Server Error` - System error

**Example Requests:**

*Standard Account:*
```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 123, "initial_balance": "1000.50"}'
```

*Zero Balance Account:*
```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 456, "initial_balance": "0.00"}'
```

**Error Examples:**
```json
// Invalid account ID
{"error": "account_id must be a positive integer"}

// Negative balance
{"error": "initial_balance cannot be negative"}

// Duplicate account
{"error": "account already exists"}
```

### 2. Query Account Balance

Retrieves current balance for specified account.

```http
GET /accounts/{account_id}
```

**Path Parameters:**
| Parameter | Type | Required | Validation | Description |
|-----------|------|----------|------------|-------------|
| `account_id` | integer | Yes | > 0 | Account identifier to query |

**Response Schema:**
```json
{
  "account_id": 123,
  "balance": "1000.50000"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `account_id` | integer | The queried account identifier |
| `balance` | string | Current balance as decimal string |

**Response Codes:**
- `200 OK` - Account found and returned
- `400 Bad Request` - Invalid account ID format
- `404 Not Found` - Account does not exist
- `500 Internal Server Error` - System error

**Example Requests:**
```bash
curl http://localhost:8080/accounts/123
```

**Example Responses:**
```json
// Success
{
  "account_id": 123,
  "balance": "1000.50"
}

// Account not found
{"error": "account not found"}

// Invalid ID format
{"error": "account_id must be a positive integer"}
```

### 3. Submit Transaction

Processes financial transfer between two accounts atomically.

```http
POST /transactions
Content-Type: application/json
```

**Request Body:**
```json
{
  "source_account_id": 123,
  "destination_account_id": 456,
  "amount": "150.50000"
}
```

**Request Schema:**
| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `source_account_id` | integer | Yes | > 0, exists | Account to transfer from |
| `destination_account_id` | integer | Yes | > 0, exists | Account to transfer to |
| `amount` | string | Yes | > 0, decimal | Transfer amount as decimal string |

**Response Schema:**
```json
{
  "transaction_id": 42,
  "status": "completed",
  "message": "Transaction processed successfully"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `transaction_id` | integer | Unique identifier for the transaction |
| `status` | string | Processing status ("completed" or "failed") |
| `message` | string | Human-readable result message |

**Response Codes:**
- `201 Created` - Transaction processed successfully
- `400 Bad Request` - Invalid input or business rule violation
- `404 Not Found` - One or both accounts not found
- `500 Internal Server Error` - System error

**Example Requests:**
```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "source_account_id": 123,
    "destination_account_id": 456,
    "amount": "150.50"
  }'
```

**Error Examples:**
```json
// Insufficient balance
{"error": "insufficient balance"}

// Same account transfer
{"error": "source and destination accounts cannot be the same"}

// Invalid amount
{"error": "amount must be positive"}

// Account not found
{"error": "one or both accounts not found"}
```

### 4. Health Check

Returns service health status for monitoring.

```http
GET /health
```

**Response:**
- `200 OK` - Service operational (returns "OK")
- `500 Internal Server Error` - Service unavailable

**Example Request:**
```bash
curl http://localhost:8080/health
```

**Response:**
```
OK
```

## Complete Workflow Example

### Step 1: Create Accounts
```bash
# Create source account
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1001, "initial_balance": "1000.00"}'
# Response: 201 Created (empty body)

# Create destination account
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1002, "initial_balance": "500.00"}'
# Response: 201 Created (empty body)
```

### Step 2: Verify Initial Balances
```bash
# Check source account
curl http://localhost:8080/accounts/1001
# Response: {"account_id": 1001, "balance": "1000"}

# Check destination account
curl http://localhost:8080/accounts/1002
# Response: {"account_id": 1002, "balance": "500"}
```

### Step 3: Process Transfer
```bash
# Transfer $150.50 from 1001 to 1002
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "source_account_id": 1001,
    "destination_account_id": 1002,
    "amount": "150.50"
  }'
# Response: {
#   "transaction_id": 1,
#   "status": "completed",
#   "message": "Transaction processed successfully"
# }
```

### Step 4: Verify Updated Balances
```bash
# Check source account (should be $849.50)
curl http://localhost:8080/accounts/1001
# Response: {"account_id": 1001, "balance": "849.5"}

# Check destination account (should be $650.50)
curl http://localhost:8080/accounts/1002
# Response: {"account_id": 1002, "balance": "650.5"}
```

## Error Handling

### Error Response Format
All errors return JSON with consistent structure:
```json
{
  "error": "Human-readable error message"
}
```

### Common Error Scenarios

#### 400 Bad Request
- Invalid JSON payload
- Missing required fields
- Invalid data types or formats
- Business rule violations (negative amounts, same account transfers)

#### 404 Not Found
- Account does not exist
- Invalid endpoint URL

#### 409 Conflict
- Duplicate account ID during creation

#### 500 Internal Server Error
- Database connection failures
- System unavailability
- Unexpected server errors

### Error Prevention

1. **Input Validation**: Validate all inputs before processing
2. **Existence Checks**: Verify accounts exist before transactions
3. **Balance Verification**: Check sufficient funds before transfers
4. **Format Validation**: Ensure decimal strings are properly formatted
5. **Range Checking**: Verify values are within acceptable ranges

## Data Types & Formats

### Account ID
- **Type**: Positive integer
- **Range**: 1 to 2,147,483,647 (32-bit signed integer max)
- **Format**: Plain integer (no quotes in JSON)
- **Example**: `123`

### Monetary Amounts
- **Type**: Decimal string
- **Precision**: Up to 5 decimal places
- **Range**: 0.00001 to 9999999999.99999
- **Format**: String with decimal notation
- **Examples**: `"1000.50"`, `"0.00001"`, `"999999.99999"`

### Transaction ID
- **Type**: Positive integer (auto-generated)
- **Range**: 1 to 2,147,483,647
- **Format**: Plain integer
- **Example**: `42`

### Status Values
- **Type**: String enumeration
- **Values**: `"completed"`, `"failed"`
- **Format**: Lowercase string
- **Example**: `"completed"`

## Rate Limiting & Performance

### Performance Characteristics
- **Response Time**: < 100ms for 95th percentile
- **Throughput**: > 1,000 requests per second
- **Concurrency**: Supports 100+ concurrent connections
- **Availability**: 99.9% uptime target

### Recommended Usage Patterns
1. **Batch Operations**: Group related operations when possible
2. **Idempotency**: Implement retry logic for failed requests
3. **Connection Pooling**: Reuse HTTP connections for multiple requests
4. **Error Handling**: Implement exponential backoff for retries

## Security Considerations

### Data Protection
- **Input Validation**: All inputs validated and sanitized
- **SQL Injection Prevention**: Parameterized queries used throughout
- **Error Information**: Minimal error details to prevent information leakage
- **Audit Logging**: All transactions logged for security monitoring

### Network Security
- **HTTPS**: Use HTTPS in production environments
- **Network Isolation**: Deploy on secure internal network
- **Access Control**: Implement network-level access restrictions
- **Monitoring**: Monitor for unusual access patterns

## Integration Guidelines

### Client Implementation
1. **HTTP Client**: Use robust HTTP client with timeout handling
2. **JSON Parsing**: Implement proper JSON parsing with error handling
3. **Retry Logic**: Implement exponential backoff for failed requests
4. **Connection Management**: Use connection pooling for performance
5. **Error Handling**: Handle all possible error responses gracefully

### Testing Recommendations
1. **Unit Tests**: Test individual API endpoints
2. **Integration Tests**: Test complete workflows
3. **Error Testing**: Test all error scenarios
4. **Load Testing**: Verify performance under load
5. **Concurrent Testing**: Test concurrent operations

### Monitoring & Observability
1. **Health Checks**: Regular polling of `/health` endpoint
2. **Response Time Monitoring**: Track API response times
3. **Error Rate Monitoring**: Monitor for increased error rates
4. **Transaction Volume**: Track transaction processing volume
5. **Balance Verification**: Periodic balance reconciliation

## Support & Troubleshooting

### Diagnostic Information
- **Request ID**: Include unique request identifiers for tracking
- **Timestamp**: Log request timestamps for correlation
- **Client Information**: Include client identifier in requests
- **Version Information**: Track API version usage

### Common Issues
1. **Account Not Found**: Verify account was created successfully
2. **Insufficient Balance**: Check current balance before transfers
3. **Invalid Format**: Ensure decimal amounts are properly formatted
4. **Duplicate Account**: Check if account ID already exists
5. **Network Timeout**: Implement proper timeout and retry logic

### Getting Help
- **Documentation**: Refer to this specification and README
- **Testing**: Use provided test examples for verification
- **Logs**: Check application logs for detailed error information
- **Health Check**: Verify service availability via `/health` endpoint

---

**Document Version**: 1.0.0  
**Last Updated**: 2024  
**API Version**: 1.0.0