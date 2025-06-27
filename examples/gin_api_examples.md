# Gin-Based API Examples

This document provides examples of how to use the refactored Internal Transfers API with Gin, GORM, and database transactions.

## API Endpoints

### 1. Create Account - POST /accounts

**Request:**
```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": 1001,
    "initial_balance": "1000.50"
  }'
```

**Success Response (201 Created):**
```json
{
  "message": "Account created successfully"
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "account_id must be a positive integer"
}
```

**Error Response (409 Conflict):**
```json
{
  "error": "account already exists"
}
```

### 2. Get Account Balance - GET /accounts/{account_id}

**Request:**
```bash
curl http://localhost:8080/accounts/1001
```

**Success Response (200 OK):**
```json
{
  "account_id": 1001,
  "balance": "1000.50"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "account not found"
}
```

### 3. Submit Transaction - POST /submit (WITH GORM DATABASE TRANSACTION)

This endpoint demonstrates the GORM database transaction implementation where the operation commits if everything succeeds and rolls back on failure.

**Request:**
```bash
curl -X POST http://localhost:8080/submit \
  -H "Content-Type: application/json" \
  -d '{
    "source_account_id": 1001,
    "destination_account_id": 1002,
    "amount": "150.25"
  }'
```

**Success Response (201 Created):**
```json
{
  "transaction_id": 42,
  "status": "completed",
  "message": "Transaction processed successfully"
}
```

**Error Responses:**

**Insufficient Balance (400 Bad Request):**
```json
{
  "error": "insufficient balance: available 100.50000, requested 150.00000"
}
```

**Same Account Transfer (400 Bad Request):**
```json
{
  "error": "source and destination accounts cannot be the same"
}
```

**Account Not Found (404 Not Found):**
```json
{
  "error": "account 9999 not found"
}
```

**Account Inactive (400 Bad Request):**
```json
{
  "error": "source account 1001 is inactive"
}
```

**Lock Failure / Concurrent Access (400 Bad Request):**
```json
{
  "error": "failed to lock account 1001: could not obtain lock"
}
```

### 4. Health Check - GET /health

**Request:**
```bash
curl http://localhost:8080/health
```

**Success Response (200 OK):**
```
OK
```

## Complete Workflow Example

```bash
# 1. Create source account
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1001, "initial_balance": "1000.00"}'
# Response: {"message": "Account created successfully"}

# 2. Create destination account  
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1002, "initial_balance": "500.00"}'
# Response: {"message": "Account created successfully"}

# 3. Check initial balances
curl http://localhost:8080/accounts/1001
# Response: {"account_id": 1001, "balance": "1000"}

curl http://localhost:8080/accounts/1002  
# Response: {"account_id": 1002, "balance": "500"}

# 4. Submit transaction (GORM database transaction)
curl -X POST http://localhost:8080/submit \
  -H "Content-Type: application/json" \
  -d '{
    "source_account_id": 1001,
    "destination_account_id": 1002,
    "amount": "150.50"
  }'
# Response: {"transaction_id": 1, "status": "completed", "message": "Transaction processed successfully"}

# 5. Verify updated balances
curl http://localhost:8080/accounts/1001
# Response: {"account_id": 1001, "balance": "849.5"}

curl http://localhost:8080/accounts/1002
# Response: {"account_id": 1002, "balance": "650.5"}
```

## Enhanced Financial Transaction System

The `/submit` endpoint implements a bank-grade financial transaction system with strict database-level locking, rollback, and commit transaction management:

### **Transaction Features:**

1. **SERIALIZABLE Isolation Level**: Maximum consistency for financial operations
2. **Deadlock Prevention**: Orders account locking by ID to prevent deadlocks
3. **Row-Level Locking**: Uses `FOR UPDATE NOWAIT` for immediate lock acquisition
4. **Audit Trail**: Creates pending transaction record before balance changes
5. **Balance Verification**: Double-checks final balances for consistency
6. **Automatic Rollback**: Any error automatically rolls back entire transaction

### **Transaction Flow:**

