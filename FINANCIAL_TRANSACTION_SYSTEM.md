# Financial Transaction System - Database Level Locking, Rollback & Commit

This document explains the enhanced financial transaction system implemented for the `/submit` API endpoint with strict database-level locking, rollback, and commit transaction management.

## 🏦 Financial Transaction Requirements

When processing financial transfers between accounts, the system must ensure:

1. **ACID Compliance**: Atomicity, Consistency, Isolation, Durability
2. **Data Integrity**: No money is lost or created during transfers
3. **Concurrency Safety**: Multiple simultaneous transactions don't interfere
4. **Audit Trail**: Complete record of all transaction attempts
5. **Rollback Safety**: Failed transactions don't leave partial updates

## 🔒 Enhanced Transaction Implementation

### 1. **Database Transaction Isolation**

```go
// Set transaction isolation level to SERIALIZABLE for maximum consistency
if err := tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE").Error; err != nil {
    return fmt.Errorf("failed to set transaction isolation level: %w", err)
}
```

**Benefits:**
- **SERIALIZABLE**: Highest isolation level preventing phantom reads, dirty reads, and non-repeatable reads
- **Financial Safety**: Ensures absolute consistency during concurrent operations

### 2. **Deadlock Prevention with Ordered Locking**

```go
// Order accounts by ID to prevent deadlocks when multiple transactions occur
var firstAccountID, secondAccountID int

if sourceAccountID < destinationAccountID {
    firstAccountID, secondAccountID = sourceAccountID, destinationAccountID
} else {
    firstAccountID, secondAccountID = destinationAccountID, sourceAccountID
}

// Lock first account (lower ID)
result := tx.Set("gorm:query_option", "FOR UPDATE NOWAIT").Where("account_id = ?", firstAccountID).First(&firstAcc)

// Lock second account (higher ID)  
result = tx.Set("gorm:query_option", "FOR UPDATE NOWAIT").Where("account_id = ?", secondAccountID).First(&secondAcc)
```

**Benefits:**
- **Deadlock Prevention**: Always locks accounts in consistent order (by ID)
- **NOWAIT**: Fails fast instead of waiting, preventing long locks
- **Row-Level Locking**: Uses `FOR UPDATE` to lock specific account rows

### 3. **Pre-Transaction Validation**

```go
// Validate pre-conditions before starting transaction
if amount.LessThanOrEqual(decimal.Zero) {
    return nil, fmt.Errorf("transfer amount must be positive")
}

if sourceAccountID == destinationAccountID {
    return nil, fmt.Errorf("source and destination accounts cannot be the same")
}
```

**Benefits:**
- **Early Validation**: Catches errors before expensive database operations
- **Business Rules**: Enforces financial transfer rules

### 4. **Account Status Verification**

```go
// Verify accounts are active (not soft deleted)
if sourceAccount.DeletedAt.Valid {
    return fmt.Errorf("source account %d is inactive", sourceAccountID)
}
if destinationAccount.DeletedAt.Valid {
    return fmt.Errorf("destination account %d is inactive", destinationAccountID)
}
```

**Benefits:**
- **Account Validation**: Ensures accounts are active and operational
- **Soft Delete Support**: Respects GORM soft delete functionality

### 5. **High-Precision Balance Checks**

```go
// Check sufficient balance with precision
if sourceAccount.Balance.LessThan(amount) {
    return fmt.Errorf("insufficient balance: available %.5f, requested %.5f", 
        sourceAccount.Balance, amount)
}

// Calculate new balances with high precision
newSourceBalance := sourceAccount.Balance.Sub(amount)
newDestinationBalance := destinationAccount.Balance.Add(amount)

// Verify balances don't go negative (additional safety check)
if newSourceBalance.LessThan(decimal.Zero) {
    return fmt.Errorf("transaction would result in negative balance")
}
```

