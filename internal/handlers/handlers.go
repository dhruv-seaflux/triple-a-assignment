package handlers

import (
	"encoding/json"
	"errors"
	"internal-transfers/internal/models"
	"internal-transfers/internal/repository"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

type Handler struct {
	repo *repository.Repository
}

func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

// CreateAccount handles POST /accounts endpoint
// Accepts JSON with account_id (int) and initial_balance (string)
// Returns empty response on success or error on failure
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req models.AccountCreationRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid JSON payload: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate account_id
	if req.AccountID <= 0 {
		log.Printf("Invalid account_id: %d", req.AccountID)
		http.Error(w, "account_id must be a positive integer", http.StatusBadRequest)
		return
	}

	// Validate initial_balance
	if req.InitialBalance == "" {
		log.Printf("Empty initial_balance provided")
		http.Error(w, "initial_balance is required", http.StatusBadRequest)
		return
	}

	initialBalance, err := decimal.NewFromString(req.InitialBalance)
	if err != nil {
		log.Printf("Invalid initial_balance format: %s, error: %v", req.InitialBalance, err)
		http.Error(w, "initial_balance must be a valid decimal number", http.StatusBadRequest)
		return
	}

	// Create account in repository
	_, err = h.repo.CreateAccount(req.AccountID, initialBalance)
	if err != nil {
		if errors.Is(err, repository.ErrAccountAlreadyExists) {
			log.Printf("Account already exists: %d", req.AccountID)
			http.Error(w, "account already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, repository.ErrInvalidAmount) {
			log.Printf("Invalid amount: %s", req.InitialBalance)
			http.Error(w, "initial_balance cannot be negative", http.StatusBadRequest)
			return
		}
		
		log.Printf("Failed to create account: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Return empty response with 201 status on success
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetAccountBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountIDStr := vars["accountID"]

	if accountIDStr == "" {
		log.Printf("Missing account_id in URL path")
		http.Error(w, "account_id is required", http.StatusBadRequest)
		return
	}

	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil || accountID <= 0 {
		log.Printf("Invalid account_id: %s", accountIDStr)
		http.Error(w, "account_id must be a positive integer", http.StatusBadRequest)
		return
	}

	account, err := h.repo.GetAccountByID(accountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			log.Printf("Account not found: %d", accountID)
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}
		
		log.Printf("Failed to get account: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Create response in the exact format specified
	response := models.AccountQueryResponse{
		AccountID: account.AccountID,
		Balance:   account.Balance.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// SubmitTransaction handles POST /transactions endpoint
// Accepts JSON with source_account_id, destination_account_id, and amount
// Processes the transaction and updates account balances atomically
func (h *Handler) SubmitTransaction(w http.ResponseWriter, r *http.Request) {
	var req models.TransactionRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid JSON payload: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate source_account_id
	if req.SourceAccountID <= 0 {
		log.Printf("Invalid source_account_id: %d", req.SourceAccountID)
		http.Error(w, "source_account_id must be a positive integer", http.StatusBadRequest)
		return
	}

	// Validate destination_account_id
	if req.DestinationAccountID <= 0 {
		log.Printf("Invalid destination_account_id: %d", req.DestinationAccountID)
		http.Error(w, "destination_account_id must be a positive integer", http.StatusBadRequest)
		return
	}

	// Validate that source and destination are different
	if req.SourceAccountID == req.DestinationAccountID {
		log.Printf("Source and destination accounts are the same: %d", req.SourceAccountID)
		http.Error(w, "source and destination accounts cannot be the same", http.StatusBadRequest)
		return
	}

	// Validate amount
	if req.Amount == "" {
		log.Printf("Empty amount provided")
		http.Error(w, "amount is required", http.StatusBadRequest)
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		log.Printf("Invalid amount format: %s, error: %v", req.Amount, err)
		http.Error(w, "amount must be a valid decimal number", http.StatusBadRequest)
		return
	}

	if amount.IsNegative() || amount.IsZero() {
		log.Printf("Invalid amount value: %s", req.Amount)
		http.Error(w, "amount must be positive", http.StatusBadRequest)
		return
	}

	// Process the transaction
	transaction, err := h.repo.ProcessTransaction(req.SourceAccountID, req.DestinationAccountID, amount)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			log.Printf("Account not found during transaction: %v", err)
			http.Error(w, "one or both accounts not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrInsufficientBalance) {
			log.Printf("Insufficient balance for transaction: %v", err)
			http.Error(w, "insufficient balance", http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrInvalidAmount) {
			log.Printf("Invalid amount for transaction: %v", err)
			http.Error(w, "invalid amount", http.StatusBadRequest)
			return
		}
		
		log.Printf("Failed to process transaction: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Create response
	response := models.TransactionResponse{
		TransactionID: transaction.ID,
		Status:        transaction.Status,
		Message:       "Transaction processed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req models.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid JSON payload: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		log.Printf("Invalid amount: %f", req.Amount)
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	if req.FromAccountNumber == "" || req.ToAccountNumber == "" {
		log.Printf("Missing account numbers: from=%s, to=%s", req.FromAccountNumber, req.ToAccountNumber)
		http.Error(w, "Both account numbers are required", http.StatusBadRequest)
		return
	}

	// Note: This still uses the old account number system and needs to be updated
	// when transfer functionality is implemented to use account IDs
	
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, "Transfer functionality not yet implemented with new account system", http.StatusNotImplemented)
}