```go
err := r.db.Transaction(func(tx *gorm.DB) error {
    // 1. Set SERIALIZABLE isolation level for maximum consistency
    tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
    
    // 2. Lock accounts in order to prevent deadlocks (lower ID first)
    result := tx.Set("gorm:query_option", "FOR UPDATE NOWAIT").Where("account_id = ?", firstAccountID).First(&firstAcc)
    result = tx.Set("gorm:query_option", "FOR UPDATE NOWAIT").Where("account_id = ?", secondAccountID).First(&secondAcc)
    
    // 3. Verify accounts are active (not soft deleted)
    if sourceAccount.DeletedAt.Valid || destinationAccount.DeletedAt.Valid {
        return fmt.Errorf("account is inactive")
    }
    
    // 4. Check sufficient balance with high precision
    if sourceAccount.Balance.LessThan(amount) {
        return fmt.Errorf("insufficient balance: available %.5f, requested %.5f", 
            sourceAccount.Balance, amount)
    }
    
    // 5. Create PENDING transaction record for audit trail
    trans = transaction.Transaction{
        FromAccountID: &sourceAccount.ID,
        ToAccountID:   &destinationAccount.ID,
        Amount:        amount,
        Status:        transaction.StatusPending,  // Start as PENDING
    }
    result = tx.Create(&trans)
    
    // 6. Update balances with conditional updates
    result = tx.Model(&sourceAccount).Where("id = ? AND balance >= ?", sourceAccount.ID, amount).
        Update("balance", newSourceBalance)
    if result.RowsAffected == 0 {
        return fmt.Errorf("insufficient balance or account changed")
    }
    
    result = tx.Model(&destinationAccount).Where("id = ?", destinationAccount.ID).
        Update("balance", newDestinationBalance)
    
    // 7. Update transaction status to COMPLETED
    result = tx.Model(&trans).Update("status", transaction.StatusCompleted)
    
    // 8. Final verification - re-read and verify balances
    var verifySource, verifyDestination account.Account
    tx.Where("id = ?", sourceAccount.ID).First(&verifySource)
    tx.Where("id = ?", destinationAccount.ID).First(&verifyDestination)
    
    if !verifySource.Balance.Equal(newSourceBalance) {
        return fmt.Errorf("balance verification failed")
    }
    
    return nil  // Success - commits automatically
})

// Any error automatically triggers rollback
if err != nil {
    return nil, fmt.Errorf("transaction failed and rolled back: %w", err)
}
```

### **Key Financial Safety Features:**

1. **Database-Level Locking**: 
   - `FOR UPDATE NOWAIT` locks specific account rows
   - Prevents concurrent modifications during transfer
   - Fails fast instead of waiting for locks

2. **Deadlock Prevention**:
   - Always locks accounts in consistent order (by ID)
   - Prevents circular waiting conditions
   - Ensures system remains responsive under load

3. **ACID Compliance**:
   - **Atomicity**: All operations succeed or fail together
   - **Consistency**: Maintains account balance integrity
   - **Isolation**: SERIALIZABLE level prevents interference
   - **Durability**: Committed transactions are permanent

4. **Rollback & Commit**:
   - **Automatic Rollback**: Any error rolls back entire transaction
   - **Automatic Commit**: Success commits all changes atomically
   - **No Partial Updates**: Prevents incomplete financial transfers

5. **Audit Trail**:
   - Records transaction as PENDING before balance changes
   - Updates to COMPLETED only after successful transfer
   - Maintains complete history of all transaction attempts

## Error Handling Examples

### Validation Errors
```bash
# Invalid account ID
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": -1, "initial_balance": "1000.00"}'
# Response: {"error": "validation failed for field 'account_id': account_id must be a positive integer"}

# Negative balance
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1003, "initial_balance": "-100.00"}'
# Response: {"error": "validation failed for field 'initial_balance': initial_balance cannot be negative"}
```

### Business Rule Violations
```bash
# Same account transfer
curl -X POST http://localhost:8080/submit \
  -H "Content-Type: application/json" \
  -d '{
    "source_account_id": 1001,
    "destination_account_id": 1001,
    "amount": "100.00"
  }'
# Response: {"error": "validation failed for field 'accounts': source and destination accounts cannot be the same"}

# Insufficient balance
curl -X POST http://localhost:8080/submit \
  -H "Content-Type: application/json" \
  -d '{
    "source_account_id": 1001,
    "destination_account_id": 1002,
    "amount": "99999.00"
  }'
# Response: {"error": "insufficient balance"}
```

## Response Format Summary

All API responses follow the standardized JSON format:

**Success Responses:**
- Simple operations: `{"message": "<success_message>"}`
- Data queries: `{"account_id": 123, "balance": "1000.50"}`
- Transactions: `{"transaction_id": 42, "status": "completed", "message": "Transaction processed successfully"}`

**Error Responses:**
- All errors: `{"error": "<error_message>"}`

**HTTP Status Codes:**
- `200 OK`: Successful queries
- `201 Created`: Successful creation operations
- `400 Bad Request`: Validation errors, business rule violations
- `404 Not Found`: Resource not found
- `409 Conflict`: Duplicate resource (account already exists)
- `500 Internal Server Error`: Unexpected server errors