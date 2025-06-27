package fixtures

import (
	"internal-transfers/internal/models"
	"time"

	"github.com/shopspring/decimal"
)

// TestAccounts provides sample account data for testing
var TestAccounts = []models.Account{
	{
		ID:        1,
		AccountID: 1001,
		Balance:   decimal.NewFromFloat(1000.00),
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	},
	{
		ID:        2,
		AccountID: 1002,
		Balance:   decimal.NewFromFloat(2500.50),
		CreatedAt: time.Now().Add(-12 * time.Hour),
		UpdatedAt: time.Now().Add(-12 * time.Hour),
	},
	{
		ID:        3,
		AccountID: 1003,
		Balance:   decimal.NewFromFloat(500.75),
		CreatedAt: time.Now().Add(-6 * time.Hour),
		UpdatedAt: time.Now().Add(-6 * time.Hour),
	},
	{
		ID:        4,
		AccountID: 1004,
		Balance:   decimal.Zero,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	},
}

// TestTransactions provides sample transaction data for testing
var TestTransactions = []models.Transaction{
	{
		ID:            1,
		FromAccountID: &[]int{1}[0],
		ToAccountID:   &[]int{2}[0],
		Amount:        decimal.NewFromFloat(100.25),
		Description:   "Test transfer 1",
		Status:        "completed",
		CreatedAt:     time.Now().Add(-2 * time.Hour),
		UpdatedAt:     time.Now().Add(-2 * time.Hour),
	},
	{
		ID:            2,
		FromAccountID: &[]int{2}[0],
		ToAccountID:   &[]int{3}[0],
		Amount:        decimal.NewFromFloat(250.00),
		Description:   "Test transfer 2",
		Status:        "completed",
		CreatedAt:     time.Now().Add(-1 * time.Hour),
		UpdatedAt:     time.Now().Add(-1 * time.Hour),
	},
}

// ValidAccountCreationRequests provides valid test requests for account creation
var ValidAccountCreationRequests = []models.AccountCreationRequest{
	{
		AccountID:      2001,
		InitialBalance: "1000.00",
	},
	{
		AccountID:      2002,
		InitialBalance: "0.00",
	},
	{
		AccountID:      2003,
		InitialBalance: "999.99999",
	},
	{
		AccountID:      2004,
		InitialBalance: "12345.67890",
	},
}

// InvalidAccountCreationRequests provides invalid test requests for account creation
var InvalidAccountCreationRequests = []struct {
	Request models.AccountCreationRequest
	Error   string
}{
	{
		Request: models.AccountCreationRequest{
			AccountID:      -1,
			InitialBalance: "1000.00",
		},
		Error: "account_id must be a positive integer",
	},
	{
		Request: models.AccountCreationRequest{
			AccountID:      0,
			InitialBalance: "1000.00",
		},
		Error: "account_id must be a positive integer",
	},
	{
		Request: models.AccountCreationRequest{
			AccountID:      123,
			InitialBalance: "",
		},
		Error: "initial_balance is required",
	},
	{
		Request: models.AccountCreationRequest{
			AccountID:      123,
			InitialBalance: "not-a-number",
		},
		Error: "initial_balance must be a valid decimal number",
	},
	{
		Request: models.AccountCreationRequest{
			AccountID:      123,
			InitialBalance: "-100.00",
		},
		Error: "initial_balance cannot be negative",
	},
}

// ValidTransactionRequests provides valid test requests for transactions
var ValidTransactionRequests = []models.TransactionRequest{
	{
		SourceAccountID:      1001,
		DestinationAccountID: 1002,
		Amount:               "100.50",
	},
	{
		SourceAccountID:      1002,
		DestinationAccountID: 1003,
		Amount:               "250.25",
	},
	{
		SourceAccountID:      1001,
		DestinationAccountID: 1004,
		Amount:               "999.99999",
	},
}

// InvalidTransactionRequests provides invalid test requests for transactions
var InvalidTransactionRequests = []struct {
	Request models.TransactionRequest
	Error   string
}{
	{
		Request: models.TransactionRequest{
			SourceAccountID:      -1,
			DestinationAccountID: 1002,
			Amount:               "100.00",
		},
		Error: "source_account_id must be a positive integer",
	},
	{
		Request: models.TransactionRequest{
			SourceAccountID:      1001,
			DestinationAccountID: 0,
			Amount:               "100.00",
		},
		Error: "destination_account_id must be a positive integer",
	},
	{
		Request: models.TransactionRequest{
			SourceAccountID:      1001,
			DestinationAccountID: 1001,
			Amount:               "100.00",
		},
		Error: "source and destination accounts cannot be the same",
	},
	{
		Request: models.TransactionRequest{
			SourceAccountID:      1001,
			DestinationAccountID: 1002,
			Amount:               "",
		},
		Error: "amount is required",
	},
	{
		Request: models.TransactionRequest{
			SourceAccountID:      1001,
			DestinationAccountID: 1002,
			Amount:               "invalid-amount",
		},
		Error: "amount must be a valid decimal number",
	},
	{
		Request: models.TransactionRequest{
			SourceAccountID:      1001,
			DestinationAccountID: 1002,
			Amount:               "-100.00",
		},
		Error: "amount must be positive",
	},
	{
		Request: models.TransactionRequest{
			SourceAccountID:      1001,
			DestinationAccountID: 1002,
			Amount:               "0.00",
		},
		Error: "amount must be positive",
	},
}

// AccountQueryTestCases provides test cases for account queries
var AccountQueryTestCases = []struct {
	AccountID      int
	ShouldExist    bool
	ExpectedStatus int
	Description    string
}{
	{
		AccountID:      1001,
		ShouldExist:    true,
		ExpectedStatus: 200,
		Description:    "Query existing account",
	},
	{
		AccountID:      9999,
		ShouldExist:    false,
		ExpectedStatus: 404,
		Description:    "Query non-existent account",
	},
	{
		AccountID:      -1,
		ShouldExist:    false,
		ExpectedStatus: 400,
		Description:    "Query with invalid account ID",
	},
}

