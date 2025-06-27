package models

import (
	"time"
	"github.com/shopspring/decimal"
)

type Account struct {
	ID        int             `json:"-" db:"id"`
	AccountID int             `json:"account_id" db:"account_id"`
	Balance   decimal.Decimal `json:"balance" db:"balance"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

type Transaction struct {
	ID            int             `json:"id" db:"id"`
	FromAccountID *int            `json:"from_account_id" db:"from_account_id"`
	ToAccountID   *int            `json:"to_account_id" db:"to_account_id"`
	Amount        decimal.Decimal `json:"amount" db:"amount"`
	Description   string          `json:"description" db:"description"`
	Status        string          `json:"status" db:"status"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

type AccountCreationRequest struct {
	AccountID      int    `json:"account_id"`
	InitialBalance string `json:"initial_balance"`
}

type AccountQueryResponse struct {
	AccountID int    `json:"account_id"`
	Balance   string `json:"balance"`
}

type TransactionRequest struct {
	SourceAccountID      int    `json:"source_account_id"`
	DestinationAccountID int    `json:"destination_account_id"`
	Amount               string `json:"amount"`
}

type TransactionResponse struct {
	TransactionID int    `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

type TransferRequest struct {
	FromAccountNumber string  `json:"from_account_number"`
	ToAccountNumber   string  `json:"to_account_number"`
	Amount            float64 `json:"amount"`
	Description       string  `json:"description"`
}

type TransferResponse struct {
	TransactionID int    `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}