// Package db implements the database operations for the transaction service.
// This file contains the CRUD operations for transactions using MySQL.
package db

import (
	"database/sql"
	"errors"
	"log"

	"github.com/abadojack/gapstack/internal/models"
)

// CreateTransaction inserts a new transaction into the database.
// The created_at timestamp is automatically set by MySQL using the DEFAULT CURRENT_TIMESTAMP.
func (db *DBImpl) CreateTransaction(transaction models.Transaction) error {
	query := "INSERT INTO transactions(id, amount, currency, sender, receiver, status) VALUES (?, ?, ?, ?, ?, ?)"

	log.Println("TEST")

	_, err := db.DB.Exec(query, transaction.ID, transaction.Amount, transaction.Currency, transaction.Sender, transaction.Receiver, transaction.Status)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// UpdateTransaction updates the status of an existing transaction.
// Only completed and failed statuses are allowed for updates.
func (db *DBImpl) UpdateTransaction(id string, status models.Status) error {
	query := "UPDATE transactions SET status = ? WHERE id = ?"
	_, err := db.DB.Exec(query, status, id)
	if err != nil {
		return err
	}
	return nil
}

// GetAllTransactions retrieves a paginated list of all transactions from the database.
// The results are ordered by transaction ID and limited by the provided limit and offset.
func (db *DBImpl) GetAllTransactions(limit, offset int) ([]models.Transaction, error) {
	query := `
		SELECT id, amount, currency, sender, receiver, status, created_at
		FROM transactions
		ORDER BY id
		LIMIT ? OFFSET ?
	`

	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction

	// Iterate through all rows and scan them into Transaction structs
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.Amount,
			&transaction.Currency,
			&transaction.Sender,
			&transaction.Receiver,
			&transaction.Status,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	// Check for any errors that occurred during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetTransaction retrieves a single transaction by its ID.
// Returns nil if no transaction is found with the given ID.
func (db *DBImpl) GetTransaction(id string) (*models.Transaction, error) {
	query := "SELECT id, amount, currency, sender, receiver, status, created_at FROM transactions WHERE id = ?"
	row := db.DB.QueryRow(query, id)

	var transaction models.Transaction
	err := row.Scan(
		&transaction.ID,
		&transaction.Amount,
		&transaction.Currency,
		&transaction.Sender,
		&transaction.Receiver,
		&transaction.Status,
		&transaction.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No transaction found with that ID
		}
		return nil, err
	}

	return &transaction, nil
}
