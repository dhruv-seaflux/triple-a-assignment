package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"internal-transfers/internal/models/account"
	"internal-transfers/internal/models/transaction"
	"internal-transfers/internal/repository"
	"internal-transfers/internal/responses"
	"internal-transfers/internal/validators"
)

// Handler handles HTTP requests using Gin framework
type Handler struct {
	repo              *repository.Repository
	accountValidator  *validators.AccountValidator
	transactionValidator *validators.TransactionValidator
}

// NewHandler creates a new Handler instance
func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{
		repo:                 repo,
		accountValidator:     validators.NewAccountValidator(),
		transactionValidator: validators.NewTransactionValidator(),
	}
}

// CreateAccountHandler handles POST /accounts
func (h *Handler) CreateAccountHandler(c *gin.Context) {
	var req account.CreateAccountRequest
	
	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request format")
		return
	}
	
	// Validate account ID
	if err := h.accountValidator.ValidateAccountID(req.AccountID); err != nil {
		if validators.IsValidationError(err) {
			responses.SendError(c, http.StatusBadRequest, err.Error())
			return
		}
		responses.SendError(c, http.StatusInternalServerError, "Validation error")
		return
	}
	
	// Validate initial balance
	if err := h.accountValidator.ValidateInitialBalance(req.InitialBalance); err != nil {
		if validators.IsValidationError(err) {
			responses.SendError(c, http.StatusBadRequest, err.Error())
			return
		}
		responses.SendError(c, http.StatusInternalServerError, "Validation error")
		return
	}
	
	// Convert initial balance to decimal
	initialBalance, err := decimal.NewFromString(req.InitialBalance)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid initial balance format")
		return
	}
	
	// Create account
	err = h.repo.CreateAccount(req.AccountID, initialBalance)
	if err != nil {
		if err.Error() == "account already exists" {
			responses.SendError(c, http.StatusConflict, "account already exists")
			return
		}
		responses.SendError(c, http.StatusInternalServerError, "Failed to create account")
		return
	}
	
	responses.SendSuccess(c, http.StatusCreated, "Account created successfully")
}

// GetAccountBalanceHandler handles GET /accounts/{account_id}
func (h *Handler) GetAccountBalanceHandler(c *gin.Context) {
	// Get account ID from path parameter
	accountIDStr := c.Param("account_id")
	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid account ID format")
		return
	}
	
	// Validate account ID
	if err := h.accountValidator.ValidateAccountID(accountID); err != nil {
		if validators.IsValidationError(err) {
			responses.SendError(c, http.StatusBadRequest, err.Error())
			return
		}
		responses.SendError(c, http.StatusInternalServerError, "Validation error")
		return
	}
	
	// Get account
	acc, err := h.repo.GetAccountByID(accountID)
	if err != nil {
		if err.Error() == "account not found" {
			responses.SendError(c, http.StatusNotFound, "account not found")
			return
		}
		responses.SendError(c, http.StatusInternalServerError, "Failed to retrieve account")
		return
	}
	
	responses.SendAccountResponse(c, acc.AccountID, acc.Balance.String())
}

// SubmitTransactionHandler handles POST /submit
func (h *Handler) SubmitTransactionHandler(c *gin.Context) {
	var req transaction.SubmitTransactionRequest
	
	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request format")
		return
	}
	
	// Validate transaction request
	err := h.transactionValidator.ValidateTransactionRequest(
		req.SourceAccountID,
		req.DestinationAccountID,
		req.Amount,
	)
	if err != nil {
		if validators.IsValidationError(err) {
			responses.SendError(c, http.StatusBadRequest, err.Error())
			return
		}
		responses.SendError(c, http.StatusInternalServerError, "Validation error")
		return
	}
	
	// Convert amount to decimal
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid amount format")
		return
	}
	
	// Process transaction with GORM database transaction
	trans, err := h.repo.ProcessTransaction(req.SourceAccountID, req.DestinationAccountID, amount)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case contains(errorMsg, "account") && contains(errorMsg, "not found"):
			responses.SendError(c, http.StatusNotFound, "One or both accounts not found")
		case contains(errorMsg, "insufficient balance"):
			responses.SendError(c, http.StatusBadRequest, "Insufficient funds: "+errorMsg)
		case contains(errorMsg, "inactive"):
			responses.SendError(c, http.StatusBadRequest, "Account is inactive")
		case contains(errorMsg, "same"):
			responses.SendError(c, http.StatusBadRequest, "Source and destination accounts cannot be the same")
		case contains(errorMsg, "negative balance"):
			responses.SendError(c, http.StatusBadRequest, "Transaction would result in negative balance")
		case contains(errorMsg, "positive"):
			responses.SendError(c, http.StatusBadRequest, "Transfer amount must be positive")
		default:
			responses.SendError(c, http.StatusInternalServerError, "Failed to process transaction")
		}
		return
	}
	
	responses.SendTransactionResponse(c, http.StatusCreated, int(trans.ID), trans.Status, "Transaction processed successfully")
}

// HealthCheckHandler handles GET /health
func (h *Handler) HealthCheckHandler(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}