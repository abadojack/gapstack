// api/handler_test.go
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/abadojack/gapstack/internal/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDB implements the db.DB interface for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) CreateTransaction(transaction models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockDB) UpdateTransaction(id string, status models.Status) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockDB) GetAllTransactions(limit, offset int) ([]models.Transaction, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

func (m *MockDB) GetTransaction(id string) (*models.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestHandler_CreateTransaction(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		// Create a transaction for testing (without ID and CreatedAt since they're generated)
		transactionInput := models.Transaction{
			Amount:   100.50,
			Currency: "USD",
			Sender:   "user-1",
			Receiver: "user-2",
			// Status and ID will be set by handler
		}

		// Set up mock expectation - the handler will generate ID and set CreatedAt
		mockDB.On("CreateTransaction", mock.MatchedBy(func(tx models.Transaction) bool {
			return tx.Amount == transactionInput.Amount &&
				tx.Currency == transactionInput.Currency &&
				tx.Sender == transactionInput.Sender &&
				tx.Receiver == transactionInput.Receiver &&
				tx.Status == models.StatusPending &&
				tx.ID != "" &&
				!tx.CreatedAt.IsZero()
		})).Return(nil)

		// Create request body
		body, err := json.Marshal(transactionInput)
		require.NoError(t, err)

		// Create request
		req := httptest.NewRequest("POST", "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call the handler
		handler.CreateTransaction(rr, req)

		// Check response
		assert.Equal(t, http.StatusCreated, rr.Code)

		var response models.Transaction
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Verify the response has the expected fields
		assert.Equal(t, transactionInput.Amount, response.Amount)
		assert.Equal(t, transactionInput.Currency, response.Currency)
		assert.Equal(t, transactionInput.Sender, response.Sender)
		assert.Equal(t, transactionInput.Receiver, response.Receiver)
		assert.Equal(t, models.StatusPending, response.Status)
		assert.NotEmpty(t, response.ID)
		assert.False(t, response.CreatedAt.IsZero())

		// Verify mock expectations
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		// Create invalid JSON
		body := bytes.NewReader([]byte(`{"invalid": json`))

		req := httptest.NewRequest("POST", "/transactions", body)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateTransaction(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid request body")
	})

	t.Run("missing required fields", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		// Transaction with missing fields
		invalidTransaction := map[string]interface{}{
			"id":     "txn-123",
			"amount": 100.50,
			// Missing currency, sender, receiver
		}

		body, err := json.Marshal(invalidTransaction)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateTransaction(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "validation failed")
	})

	t.Run("negative amount", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		transaction := models.Transaction{
			ID:       "txn-123",
			Amount:   -100.50,
			Currency: "USD",
			Sender:   "user-1",
			Receiver: "user-2",
		}

		body, err := json.Marshal(transaction)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateTransaction(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "amount must be greater than 0")
	})

	t.Run("invalid currency", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		transaction := models.Transaction{
			ID:       "txn-123",
			Amount:   100.50,
			Currency: "INVALID",
			Sender:   "user-1",
			Receiver: "user-2",
		}

		body, err := json.Marshal(transaction)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateTransaction(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "currency must be a valid 3-letter ISO code")
	})

	t.Run("same sender and receiver", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		transaction := models.Transaction{
			ID:       "txn-123",
			Amount:   100.50,
			Currency: "USD",
			Sender:   "user-1",
			Receiver: "user-1",
		}

		body, err := json.Marshal(transaction)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateTransaction(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "sender and receiver must be different")
	})

	t.Run("amount too large", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		transaction := models.Transaction{
			ID:       "txn-123",
			Amount:   100000000.00,
			Currency: "USD",
			Sender:   "user-1",
			Receiver: "user-2",
		}

		body, err := json.Marshal(transaction)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateTransaction(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "amount must be less than 100,000,000")
	})

	t.Run("database error", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		transactionInput := models.Transaction{
			Amount:   100.50,
			Currency: "USD",
			Sender:   "user-1",
			Receiver: "user-2",
		}

		mockDB.On("CreateTransaction", mock.MatchedBy(func(tx models.Transaction) bool {
			return tx.Amount == transactionInput.Amount &&
				tx.Currency == transactionInput.Currency &&
				tx.Sender == transactionInput.Sender &&
				tx.Receiver == transactionInput.Receiver &&
				tx.Status == models.StatusPending &&
				tx.ID != "" &&
				!tx.CreatedAt.IsZero()
		})).Return(errors.New("database error"))

		body, err := json.Marshal(transactionInput)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateTransaction(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "error creating transaction")

		mockDB.AssertExpectations(t)
	})
}

