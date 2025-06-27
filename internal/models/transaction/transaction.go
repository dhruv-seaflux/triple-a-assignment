package transaction

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Transaction represents the transaction database model
type Transaction struct {
	ID            uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	FromAccountID *uint           `gorm:"index" json:"from_account_id"`
	ToAccountID   *uint           `gorm:"index" json:"to_account_id"`
	Amount        decimal.Decimal `gorm:"type:decimal(15,5);not null" json:"amount"`
	Description   string          `gorm:"size:255" json:"description"`
	Status        string          `gorm:"size:20;default:pending" json:"status"`
	CreatedAt     time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName specifies the table name for GORM
func (Transaction) TableName() string {
	return "transactions"
}

// SubmitTransactionRequest represents the request payload for submitting a transaction
type SubmitTransactionRequest struct {
	SourceAccountID      int    `json:"source_account_id" binding:"required,min=1" validate:"required,min=1"`
	DestinationAccountID int    `json:"destination_account_id" binding:"required,min=1" validate:"required,min=1"`
	Amount               string `json:"amount" binding:"required" validate:"required,decimal,positive"`
}

// TransactionResponse represents the response for transaction operations
type TransactionResponse struct {
	TransactionID int    `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

// TransactionStatus constants
const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)