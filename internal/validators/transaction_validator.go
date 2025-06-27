package validators

import (
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// TransactionValidator handles validation for transaction-related operations
type TransactionValidator struct {
	validator *validator.Validate
}

// NewTransactionValidator creates a new transaction validator instance
func NewTransactionValidator() *TransactionValidator {
	v := validator.New()
	
	// Register custom validation rules
	v.RegisterValidation("decimal", isValidDecimal)
	v.RegisterValidation("positive", isPositive)
	
	return &TransactionValidator{validator: v}
}

// ValidateTransactionRequest validates a complete transaction request
func (tv *TransactionValidator) ValidateTransactionRequest(sourceAccountID, destinationAccountID int, amount string) error {
	// Validate source account ID
	if err := tv.ValidateAccountID(sourceAccountID, "source_account_id"); err != nil {
		return err
	}
	
	// Validate destination account ID
	if err := tv.ValidateAccountID(destinationAccountID, "destination_account_id"); err != nil {
		return err
	}
	
	// Validate that source and destination are different
	if sourceAccountID == destinationAccountID {
		return &ValidationError{
			Field:   "accounts", 
			Message: "source and destination accounts cannot be the same",
		}
	}
	
	// Validate amount
	if err := tv.ValidateAmount(amount); err != nil {
		return err
	}
	
	return nil
}

// ValidateAccountID validates if account ID is a positive integer
func (tv *TransactionValidator) ValidateAccountID(accountID int, fieldName string) error {
	if accountID <= 0 {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " must be a positive integer",
		}
	}
	return nil
}

// ValidateAmount validates if amount is a valid positive decimal
func (tv *TransactionValidator) ValidateAmount(amount string) error {
	if amount == "" {
		return &ValidationError{Field: "amount", Message: "amount is required"}
	}
	
	// Check if it's a valid decimal
	amountDecimal, err := decimal.NewFromString(amount)
	if err != nil {
		return &ValidationError{Field: "amount", Message: "amount must be a valid decimal number"}
	}
	
	// Check if it's positive
	if amountDecimal.LessThanOrEqual(decimal.Zero) {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}
	
	// Check for reasonable precision (up to 5 decimal places)
	if amountDecimal.Exponent() < -5 {
		return &ValidationError{Field: "amount", Message: "amount supports up to 5 decimal places"}
	}
	
	return nil
}