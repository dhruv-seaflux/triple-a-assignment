package validators

import (
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// AccountValidator handles validation for account-related operations
type AccountValidator struct {
	validator *validator.Validate
}

// NewAccountValidator creates a new account validator instance
func NewAccountValidator() *AccountValidator {
	v := validator.New()
	
	// Register custom validation rules
	v.RegisterValidation("decimal", isValidDecimal)
	v.RegisterValidation("positive", isPositive)
	v.RegisterValidation("non_negative", isNonNegative)
	
	return &AccountValidator{validator: v}
}

// ValidateAccountID validates if account ID is a positive integer
func (av *AccountValidator) ValidateAccountID(accountID int) error {
	if accountID <= 0 {
		return &ValidationError{Field: "account_id", Message: "account_id must be a positive integer"}
	}
	return nil
}

// ValidateInitialBalance validates if initial balance is a valid non-negative decimal
func (av *AccountValidator) ValidateInitialBalance(balance string) error {
	if balance == "" {
		return &ValidationError{Field: "initial_balance", Message: "initial_balance is required"}
	}
	
	// Check if it's a valid decimal
	_, err := decimal.NewFromString(balance)
	if err != nil {
		return &ValidationError{Field: "initial_balance", Message: "initial_balance must be a valid decimal number"}
	}
	
	// Check if it's non-negative
	balanceDecimal, _ := decimal.NewFromString(balance)
	if balanceDecimal.LessThan(decimal.Zero) {
		return &ValidationError{Field: "initial_balance", Message: "initial_balance cannot be negative"}
	}
	
	return nil
}

// isValidDecimal checks if a string represents a valid decimal number
func isValidDecimal(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	_, err := decimal.NewFromString(value)
	return err == nil
}

// isPositive checks if a decimal string represents a positive number
func isPositive(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return false
	}
	
	decimalValue, err := decimal.NewFromString(value)
	if err != nil {
		return false
	}
	
	return decimalValue.GreaterThan(decimal.Zero)
}

// isNonNegative checks if a decimal string represents a non-negative number
func isNonNegative(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return false
	}
	
	decimalValue, err := decimal.NewFromString(value)
	if err != nil {
		return false
	}
	
	return decimalValue.GreaterThanOrEqual(decimal.Zero)
}