// TransactionScenarios provides complex transaction test scenarios
var TransactionScenarios = []struct {
	Name               string
	InitialAccounts    []models.Account
	Transactions       []models.TransactionRequest
	ExpectedBalances   map[int]string
	ExpectedErrors     []string
	Description        string
}{
	{
		Name: "Simple transfer between two accounts",
		InitialAccounts: []models.Account{
			{AccountID: 3001, Balance: decimal.NewFromFloat(1000.00)},
			{AccountID: 3002, Balance: decimal.NewFromFloat(500.00)},
		},
		Transactions: []models.TransactionRequest{
			{
				SourceAccountID:      3001,
				DestinationAccountID: 3002,
				Amount:               "150.50",
			},
		},
		ExpectedBalances: map[int]string{
			3001: "849.50",
			3002: "650.50",
		},
		ExpectedErrors: []string{},
		Description:    "Transfer 150.50 from account 3001 to 3002",
	},
	{
		Name: "Multiple transactions in sequence",
		InitialAccounts: []models.Account{
			{AccountID: 4001, Balance: decimal.NewFromFloat(1000.00)},
			{AccountID: 4002, Balance: decimal.NewFromFloat(500.00)},
			{AccountID: 4003, Balance: decimal.NewFromFloat(200.00)},
		},
		Transactions: []models.TransactionRequest{
			{
				SourceAccountID:      4001,
				DestinationAccountID: 4002,
				Amount:               "100.00",
			},
			{
				SourceAccountID:      4002,
				DestinationAccountID: 4003,
				Amount:               "50.00",
			},
		},
		ExpectedBalances: map[int]string{
			4001: "900.00",
			4002: "550.00",
			4003: "250.00",
		},
		ExpectedErrors: []string{},
		Description:    "Chain of transfers: 4001 -> 4002 -> 4003",
	},
	{
		Name: "Insufficient balance scenario",
		InitialAccounts: []models.Account{
			{AccountID: 5001, Balance: decimal.NewFromFloat(100.00)},
			{AccountID: 5002, Balance: decimal.NewFromFloat(500.00)},
		},
		Transactions: []models.TransactionRequest{
			{
				SourceAccountID:      5001,
				DestinationAccountID: 5002,
				Amount:               "200.00",
			},
		},
		ExpectedBalances: map[int]string{
			5001: "100.00", // Unchanged due to error
			5002: "500.00", // Unchanged due to error
		},
		ExpectedErrors: []string{"insufficient balance"},
		Description:    "Attempt to transfer more than available balance",
	},
}

// PerformanceTestData provides data for performance testing
var PerformanceTestData = struct {
	AccountCount      int
	TransactionCount  int
	ConcurrentThreads int
}{
	AccountCount:      1000,
	TransactionCount:  10000,
	ConcurrentThreads: 10,
}

// EdgeCaseTestData provides edge case test scenarios
var EdgeCaseTestData = []struct {
	Name        string
	Data        interface{}
	Description string
}{
	{
		Name: "Maximum account ID",
		Data: models.AccountCreationRequest{
			AccountID:      2147483647, // Max int32
			InitialBalance: "1000.00",
		},
		Description: "Test with maximum possible account ID",
	},
	{
		Name: "High precision decimal",
		Data: models.AccountCreationRequest{
			AccountID:      123,
			InitialBalance: "123.123456789012345",
		},
		Description: "Test with very high precision decimal",
	},
	{
		Name: "Very large amount",
		Data: models.TransactionRequest{
			SourceAccountID:      1001,
			DestinationAccountID: 1002,
			Amount:               "999999999999.99999",
		},
		Description: "Test with very large transaction amount",
	},
	{
		Name: "Minimum positive amount",
		Data: models.TransactionRequest{
			SourceAccountID:      1001,
			DestinationAccountID: 1002,
			Amount:               "0.00001",
		},
		Description: "Test with minimum positive transaction amount",
	},
}

// DatabaseTestQueries provides SQL queries for test database setup and cleanup
var DatabaseTestQueries = struct {
	CreateTables []string
	DropTables   []string
	SeedData     []string
	CleanupData  []string
}{
	CreateTables: []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id SERIAL PRIMARY KEY,
			account_id INTEGER UNIQUE NOT NULL,
			balance DECIMAL(15,5) NOT NULL DEFAULT 0.00000,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			from_account_id INTEGER REFERENCES accounts(id),
			to_account_id INTEGER REFERENCES accounts(id),
			amount DECIMAL(15,5) NOT NULL,
			description VARCHAR(255),
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_accounts_account_id ON accounts(account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_from_account ON transactions(from_account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_to_account ON transactions(to_account_id)`,
	},
	DropTables: []string{
		`DROP TABLE IF EXISTS transactions CASCADE`,
		`DROP TABLE IF EXISTS accounts CASCADE`,
	},
	SeedData: []string{
		`INSERT INTO accounts (account_id, balance) VALUES (1001, 1000.00) ON CONFLICT DO NOTHING`,
		`INSERT INTO accounts (account_id, balance) VALUES (1002, 2500.50) ON CONFLICT DO NOTHING`,
		`INSERT INTO accounts (account_id, balance) VALUES (1003, 500.75) ON CONFLICT DO NOTHING`,
		`INSERT INTO accounts (account_id, balance) VALUES (1004, 0.00) ON CONFLICT DO NOTHING`,
	},
	CleanupData: []string{
		`DELETE FROM transactions`,
		`DELETE FROM accounts`,
	},
}