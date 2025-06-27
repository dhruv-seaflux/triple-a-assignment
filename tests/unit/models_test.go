package unit

import (
	"encoding/json"
	"internal-transfers/internal/models"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccount_JSONSerialization tests JSON serialization/deserialization of Account model
func TestAccount_JSONSerialization(t *testing.T) {
	tests := []struct {
		name    string
		account models.Account
	}{
		{
			name: "Valid account with positive balance",
			account: models.Account{
				ID:        1,
				AccountID: 123,
				Balance:   decimal.NewFromFloat(1000.50),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			name: "Account with zero balance",
			account: models.Account{
				ID:        2,
				AccountID: 456,
				Balance:   decimal.Zero,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			name: "Account with high precision balance",
			account: models.Account{
				ID:        3,
				AccountID: 789,
				Balance:   decimal.RequireFromString("12345.67890"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test serialization
			jsonData, err := json.Marshal(tt.account)
			require.NoError(t, err, "Failed to marshal account to JSON")

			// Test deserialization
			var unmarshaled models.Account
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err, "Failed to unmarshal account from JSON")

			// Verify fields (note: ID should not be in JSON due to json:"-" tag)
			assert.Equal(t, tt.account.AccountID, unmarshaled.AccountID)
			assert.True(t, tt.account.Balance.Equal(unmarshaled.Balance))
		})
	}
}

// TestAccountCreationRequest_Validation tests validation of AccountCreationRequest
func TestAccountCreationRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request models.AccountCreationRequest
		valid   bool
	}{
		{
			name: "Valid account creation request",
			request: models.AccountCreationRequest{
				AccountID:      123,
				InitialBalance: "1000.50",
			},
			valid: true,
		},
		{
			name: "Valid account with zero balance",
			request: models.AccountCreationRequest{
				AccountID:      456,
				InitialBalance: "0.00",
			},
			valid: true,
		},
		{
			name: "Valid account with high precision",
			request: models.AccountCreationRequest{
				AccountID:      789,
				InitialBalance: "12345.67890",
			},
			valid: true,
		},
		{
			name: "Invalid negative account ID",
			request: models.AccountCreationRequest{
				AccountID:      -1,
				InitialBalance: "1000.00",
			},
			valid: false,
		},
		{
			name: "Invalid zero account ID",
			request: models.AccountCreationRequest{
				AccountID:      0,
				InitialBalance: "1000.00",
			},
			valid: false,
		},
		{
			name: "Invalid empty balance",
			request: models.AccountCreationRequest{
				AccountID:      123,
				InitialBalance: "",
			},
			valid: false,
		},
		{
			name: "Invalid non-numeric balance",
			request: models.AccountCreationRequest{
				AccountID:      123,
				InitialBalance: "not-a-number",
			},
			valid: false,
		},
		{
			name: "Invalid negative balance",
			request: models.AccountCreationRequest{
				AccountID:      123,
				InitialBalance: "-100.00",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := json.Marshal(tt.request)
			require.NoError(t, err, "Failed to marshal request to JSON")

			// Test JSON deserialization
			var unmarshaled models.AccountCreationRequest
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err, "Failed to unmarshal request from JSON")

			// Verify basic field equality
			assert.Equal(t, tt.request.AccountID, unmarshaled.AccountID)
			assert.Equal(t, tt.request.InitialBalance, unmarshaled.InitialBalance)

			// Test validation logic
			isValid := validateAccountCreationRequest(tt.request)
			assert.Equal(t, tt.valid, isValid, "Validation result mismatch")
		})
	}
}

