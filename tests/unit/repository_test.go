package unit

import (
	"database/sql/driver"
	"internal-transfers/internal/repository"
	"internal-transfers/tests/testutils"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRepository_CreateAccount tests account creation with various scenarios
func TestRepository_CreateAccount(t *testing.T) {
	tests := []struct {
		name           string
		accountID      int
		initialBalance decimal.Decimal
		mockSetup      func(sqlmock.Sqlmock)
		expectError    bool
		expectedError  error
	}{
		{
			name:           "Successful account creation",
			accountID:      123,
			initialBalance: decimal.NewFromFloat(1000.50),
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "account_id", "balance", "created_at", "updated_at"}).
					AddRow(1, 123, "1000.50", time.Now(), time.Now())
				mock.ExpectQuery(`INSERT INTO accounts \(account_id, balance\)`).
					WithArgs(123, "1000.50").
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name:           "Account creation with zero balance",
			accountID:      456,
			initialBalance: decimal.Zero,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "account_id", "balance", "created_at", "updated_at"}).
					AddRow(2, 456, "0", time.Now(), time.Now())
				mock.ExpectQuery(`INSERT INTO accounts \(account_id, balance\)`).
					WithArgs(456, "0").
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name:           "Account creation with negative balance should fail",
			accountID:      789,
			initialBalance: decimal.NewFromFloat(-100.00),
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectError:    true,
			expectedError:  repository.ErrInvalidAmount,
		},
		{
			name:           "Duplicate account ID should fail",
			accountID:      123,
			initialBalance: decimal.NewFromFloat(500.00),
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO accounts \(account_id, balance\)`).
					WithArgs(123, "500").
					WillReturnError(&MockPQError{Code: "23505"}) // Unique violation
			},
			expectError:   true,
			expectedError: repository.ErrAccountAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := testutils.SetupMockDatabase(t)
			defer db.Close()

			tt.mockSetup(mock)

			repo := repository.NewRepository(db)
			account, err := repo.CreateAccount(tt.accountID, tt.initialBalance)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
				assert.Nil(t, account)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, account)
				assert.Equal(t, tt.accountID, account.AccountID)
				assert.True(t, tt.initialBalance.Equal(account.Balance))
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestRepository_GetAccountByID tests account retrieval by ID
func TestRepository_GetAccountByID(t *testing.T) {
	tests := []struct {
		name          string
		accountID     int
		mockSetup     func(sqlmock.Sqlmock)
		expectError   bool
		expectedError error
	}{
		{
			name:      "Successfully retrieve existing account",
			accountID: 123,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "account_id", "balance", "created_at", "updated_at"}).
					AddRow(1, 123, "1000.50", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT id, account_id, balance, created_at, updated_at FROM accounts WHERE account_id = \$1`).
					WithArgs(123).
					WillReturnRows(rows)
			},
			expectError: false,
		},
		{
			name:      "Account not found should return error",
			accountID: 999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, account_id, balance, created_at, updated_at FROM accounts WHERE account_id = \$1`).
					WithArgs(999).
					WillReturnError(sqlmock.ErrCancelled) // Simulate no rows
			},
			expectError:   true,
			expectedError: repository.ErrAccountNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := testutils.SetupMockDatabase(t)
			defer db.Close()

			// Update the mock setup to handle the error properly
			if tt.expectedError == repository.ErrAccountNotFound {
				mock.ExpectQuery(`SELECT id, account_id, balance, created_at, updated_at FROM accounts WHERE account_id = \$1`).
					WithArgs(tt.accountID).
					WillReturnError(sqlmock.ErrCancelled)
			} else {
				tt.mockSetup(mock)
			}

			repo := repository.NewRepository(db)
			account, err := repo.GetAccountByID(tt.accountID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, account)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, account)
				assert.Equal(t, tt.accountID, account.AccountID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestRepository_ProcessTransaction tests the complete transaction flow
func TestRepository_ProcessTransaction(t *testing.T) {
	tests := []struct {
		name                 string
		sourceAccountID      int
		destinationAccountID int
		amount               decimal.Decimal
		mockSetup            func(sqlmock.Sqlmock)
		expectError          bool
		expectedError        error
	}{
		{
			name:                 "Successful transaction between accounts",
			sourceAccountID:      123,
			destinationAccountID: 456,
			amount:               decimal.NewFromFloat(100.50),
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Begin transaction
				mock.ExpectBegin()

				// Get source account
				sourceRows := sqlmock.NewRows([]string{"id", "account_id", "balance", "created_at", "updated_at"}).
					AddRow(1, 123, "1000.00", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT id, account_id, balance, created_at, updated_at FROM accounts WHERE account_id = \$1`).
					WithArgs(123).
					WillReturnRows(sourceRows)

				// Get destination account
				destRows := sqlmock.NewRows([]string{"id", "account_id", "balance", "created_at", "updated_at"}).
					AddRow(2, 456, "500.00", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT id, account_id, balance, created_at, updated_at FROM accounts WHERE account_id = \$1`).
					WithArgs(456).
					WillReturnRows(destRows)

				// Update source account balance
				mock.ExpectExec(`UPDATE accounts SET balance = \$1, updated_at = \$2 WHERE id = \$3`).
					WithArgs("899.5", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// Update destination account balance
				mock.ExpectExec(`UPDATE accounts SET balance = \$1, updated_at = \$2 WHERE id = \$3`).
					WithArgs("600.5", sqlmock.AnyArg(), 2).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// Create transaction record
				transactionRows := sqlmock.NewRows([]string{"id", "from_account_id", "to_account_id", "amount", "description", "status", "created_at", "updated_at"}).
					AddRow(1, 1, 2, "100.5", "Transfer from account 123 to account 456", "completed", time.Now(), time.Now())
				mock.ExpectQuery(`INSERT INTO transactions \(from_account_id, to_account_id, amount, description, status\)`).
					WithArgs(1, 2, "100.5", "Transfer from account 123 to account 456", "completed").
					WillReturnRows(transactionRows)

				// Commit transaction
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name:                 "Transaction with insufficient balance should fail",
			sourceAccountID:      123,
			destinationAccountID: 456,
			amount:               decimal.NewFromFloat(2000.00),
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Begin transaction
				mock.ExpectBegin()

				// Get source account with insufficient balance
				sourceRows := sqlmock.NewRows([]string{"id", "account_id", "balance", "created_at", "updated_at"}).
					AddRow(1, 123, "1000.00", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT id, account_id, balance, created_at, updated_at FROM accounts WHERE account_id = \$1`).
					WithArgs(123).
					WillReturnRows(sourceRows)

				// Get destination account
				destRows := sqlmock.NewRows([]string{"id", "account_id", "balance", "created_at", "updated_at"}).
					AddRow(2, 456, "500.00", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT id, account_id, balance, created_at, updated_at FROM accounts WHERE account_id = \$1`).
					WithArgs(456).
					WillReturnRows(destRows)

				// Rollback due to insufficient balance
				mock.ExpectRollback()
			},
			expectError:   true,
			expectedError: repository.ErrInsufficientBalance,
		},
		{
			name:                 "Same account transaction should fail",
			sourceAccountID:      123,
			destinationAccountID: 123,
			amount:               decimal.NewFromFloat(100.00),
			mockSetup:            func(mock sqlmock.Sqlmock) {},
			expectError:          true,
		},
		{
			name:                 "Negative amount should fail",
			sourceAccountID:      123,
			destinationAccountID: 456,
			amount:               decimal.NewFromFloat(-100.00),
			mockSetup:            func(mock sqlmock.Sqlmock) {},
			expectError:          true,
			expectedError:        repository.ErrInvalidAmount,
		},
		{
			name:                 "Zero amount should fail",
			sourceAccountID:      123,
			destinationAccountID: 456,
			amount:               decimal.Zero,
			mockSetup:            func(mock sqlmock.Sqlmock) {},
			expectError:          true,
			expectedError:        repository.ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := testutils.SetupMockDatabase(t)
			defer db.Close()

			tt.mockSetup(mock)

			repo := repository.NewRepository(db)
			transaction, err := repo.ProcessTransaction(tt.sourceAccountID, tt.destinationAccountID, tt.amount)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
				assert.Nil(t, transaction)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, transaction)
				assert.Equal(t, "completed", transaction.Status)
				assert.True(t, tt.amount.Equal(transaction.Amount))
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestRepository_Concurrency tests concurrent operations on the repository
func TestRepository_Concurrency(t *testing.T) {
	// This test would require a real database connection for meaningful concurrency testing
	t.Skip("Concurrency tests require integration test setup with real database")
}

// TestRepository_EdgeCases tests edge cases and boundary conditions
func TestRepository_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		operation   func(*repository.Repository) error
		expectError bool
	}{
		{
			name: "Very large account ID",
			operation: func(repo *repository.Repository) error {
				_, err := repo.CreateAccount(999999999, decimal.NewFromFloat(100.00))
				return err
			},
			expectError: false, // Should handle large IDs
		},
		{
			name: "Very high precision decimal",
			operation: func(repo *repository.Repository) error {
				highPrecision := decimal.RequireFromString("123.123456789")
				_, err := repo.CreateAccount(100, highPrecision)
				return err
			},
			expectError: false, // Should handle high precision
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := testutils.SetupMockDatabase(t)
			defer db.Close()

			// Setup basic mock expectations for successful operations
			if !tt.expectError {
				rows := sqlmock.NewRows([]string{"id", "account_id", "balance", "created_at", "updated_at"}).
					AddRow(1, sqlmock.AnyArg(), sqlmock.AnyArg(), time.Now(), time.Now())
				mock.ExpectQuery(`INSERT INTO accounts`).
					WillReturnRows(rows)
			}

			repo := repository.NewRepository(db)
			err := tt.operation(repo)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// MockPQError implements error interface for PostgreSQL errors
type MockPQError struct {
	Code string
}

func (e *MockPQError) Error() string {
	return "mock pq error"
}

// Ensure MockPQError satisfies the driver.Valuer interface requirements
var _ driver.Valuer = (*MockPQError)(nil)

func (e *MockPQError) Value() (driver.Value, error) {
	return e.Code, nil
}