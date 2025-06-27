package account

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Account represents the account database model
type Account struct {
	ID        uint            `gorm:"primaryKey;autoIncrement" json:"-"`
	AccountID int             `gorm:"uniqueIndex;not null" json:"account_id"`
	Balance   decimal.Decimal `gorm:"type:decimal(15,5);not null;default:0" json:"balance"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName specifies the table name for GORM
func (Account) TableName() string {
	return "accounts"
}

// CreateAccountRequest represents the request payload for creating an account
type CreateAccountRequest struct {
	AccountID      int    `json:"account_id" binding:"required,min=1" validate:"required,min=1"`
	InitialBalance string `json:"initial_balance" binding:"required" validate:"required,decimal"`
}

// AccountQueryResponse represents the response for account queries
type AccountQueryResponse struct {
	AccountID int    `json:"account_id"`
	Balance   string `json:"balance"`
}