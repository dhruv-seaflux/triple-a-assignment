package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"internal-transfers/internal/models"
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAccountByID(accountID int) (*models.Account, error) {
	query := `SELECT id, account_id, balance, created_at, updated_at 
			  FROM accounts WHERE account_id = $1`
	
	var account models.Account
	var balanceStr string
	
	err := r.db.QueryRow(query, accountID).Scan(
		&account.ID,
		&account.AccountID,
		&balanceStr,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to query account: %w", err)
	}
	
	balance, err := decimal.NewFromString(balanceStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance: %w", err)
	}
	account.Balance = balance
	
	return &account, nil
}

func (r *Repository) CreateAccount(accountID int, initialBalance decimal.Decimal) (*models.Account, error) {
	if initialBalance.IsNegative() {
		return nil, ErrInvalidAmount
	}

	query := `INSERT INTO accounts (account_id, balance) 
			  VALUES ($1, $2) 
			  RETURNING id, account_id, balance, created_at, updated_at`
	
	var account models.Account
	var balanceStr string
	
	err := r.db.QueryRow(query, accountID, initialBalance.String()).Scan(
		&account.ID,
		&account.AccountID,
		&balanceStr,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrAccountAlreadyExists
		}
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	
	balance, err := decimal.NewFromString(balanceStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance: %w", err)
	}
	account.Balance = balance
	
	return &account, nil
}

func (r *Repository) UpdateAccountBalance(accountID int, newBalance decimal.Decimal) error {
	query := `UPDATE accounts SET balance = $1, updated_at = $2 WHERE id = $3`
	result, err := r.db.Exec(query, newBalance.String(), time.Now(), accountID)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return ErrAccountNotFound
	}
	
	return nil
}

func (r *Repository) CreateTransaction(tx *sql.Tx, fromAccountID, toAccountID *int, amount decimal.Decimal, description string) (*models.Transaction, error) {
	query := `INSERT INTO transactions (from_account_id, to_account_id, amount, description, status) 
			  VALUES ($1, $2, $3, $4, 'completed') 
			  RETURNING id, from_account_id, to_account_id, amount, description, status, created_at, updated_at`
	
	var transaction models.Transaction
	var amountStr string
	
	err := tx.QueryRow(query, fromAccountID, toAccountID, amount.String(), description).Scan(
		&transaction.ID,
		&transaction.FromAccountID,
		&transaction.ToAccountID,
		&amountStr,
		&transaction.Description,
		&transaction.Status,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	
	parsedAmount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction amount: %w", err)
	}
	transaction.Amount = parsedAmount
	
	return &transaction, nil
}

// ProcessTransaction handles money transfer between accounts using account IDs
// This method ensures atomicity and data consistency during the transfer process
func (r *Repository) ProcessTransaction(sourceAccountID, destinationAccountID int, amount decimal.Decimal) (*models.Transaction, error) {
	// Validate input parameters
	if amount.IsNegative() || amount.IsZero() {
		return nil, ErrInvalidAmount
	}
	
	if sourceAccountID == destinationAccountID {
		return nil, fmt.Errorf("source and destination accounts cannot be the same")
	}
	
	if sourceAccountID <= 0 || destinationAccountID <= 0 {
		return nil, fmt.Errorf("invalid account ID")
	}

	// Begin database transaction for atomicity
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get source account and validate it exists
	sourceAccount, err := r.GetAccountByID(sourceAccountID)
	if err != nil {
		return nil, fmt.Errorf("source account error: %w", err)
	}

	// Get destination account and validate it exists
	destinationAccount, err := r.GetAccountByID(destinationAccountID)
	if err != nil {
		return nil, fmt.Errorf("destination account error: %w", err)
	}

	// Check sufficient balance
	if sourceAccount.Balance.LessThan(amount) {
		return nil, ErrInsufficientBalance
	}

	// Update account balances atomically
	updateQuery := `UPDATE accounts SET balance = $1, updated_at = $2 WHERE id = $3`
	
	newSourceBalance := sourceAccount.Balance.Sub(amount)
	_, err = tx.Exec(updateQuery, newSourceBalance.String(), time.Now(), sourceAccount.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update source account: %w", err)
	}

	newDestinationBalance := destinationAccount.Balance.Add(amount)
	_, err = tx.Exec(updateQuery, newDestinationBalance.String(), time.Now(), destinationAccount.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update destination account: %w", err)
	}

	// Create transaction record
	description := fmt.Sprintf("Transfer from account %d to account %d", sourceAccountID, destinationAccountID)
	transaction, err := r.CreateTransaction(tx, &sourceAccount.ID, &destinationAccount.ID, amount, description)
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return transaction, nil
}

func (r *Repository) ProcessTransfer(fromAccountID, toAccountID int, amount decimal.Decimal, description string) (*models.Transaction, error) {
	if amount.IsNegative() || amount.IsZero() {
		return nil, ErrInvalidAmount
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	fromAccount, err := r.GetAccountByID(fromAccountID)
	if err != nil {
		return nil, fmt.Errorf("source account error: %w", err)
	}

	toAccount, err := r.GetAccountByID(toAccountID)
	if err != nil {
		return nil, fmt.Errorf("destination account error: %w", err)
	}

	if fromAccount.Balance.LessThan(amount) {
		return nil, ErrInsufficientBalance
	}

	updateQuery := `UPDATE accounts SET balance = $1, updated_at = $2 WHERE id = $3`
	
	newFromBalance := fromAccount.Balance.Sub(amount)
	_, err = tx.Exec(updateQuery, newFromBalance.String(), time.Now(), fromAccount.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update source account: %w", err)
	}

	newToBalance := toAccount.Balance.Add(amount)
	_, err = tx.Exec(updateQuery, newToBalance.String(), time.Now(), toAccount.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update destination account: %w", err)
	}

	transaction, err := r.CreateTransaction(tx, &fromAccount.ID, &toAccount.ID, amount, description)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return transaction, nil
}