// Package models defines the data structures used throughout the application.
// This package contains the core domain models for the transaction service.
package models

import "time"

// Status represents the current state of a transaction.
// Transactions can be in one of three states: pending, completed, or failed.
type Status string

const (
	// StatusPending indicates a transaction that has been created but not yet processed
	StatusPending Status = "pending"
	// StatusCompleted indicates a transaction that has been successfully processed
	StatusCompleted Status = "completed"
	// StatusFailed indicates a transaction that failed during processing
	StatusFailed Status = "failed"
)

// Transaction represents a financial transaction between two parties.
// It contains all the necessary information to process and track a money transfer.
type Transaction struct {
	// ID is a unique identifier for the transaction (max 64 characters)
	ID string `json:"id"`
	// Amount is the monetary value of the transaction (must be positive)
	Amount float64 `json:"amount"`
	// Currency is the 3-letter ISO currency code (e.g., USD, EUR, GBP)
	Currency string `json:"currency"`
	// Sender is the identifier of the party sending the money
	Sender string `json:"sender"`
	// Receiver is the identifier of the party receiving the money
	Receiver string `json:"receiver"`
	// Status indicates the current state of the transaction
	Status Status `json:"status"`
	// CreatedAt is the timestamp when the transaction was created
	CreatedAt time.Time `json:"created_at"`
}