// TestTransactionRequest_Validation tests validation of TransactionRequest
func TestTransactionRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request models.TransactionRequest
		valid   bool
	}{
		{
			name: "Valid transaction request",
			request: models.TransactionRequest{
				SourceAccountID:      123,
				DestinationAccountID: 456,
				Amount:               "100.50",
			},
			valid: true,
		},
		{
			name: "Valid high precision amount",
			request: models.TransactionRequest{
				SourceAccountID:      123,
				DestinationAccountID: 456,
				Amount:               "999.99999",
			},
			valid: true,
		},
		{
			name: "Invalid same account transfer",
			request: models.TransactionRequest{
				SourceAccountID:      123,
				DestinationAccountID: 123,
				Amount:               "100.00",
			},
			valid: false,
		},
		{
			name: "Invalid negative source account",
			request: models.TransactionRequest{
				SourceAccountID:      -1,
				DestinationAccountID: 456,
				Amount:               "100.00",
			},
			valid: false,
		},
		{
			name: "Invalid zero destination account",
			request: models.TransactionRequest{
				SourceAccountID:      123,
				DestinationAccountID: 0,
				Amount:               "100.00",
			},
			valid: false,
		},
		{
			name: "Invalid empty amount",
			request: models.TransactionRequest{
				SourceAccountID:      123,
				DestinationAccountID: 456,
				Amount:               "",
			},
			valid: false,
		},
		{
			name: "Invalid negative amount",
			request: models.TransactionRequest{
				SourceAccountID:      123,
				DestinationAccountID: 456,
				Amount:               "-100.00",
			},
			valid: false,
		},
		{
			name: "Invalid zero amount",
			request: models.TransactionRequest{
				SourceAccountID:      123,
				DestinationAccountID: 456,
				Amount:               "0.00",
			},
			valid: false,
		},
		{
			name: "Invalid non-numeric amount",
			request: models.TransactionRequest{
				SourceAccountID:      123,
				DestinationAccountID: 456,
				Amount:               "invalid-amount",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := json.Marshal(tt.request)
			require.NoError(t, err, "Failed to marshal request to JSON")

			// Test JSON deserialization
			var unmarshaled models.TransactionRequest
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err, "Failed to unmarshal request from JSON")

			// Verify field equality
			assert.Equal(t, tt.request.SourceAccountID, unmarshaled.SourceAccountID)
			assert.Equal(t, tt.request.DestinationAccountID, unmarshaled.DestinationAccountID)
			assert.Equal(t, tt.request.Amount, unmarshaled.Amount)

			// Test validation logic
			isValid := validateTransactionRequest(tt.request)
			assert.Equal(t, tt.valid, isValid, "Validation result mismatch")
		})
	}
}

// TestAccountQueryResponse_JSONFormat tests the exact JSON format for account query response
func TestAccountQueryResponse_JSONFormat(t *testing.T) {
	response := models.AccountQueryResponse{
		AccountID: 123,
		Balance:   "1000.50000",
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err, "Failed to marshal response")

	expected := `{"account_id":123,"balance":"1000.50000"}`
	assert.JSONEq(t, expected, string(jsonData), "JSON format mismatch")
}

// TestTransactionResponse_JSONFormat tests the exact JSON format for transaction response
func TestTransactionResponse_JSONFormat(t *testing.T) {
	response := models.TransactionResponse{
		TransactionID: 42,
		Status:        "completed",
		Message:       "Transaction processed successfully",
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err, "Failed to marshal response")

	expected := `{
		"transaction_id": 42,
		"status": "completed",
		"message": "Transaction processed successfully"
	}`
	assert.JSONEq(t, expected, string(jsonData), "JSON format mismatch")
}

// Helper function to validate AccountCreationRequest
func validateAccountCreationRequest(req models.AccountCreationRequest) bool {
	if req.AccountID <= 0 {
		return false
	}

	if req.InitialBalance == "" {
		return false
	}

	amount, err := decimal.NewFromString(req.InitialBalance)
	if err != nil {
		return false
	}

	return !amount.IsNegative()
}

// Helper function to validate TransactionRequest
func validateTransactionRequest(req models.TransactionRequest) bool {
	if req.SourceAccountID <= 0 || req.DestinationAccountID <= 0 {
		return false
	}

	if req.SourceAccountID == req.DestinationAccountID {
		return false
	}

	if req.Amount == "" {
		return false
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return false
	}

	return amount.IsPositive()
}