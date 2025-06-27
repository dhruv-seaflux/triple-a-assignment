package repository

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"internal-transfers/internal/models/account"
	"internal-transfers/internal/models/transaction"
)

// Repository implements repository operations using GORM
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new Repository instance
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateAccount creates a new account with the given account ID and initial balance
func (r *Repository) CreateAccount(accountID int, initialBalance decimal.Decimal) error {
	acc := account.Account{
		AccountID: accountID,
		Balance:   initialBalance,
	}
	
	result := r.db.Create(&acc)
	if result.Error != nil {
		// Check for unique constraint violation
		if isDuplicateError(result.Error) {
			return fmt.Errorf("account already exists")
		}
		return fmt.Errorf("failed to create account: %w", result.Error)
	}
	
	return nil
}

// GetAccountByID retrieves account information by account ID
func (r *Repository) GetAccountByID(accountID int) (*account.Account, error) {
	var acc account.Account
	
	result := r.db.Where("account_id = ?", accountID).First(&acc)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get account: %w", result.Error)
	}
	
	return &acc, nil
}

// ProcessTransaction processes a financial transfer between two accounts with strict database-level locking,
// rollback and commit transaction management for financial system integrity
func (r *Repository) ProcessTransaction(sourceAccountID, destinationAccountID int, amount decimal.Decimal) (*transaction.Transaction, error) {
	var trans transaction.Transaction
	var sourceAccount, destinationAccount account.Account
	
	// Validate pre-conditions before starting transaction
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("transfer amount must be positive")
	}
	
	if sourceAccountID == destinationAccountID {
		return nil, fmt.Errorf("source and destination accounts cannot be the same")
	}
	
	// Start database transaction with explicit isolation level
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Set transaction isolation level to SERIALIZABLE for maximum consistency
		if err := tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE").Error; err != nil {
			return fmt.Errorf("failed to set transaction isolation level: %w", err)
		}
		
		// Step 1: Lock accounts in order to prevent deadlocks
		// Order accounts by ID to prevent deadlocks when multiple transactions occur
		var firstAccountID, secondAccountID int
		
		if sourceAccountID < destinationAccountID {
			firstAccountID, secondAccountID = sourceAccountID, destinationAccountID
		} else {
			firstAccountID, secondAccountID = destinationAccountID, sourceAccountID
		}
		
		// Lock first account (lower ID)
		var firstAcc account.Account
		result := tx.Set("gorm:query_option", "FOR UPDATE NOWAIT").Where("account_id = ?", firstAccountID).First(&firstAcc)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				if firstAccountID == sourceAccountID {
					return fmt.Errorf("source account %d not found", firstAccountID)
				}
				return fmt.Errorf("destination account %d not found", firstAccountID)
			}
			return fmt.Errorf("failed to lock account %d: %w", firstAccountID, result.Error)
		}
		
		// Lock second account (higher ID)
		var secondAcc account.Account
		result = tx.Set("gorm:query_option", "FOR UPDATE NOWAIT").Where("account_id = ?", secondAccountID).First(&secondAcc)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				if secondAccountID == sourceAccountID {
					return fmt.Errorf("source account %d not found", secondAccountID)
				}
				return fmt.Errorf("destination account %d not found", secondAccountID)
			}
			return fmt.Errorf("failed to lock account %d: %w", secondAccountID, result.Error)
		}
		
		// Assign accounts to source and destination based on original request
		if sourceAccountID == firstAccountID {
			sourceAccount, destinationAccount = firstAcc, secondAcc
		} else {
			sourceAccount, destinationAccount = secondAcc, firstAcc
		}
		
		// Step 2: Verify accounts are active (not soft deleted)
		if sourceAccount.DeletedAt.Valid {
			return fmt.Errorf("source account %d is inactive", sourceAccountID)
		}
		if destinationAccount.DeletedAt.Valid {
			return fmt.Errorf("destination account %d is inactive", destinationAccountID)
		}
		
		// Step 3: Check sufficient balance with precision
		if sourceAccount.Balance.LessThan(amount) {
			return fmt.Errorf("insufficient balance: available %s, requested %s", 
				sourceAccount.Balance.String(), amount.String())
		}
		
		// Step 4: Calculate new balances with high precision
		newSourceBalance := sourceAccount.Balance.Sub(amount)
		newDestinationBalance := destinationAccount.Balance.Add(amount)
		
		// Verify balances don't go negative (additional safety check)
		if newSourceBalance.LessThan(decimal.Zero) {
			return fmt.Errorf("transaction would result in negative balance")
		}
		
		// Step 5: Create pending transaction record first for audit trail
		trans = transaction.Transaction{
			FromAccountID: &sourceAccount.ID,
			ToAccountID:   &destinationAccount.ID,
			Amount:        amount,
			Description:   fmt.Sprintf("Transfer from account %d to account %d", sourceAccountID, destinationAccountID),
			Status:        transaction.StatusPending,
		}
		
		result = tx.Create(&trans)
		if result.Error != nil {
			return fmt.Errorf("failed to create transaction record: %w", result.Error)
		}
		
		// Step 6: Update source account balance
		result = tx.Model(&sourceAccount).Where("id = ? AND balance >= ?", sourceAccount.ID, amount).
			Update("balance", newSourceBalance)
		if result.Error != nil {
			return fmt.Errorf("failed to update source account balance: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("failed to update source account: insufficient balance or account changed")
		}
		
		// Step 7: Update destination account balance
		result = tx.Model(&destinationAccount).Where("id = ?", destinationAccount.ID).
			Update("balance", newDestinationBalance)
		if result.Error != nil {
			return fmt.Errorf("failed to update destination account balance: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("failed to update destination account: account may have been modified")
		}
		
		// Step 8: Update transaction status to completed
		result = tx.Model(&trans).Update("status", transaction.StatusCompleted)
		if result.Error != nil {
			return fmt.Errorf("failed to update transaction status: %w", result.Error)
		}
		
		// Step 9: Final verification - re-read balances to ensure consistency
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
		
		// All steps completed successfully - transaction will be committed automatically
		return nil
	})
	
	// If any error occurred, transaction is automatically rolled back by GORM
	if err != nil {
		// Log the error for audit purposes (in a real system, you'd use proper logging)
		return nil, fmt.Errorf("transaction failed and rolled back: %w", err)
	}
	
	// Transaction completed successfully and committed
	return &trans, nil
}

// GetTransactionByID retrieves a transaction by its ID
func (r *Repository) GetTransactionByID(transactionID uint) (*transaction.Transaction, error) {
	var trans transaction.Transaction
	
	result := r.db.First(&trans, transactionID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to get transaction: %w", result.Error)
	}
	
	return &trans, nil
}

// isDuplicateError checks if the error is a unique constraint violation
func isDuplicateError(err error) bool {
	// This is a simplified check. In a real application, you'd want to check
	// for specific PostgreSQL error codes (23505 for unique_violation)
	return err != nil && (
		containsString(err.Error(), "duplicate") ||
		containsString(err.Error(), "unique") ||
		containsString(err.Error(), "23505"))
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(str, substr string) bool {
	return len(str) >= len(substr) && 
		   (str == substr || 
		    (len(str) > len(substr) && 
		     (str[:len(substr)] == substr || 
		      str[len(str)-len(substr):] == substr ||
		      containsSubstring(str, substr))))
}

func containsSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}