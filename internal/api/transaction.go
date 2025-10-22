// Package api provides HTTP handlers for the transaction service.
// It implements RESTful endpoints for creating, reading, updating, and listing transactions.
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/abadojack/gapstack/internal/db"
	"github.com/abadojack/gapstack/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	// defaultPageSize is the default number of transactions to return per page
	defaultPageSize = 10
)

// Handler contains the HTTP handlers for transaction operations.
// It holds a reference to the database interface for data persistence.
type Handler struct {
	DB db.DB
}

// NewHandler creates a new Handler instance with the provided database interface.
func NewHandler(db db.DB) *Handler {
	return &Handler{
		DB: db,
	}
}

// RegisterRoutes sets up all the HTTP routes for the transaction API.
// It registers endpoints for CRUD operations on transactions.
func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/transactions", h.CreateTransaction).Methods("POST")
	r.HandleFunc("/transactions", h.ListTransactions).Methods("GET")
	r.HandleFunc("/transactions/{id}", h.GetTransaction).Methods("GET")
	r.HandleFunc("/transactions/{id}", h.UpdateTransaction).Methods("PUT")
}

// CreateTransaction handles POST requests to create a new transaction.
// It validates the input, sets the default status to pending, and stores the transaction.
func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var transaction models.Transaction

	// Decode request body into transaction struct
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		log.Println(err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Input validation
	if err := validateTransaction(transaction); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Default status = pending
	transaction.Status = models.StatusPending
	transaction.ID = uuid.NewString()
	transaction.CreatedAt = time.Now()

	// Store transaction in database
	if err := h.DB.CreateTransaction(transaction); err != nil {
		log.Println(err)
		http.Error(w, "error creating transaction", http.StatusInternalServerError)
		return
	}

	// Respond with created transaction
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(transaction); err != nil {
		log.Println(err)
		http.Error(w, "error encoding transaction", http.StatusInternalServerError)
		return
	}
}

// GetTransaction handles GET requests to retrieve a single transaction by ID.
// It extracts the ID from the URL path and returns the transaction data.
func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	// Extract transaction ID from URL path
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "missing transaction id", http.StatusBadRequest)
		return
	}

	// Retrieve transaction from database
	transaction, err := h.DB.GetTransaction(id)
	if err != nil {
		http.Error(w, "error getting transaction", http.StatusInternalServerError)
		return
	}

	// Return transaction data
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(transaction); err != nil {
		http.Error(w, "error encoding transaction", http.StatusInternalServerError)
		return
	}
}

// ListTransactions handles GET requests to retrieve a paginated list of transactions.
// It supports query parameters for pagination: page and page_size.
func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	// Parse pagination query params
	pageParam := r.URL.Query().Get("page")
	pageSizeParam := r.URL.Query().Get("page_size")

	// Parse page number with validation
	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		log.Println(err)
		page = 1
	}

	// Parse page size with validation
	pageSize, err := strconv.Atoi(pageSizeParam)
	if err != nil || pageSize < 1 {
		log.Println(err)
		pageSize = defaultPageSize // default page size
	}

	// Calculate offset for database query
	offset := (page - 1) * pageSize

	// Retrieve transactions from database
	transactions, err := h.DB.GetAllTransactions(pageSize, offset)
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting transactions", http.StatusInternalServerError)
		return
	}

	// Build paginated response
	response := map[string]interface{}{
		"page":         page,
		"page_size":    len(transactions),
		"transactions": transactions,
	}

	// Return paginated response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

// updateRequest represents the request body for updating a transaction status.
type updateRequest struct {
	Status models.Status `json:"status"`
}

// UpdateTransaction handles PUT requests to update a transaction's status.
// Only completed and failed statuses are allowed for updates.
func (h *Handler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	// Extract transaction ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		log.Println("missing transaction id")
		http.Error(w, "missing transaction id", http.StatusBadRequest)
		return
	}

	// Parse JSON body
	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate status - only allow completed or failed
	if (req.Status != models.StatusFailed) && (req.Status != models.StatusCompleted) {
		log.Println("invalid status requested")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Update transaction in database
	if err := h.DB.UpdateTransaction(id, req.Status); err != nil {
		log.Println(err)
		http.Error(w, "error updating transaction", http.StatusInternalServerError)
		return
	}

	// Return success response (no content)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// validateTransaction performs comprehensive input validation on transaction data.
// It checks all required fields, validates formats, and ensures business rules are followed.
func validateTransaction(transaction models.Transaction) error {
	var errors []string

	// Validate amount
	if transaction.Amount <= 0 {
		errors = append(errors, "amount must be greater than 0")
	}
	if transaction.Amount > 99999999.99 {
		errors = append(errors, "amount must be less than 100,000,000")
	}

	// Validate currency
	if transaction.Currency == "" {
		errors = append(errors, "currency is required")
	} else if !isValidCurrency(transaction.Currency) {
		errors = append(errors, "currency must be a valid 3-letter ISO code (e.g., USD, EUR, GBP)")
	}

	// Validate sender
	if transaction.Sender == "" {
		errors = append(errors, "sender is required")
	} else if len(transaction.Sender) > 255 {
		errors = append(errors, "sender must be 255 characters or less")
	}

	// Validate receiver
	if transaction.Receiver == "" {
		errors = append(errors, "receiver is required")
	} else if len(transaction.Receiver) > 255 {
		errors = append(errors, "receiver must be 255 characters or less")
	}

	// Check if sender and receiver are different
	if transaction.Sender == transaction.Receiver {
		errors = append(errors, "sender and receiver must be different")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// isValidCurrency checks if the currency code is valid according to ISO 4217 standards.
// It validates that the currency is a 3-letter uppercase code from a predefined list.
func isValidCurrency(currency string) bool {
	validCurrencies := map[string]bool{
		"USD": true, "EUR": true, "GBP": true, "JPY": true, "CAD": true,
		"AUD": true, "CHF": true, "CNY": true, "SEK": true, "NZD": true,
		"MXN": true, "SGD": true, "HKD": true, "NOK": true, "TRY": true,
		"RUB": true, "INR": true, "BRL": true, "ZAR": true, "KRW": true,
		"KES": true,
	}

	// Check if it's a 3-letter uppercase code
	if len(currency) != 3 {
		return false
	}

	return validCurrencies[strings.ToUpper(currency)]
}
