package testutils

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"internal-transfers/internal/config"
	"internal-transfers/internal/handlers"
	"internal-transfers/internal/repository"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

// TestConfig holds test configuration
type TestConfig struct {
	DatabaseURL string
	TestDBName  string
}

// LoadTestConfig loads configuration for testing
func LoadTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseURL: getEnv("TEST_DATABASE_URL", "postgres://user:password@localhost:5432/internal_transfers_test?sslmode=disable"),
		TestDBName:  "internal_transfers_test",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupTestDatabase creates a test database connection
func SetupTestDatabase(t *testing.T) *sql.DB {
	cfg := LoadTestConfig()
	db, err := config.ConnectDB(cfg.DatabaseURL)
	require.NoError(t, err, "Failed to connect to test database")
	return db
}

// SetupMockDatabase creates a mock database for unit testing
func SetupMockDatabase(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")
	return db, mock
}

// SetupTestServer creates a test HTTP server with all routes
func SetupTestServer(repo *repository.Repository) *httptest.Server {
	handler := handlers.NewHandler(repo)
	r := mux.NewRouter()

	// Register all routes
	r.HandleFunc("/accounts", handler.CreateAccount).Methods("POST")
	r.HandleFunc("/accounts/{accountID}", handler.GetAccountBalance).Methods("GET")
	r.HandleFunc("/transactions", handler.SubmitTransaction).Methods("POST")
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	return httptest.NewServer(r)
}

// MakeJSONRequest makes an HTTP request with JSON payload
func MakeJSONRequest(t *testing.T, method, url string, payload interface{}) *http.Response {
	var body bytes.Buffer
	if payload != nil {
		err := json.NewEncoder(&body).Encode(payload)
		require.NoError(t, err, "Failed to encode JSON payload")
	}

	req, err := http.NewRequest(method, url, &body)
	require.NoError(t, err, "Failed to create HTTP request")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to make HTTP request")

	return resp
}

// DecodeJSONResponse decodes JSON response body into target struct
func DecodeJSONResponse(t *testing.T, resp *http.Response, target interface{}) {
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(target)
	require.NoError(t, err, "Failed to decode JSON response")
}

// CleanupTestDatabase cleans up test database
func CleanupTestDatabase(t *testing.T, db *sql.DB) {
	queries := []string{
		"DELETE FROM transactions",
		"DELETE FROM accounts",
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			t.Logf("Warning: Failed to cleanup table with query %s: %v", query, err)
		}
	}
}

// CreateTestAccount creates a test account for testing purposes
func CreateTestAccount(t *testing.T, repo *repository.Repository, accountID int, balance string) {
	decimalBalance, err := decimal.NewFromString(balance)
	require.NoError(t, err, "Invalid test balance")

	_, err = repo.CreateAccount(accountID, decimalBalance)
	require.NoError(t, err, "Failed to create test account")
}

// AssertAccountBalance verifies account balance matches expected value
func AssertAccountBalance(t *testing.T, repo *repository.Repository, accountID int, expectedBalance string) {
	account, err := repo.GetAccountByID(accountID)
	require.NoError(t, err, "Failed to get account")

	expected, err := decimal.NewFromString(expectedBalance)
	require.NoError(t, err, "Invalid expected balance")

	require.True(t, account.Balance.Equal(expected), 
		fmt.Sprintf("Balance mismatch: expected %s, got %s", expectedBalance, account.Balance.String()))
}

// TestAccount represents a test account structure
type TestAccount struct {
	AccountID int
	Balance   string
}

// TestTransaction represents a test transaction structure
type TestTransaction struct {
	SourceAccountID      int
	DestinationAccountID int
	Amount               string
}

// GenerateTestAccounts creates multiple test accounts
func GenerateTestAccounts() []TestAccount {
	return []TestAccount{
		{AccountID: 1001, Balance: "1000.00"},
		{AccountID: 1002, Balance: "2000.50"},
		{AccountID: 1003, Balance: "500.75"},
		{AccountID: 1004, Balance: "0.00"},
	}
}

// GenerateTestTransactions creates multiple test transactions
func GenerateTestTransactions() []TestTransaction {
	return []TestTransaction{
		{SourceAccountID: 1001, DestinationAccountID: 1002, Amount: "100.25"},
		{SourceAccountID: 1002, DestinationAccountID: 1003, Amount: "250.00"},
		{SourceAccountID: 1003, DestinationAccountID: 1004, Amount: "50.75"},
	}
}