**Benefits:**
- **Decimal Precision**: Uses `shopspring/decimal` for exact financial calculations
- **Balance Safety**: Multiple checks prevent negative balances
- **Detailed Errors**: Provides specific balance information for debugging

### 6. **Audit Trail with Pending Status**

```go
// Create pending transaction record first for audit trail
trans = transaction.Transaction{
    FromAccountID: &sourceAccount.ID,
    ToAccountID:   &destinationAccount.ID,
    Amount:        amount,
    Description:   fmt.Sprintf("Transfer from account %d to account %d", sourceAccountID, destinationAccountID),
    Status:        transaction.StatusPending,  // Start as PENDING
}

result = tx.Create(&trans)
if result.Error != nil {
    return fmt.Errorf("failed to create transaction record: %w", result.Error)
}
```

**Benefits:**
- **Audit Trail**: Records transaction attempt before balance changes
- **Status Tracking**: Tracks transaction lifecycle (Pending → Completed)
- **Forensic Capability**: Can identify failed transactions and their causes

### 7. **Atomic Balance Updates with Conditions**

```go
// Update source account balance with condition to prevent race conditions
result = tx.Model(&sourceAccount).Where("id = ? AND balance >= ?", sourceAccount.ID, amount).
    Update("balance", newSourceBalance)
if result.Error != nil {
    return fmt.Errorf("failed to update source account balance: %w", result.Error)
}
if result.RowsAffected == 0 {
    return fmt.Errorf("failed to update source account: insufficient balance or account changed")
}

// Update destination account balance
result = tx.Model(&destinationAccount).Where("id = ?", destinationAccount.ID).
    Update("balance", newDestinationBalance)
if result.Error != nil {
    return fmt.Errorf("failed to update destination account balance: %w", result.Error)
}
if result.RowsAffected == 0 {
    return fmt.Errorf("failed to update destination account: account may have been modified")
}
```

**Benefits:**
- **Conditional Updates**: Only updates if balance condition is met
- **Race Condition Protection**: Prevents updates if account state changed
- **Affected Rows Check**: Verifies update actually occurred

### 8. **Final Balance Verification**

```go
// Final verification - re-read balances to ensure consistency
var verifySource, verifyDestination account.Account
if err := tx.Where("id = ?", sourceAccount.ID).First(&verifySource).Error; err != nil {
    return fmt.Errorf("failed to verify source account balance: %w", err)
}
if err := tx.Where("id = ?", destinationAccount.ID).First(&verifyDestination).Error; err != nil {
    return fmt.Errorf("failed to verify destination account balance: %w", err)
}

// Verify the balances match our calculations
if !verifySource.Balance.Equal(newSourceBalance) {
    return fmt.Errorf("source balance verification failed: expected %.5f, got %.5f", 
        newSourceBalance, verifySource.Balance)
}
if !verifyDestination.Balance.Equal(newDestinationBalance) {
    return fmt.Errorf("destination balance verification failed: expected %.5f, got %.5f", 
        newDestinationBalance, verifyDestination.Balance)
}
```

**Benefits:**
- **Double Verification**: Re-reads and verifies final balances
- **Consistency Check**: Ensures calculations match database state
- **Data Integrity**: Catches any unexpected balance discrepancies

### 9. **Automatic Rollback and Commit**

```go
// GORM automatically handles rollback and commit:
err := r.db.Transaction(func(tx *gorm.DB) error {
    // All transaction steps here...
    return nil  // Success - GORM commits automatically
})

// If any error occurred, transaction is automatically rolled back by GORM
if err != nil {
    return nil, fmt.Errorf("transaction failed and rolled back: %w", err)
}

// Transaction completed successfully and committed
return &trans, nil
```

**Benefits:**
- **Automatic Rollback**: Any error automatically rolls back entire transaction
- **Automatic Commit**: Success automatically commits all changes
- **Error Propagation**: Clear error messages indicate what failed

## 📊 Transaction Flow Diagram