func TestHandler_GetTransaction(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		transaction := &models.Transaction{
			ID:       "txn-123",
			Amount:   100.50,
			Currency: "USD",
			Sender:   "user-1",
			Receiver: "user-2",
			Status:   models.StatusCompleted,
		}

		mockDB.On("GetTransaction", "txn-123").Return(transaction, nil)

		req := httptest.NewRequest("GET", "/transactions/txn-123", nil)
		rr := httptest.NewRecorder()

		// Create router and set up the route
		router := mux.NewRouter()
		router.HandleFunc("/transactions/{id}", handler.GetTransaction).Methods("GET")

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response models.Transaction
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.EqualValues(t, *transaction, response)

		mockDB.AssertExpectations(t)
	})

	t.Run("transaction not found", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		// The actual implementation returns nil, nil for not found
		mockDB.On("GetTransaction", "non-existent").Return(nil, nil)

		req := httptest.NewRequest("GET", "/transactions/non-existent", nil)
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/transactions/{id}", handler.GetTransaction).Methods("GET")

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Should return null for non-existent transaction
		assert.Equal(t, "null", strings.TrimSpace(rr.Body.String()))

		mockDB.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		mockDB.On("GetTransaction", "txn-123").Return(nil, errors.New("database error"))

		req := httptest.NewRequest("GET", "/transactions/txn-123", nil)
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/transactions/{id}", handler.GetTransaction).Methods("GET")

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "error getting transaction")

		mockDB.AssertExpectations(t)
	})

	t.Run("missing transaction id", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		req := httptest.NewRequest("GET", "/transactions/", nil)
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/transactions/{id}", handler.GetTransaction).Methods("GET")

		router.ServeHTTP(rr, req)

		// The actual implementation returns 400 for missing ID, but mux returns 404 for route mismatch
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestHandler_ListTransactions(t *testing.T) {
	t.Run("successful list with pagination", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		transactions := []models.Transaction{
			{
				ID:        "txn-1",
				Amount:    100.50,
				Currency:  "USD",
				Sender:    "user-1",
				Receiver:  "user-2",
				Status:    models.StatusCompleted,
				CreatedAt: time.Now(),
			},
			{
				ID:        "txn-2",
				Amount:    200.75,
				Currency:  "EUR",
				Sender:    "user-3",
				Receiver:  "user-4",
				Status:    models.StatusPending,
				CreatedAt: time.Now(),
			},
		}

		mockDB.On("GetAllTransactions", 10, 0).Return(transactions, nil)

		req := httptest.NewRequest("GET", "/transactions?page=1&page_size=10", nil)
		rr := httptest.NewRecorder()

		handler.ListTransactions(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, float64(1), response["page"])
		assert.Equal(t, float64(2), response["page_size"]) // Should be actual count, not requested size

		// Verify transactions are in response
		transactionsData, ok := response["transactions"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, transactionsData, 2)

		mockDB.AssertExpectations(t)
	})

	t.Run("successful list with default pagination", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		transactions := []models.Transaction{}

		mockDB.On("GetAllTransactions", defaultPageSize, 0).Return(transactions, nil)

		req := httptest.NewRequest("GET", "/transactions", nil)
		rr := httptest.NewRecorder()

		handler.ListTransactions(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, float64(1), response["page"])
		assert.Equal(t, float64(0), response["page_size"]) // Should be actual count

		mockDB.AssertExpectations(t)
	})

	t.Run("invalid page parameters", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		transactions := []models.Transaction{}

		// Should use defaults for invalid page/page_size
		mockDB.On("GetAllTransactions", defaultPageSize, 0).Return(transactions, nil)

		req := httptest.NewRequest("GET", "/transactions?page=invalid&page_size=invalid", nil)
		rr := httptest.NewRecorder()

		handler.ListTransactions(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, float64(1), response["page"])
		assert.Equal(t, float64(0), response["page_size"]) // Should be actual count

		mockDB.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		mockDB.On("GetAllTransactions", 10, 0).Return([]models.Transaction{}, errors.New("database error"))

		req := httptest.NewRequest("GET", "/transactions?page=1&page_size=10", nil)
		rr := httptest.NewRecorder()

		handler.ListTransactions(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "error getting transactions")

		mockDB.AssertExpectations(t)
	})
}

