package responses

import "github.com/gin-gonic/gin"

// StandardResponse represents the standard success response format
type StandardResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Error string `json:"error"`
}

// AccountResponse represents account query response
type AccountResponse struct {
	AccountID int    `json:"account_id"`
	Balance   string `json:"balance"`
}

// TransactionResponse represents transaction submission response
type TransactionResponse struct {
	TransactionID int    `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

// SendSuccess sends a standardized success response
func SendSuccess(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, StandardResponse{Message: message})
}

// SendError sends a standardized error response
func SendError(c *gin.Context, statusCode int, errorMessage string) {
	c.JSON(statusCode, ErrorResponse{Error: errorMessage})
}

// SendAccountResponse sends account query response
func SendAccountResponse(c *gin.Context, accountID int, balance string) {
	c.JSON(200, AccountResponse{
		AccountID: accountID,
		Balance:   balance,
	})
}

// SendTransactionResponse sends transaction response
func SendTransactionResponse(c *gin.Context, statusCode int, transactionID int, status, message string) {
	c.JSON(statusCode, TransactionResponse{
		TransactionID: transactionID,
		Status:        status,
		Message:       message,
	})
}