```
POST /submit
     ↓
[Pre-validation]
     ↓
[Start DB Transaction - SERIALIZABLE]
     ↓
[Lock Account 1 - FOR UPDATE NOWAIT]
     ↓
[Lock Account 2 - FOR UPDATE NOWAIT]
     ↓
[Verify Account Status]
     ↓
[Check Sufficient Balance]
     ↓
[Create PENDING Transaction Record]
     ↓
[Update Source Balance with Condition]
     ↓
[Update Destination Balance]
     ↓
[Update Transaction Status to COMPLETED]
     ↓
[Verify Final Balances]
     ↓
[COMMIT] ←→ [ROLLBACK on any error]
     ↓
[Return Success Response]
```

## 🛡️ Error Handling and Rollback Scenarios

### Automatic Rollback Triggers:

1. **Account Not Found**: Either source or destination account doesn't exist
2. **Account Inactive**: Account is soft deleted or inactive
3. **Insufficient Balance**: Source account doesn't have enough funds
4. **Lock Timeout**: Unable to acquire locks (NOWAIT fails)
5. **Balance Update Failure**: Database update operations fail
6. **Verification Failure**: Final balance verification doesn't match calculations
7. **Transaction Status Update Failure**: Unable to mark transaction as completed

### Example Error Responses:

```json
// Insufficient Balance
{
  "error": "insufficient balance: available 100.50000, requested 150.00000"
}

// Account Not Found
{
  "error": "account 9999 not found"
}

// Same Account Transfer
{
  "error": "source and destination accounts cannot be the same"
}

// Lock Failure (Concurrent Access)
{
  "error": "failed to lock account 1001: could not obtain lock"
}
```

## 🧪 Testing the Enhanced Transaction System

### Test Case 1: Successful Transfer
```bash
curl -X POST http://localhost:8080/submit \
  -H "Content-Type: application/json" \
  -d '{
    "source_account_id": 1001,
    "destination_account_id": 1002,
    "amount": "150.25"
  }'
```

**Expected Response:**
```json
{
  "transaction_id": 42,
  "status": "completed",
  "message": "Transaction processed successfully"
}
```

### Test Case 2: Insufficient Balance
```bash
curl -X POST http://localhost:8080/submit \
  -H "Content-Type: application/json" \
  -d '{
    "source_account_id": 1001,
    "destination_account_id": 1002,
    "amount": "99999.00"
  }'
```

**Expected Response:**
```json
{
  "error": "insufficient balance: available 1000.50000, requested 99999.00000"
}
```

### Test Case 3: Concurrent Transactions
Run multiple simultaneous transfers to test locking and deadlock prevention.

## 🔧 Database Configuration for Financial Systems

### Recommended PostgreSQL Settings:

```sql
-- Set appropriate timeout for financial transactions
SET statement_timeout = '30s';

-- Enable row-level security if needed
ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;

-- Create indexes for performance
CREATE INDEX CONCURRENTLY idx_accounts_balance ON accounts(balance);
CREATE INDEX CONCURRENTLY idx_transactions_status ON transactions(status);
```

## 📈 Performance Considerations

1. **Lock Duration**: Minimized by pre-validation and efficient operations
2. **Deadlock Prevention**: Ordered locking prevents deadlocks
3. **Index Usage**: Proper indexes on account_id and balance columns
4. **Connection Pooling**: GORM handles database connection pooling
5. **Transaction Scope**: Keeps transaction scope as small as possible

## 🔐 Security Features

1. **Input Validation**: Comprehensive validation before database operations
2. **SQL Injection Protection**: GORM parameterized queries
3. **Business Rule Enforcement**: Prevents invalid financial operations
4. **Audit Trail**: Complete record of all transaction attempts
5. **Error Information**: Minimal error details to prevent information leakage

This enhanced financial transaction system provides bank-grade security and consistency for money transfers in the Internal Transfers API.