func TestHandler_UpdateTransaction(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		updateReq := updateRequest{
			Status: models.StatusCompleted,
		}

		mockDB.On("UpdateTransaction", "txn-123", models.StatusCompleted).Return(nil)

		body, err := json.Marshal(updateReq)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/transactions/txn-123", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/transactions/{id}", handler.UpdateTransaction).Methods("PUT")

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Empty(t, rr.Body.String())

		mockDB.AssertExpectations(t)
	})

	t.Run("missing transaction id", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		updateReq := updateRequest{
			Status: models.StatusCompleted,
		}

		body, err := json.Marshal(updateReq)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/transactions/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/transactions/{id}", handler.UpdateTransaction).Methods("PUT")

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		body := bytes.NewReader([]byte(`{"status": "invalid"`))

		req := httptest.NewRequest("PUT", "/transactions/txn-123", body)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/transactions/{id}", handler.UpdateTransaction).Methods("PUT")

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid request body")
	})

	t.Run("invalid status", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		// Test with pending status (not allowed for updates)
		invalidReq := updateRequest{
			Status: models.StatusPending,
		}

		body, err := json.Marshal(invalidReq)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/transactions/txn-123", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/transactions/{id}", handler.UpdateTransaction).Methods("PUT")

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid request body")
	})

	t.Run("database error", func(t *testing.T) {
		mockDB := new(MockDB)
		handler := NewHandler(mockDB)

		updateReq := updateRequest{
			Status: models.StatusCompleted,
		}

		mockDB.On("UpdateTransaction", "txn-123", models.StatusCompleted).Return(errors.New("database error"))

		body, err := json.Marshal(updateReq)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/transactions/txn-123", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/transactions/{id}", handler.UpdateTransaction).Methods("PUT")

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "error updating transaction")

		mockDB.AssertExpectations(t)
	})
}

func TestHandler_RegisterRoutes(t *testing.T) {
	mockDB := new(MockDB)
	handler := NewHandler(mockDB)
	router := mux.NewRouter()

	handler.RegisterRoutes(router)

	// Collect all registered routes
	registeredRoutes := make(map[string]bool)

	err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			methods, err := route.GetMethods()
			if err == nil {
				for _, method := range methods {
					key := method + " " + pathTemplate
					registeredRoutes[key] = true
				}
			}
		}
		return nil
	})
	require.NoError(t, err)

	// Check expected routes
	expectedRoutes := []string{
		"POST /transactions",
		"GET /transactions",
		"GET /transactions/{id}",
		"PUT /transactions/{id}",
	}

	for _, expectedRoute := range expectedRoutes {
		assert.True(t, registeredRoutes[expectedRoute], "Route %s should be registered", expectedRoute)
	}
}
