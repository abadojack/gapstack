package db

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/abadojack/gapstack/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTransaction(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		transaction := models.Transaction{
			ID:       "txn-123",
			Amount:   100.50,
			Currency: "USD",
			Sender:   "user-1",
			Receiver: "user-2",
			Status:   models.StatusPending,
		}

		mock.ExpectExec("INSERT INTO transactions").
			WithArgs(transaction.ID, transaction.Amount, transaction.Currency,
				transaction.Sender, transaction.Receiver, transaction.Status).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = mockDB.CreateTransaction(transaction)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("creation fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		transaction := models.Transaction{
			ID:       "txn-123",
			Amount:   100.50,
			Currency: "USD",
			Sender:   "user-1",
			Receiver: "user-2",
			Status:   models.StatusPending,
		}

		expectedErr := errors.New("database error")
		mock.ExpectExec("INSERT INTO transactions").
			WithArgs(transaction.ID, transaction.Amount, transaction.Currency,
				transaction.Sender, transaction.Receiver, transaction.Status).
			WillReturnError(expectedErr)

		err = mockDB.CreateTransaction(transaction)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateTransaction(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		id := "txn-123"
		status := models.StatusCompleted

		mock.ExpectExec("UPDATE transactions SET status = \\? WHERE id = \\?").
			WithArgs(status, id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = mockDB.UpdateTransaction(id, status)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		id := "txn-123"
		status := models.StatusCompleted

		expectedErr := errors.New("update error")
		mock.ExpectExec("UPDATE transactions SET status = \\? WHERE id = \\?").
			WithArgs(status, id).
			WillReturnError(expectedErr)

		err = mockDB.UpdateTransaction(id, status)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update with no rows affected", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		id := "non-existent-id"
		status := models.StatusCompleted

		mock.ExpectExec("UPDATE transactions SET status = \\? WHERE id = \\?").
			WithArgs(status, id).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err = mockDB.UpdateTransaction(id, status)
		assert.NoError(t, err) // No error expected even if no rows updated
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetAllTransactions(t *testing.T) {
	t.Run("successful get all transactions", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		limit, offset := 10, 0

		expectedTransactions := []models.Transaction{
			{
				ID:       "txn-1",
				Amount:   100.50,
				Currency: "USD",
				Sender:   "user-1",
				Receiver: "user-2",
				Status:   models.StatusCompleted,
			},
			{
				ID:       "txn-2",
				Amount:   200.75,
				Currency: "EUR",
				Sender:   "user-3",
				Receiver: "user-4",
				Status:   models.StatusPending,
			},
		}

		rows := sqlmock.NewRows([]string{"id", "amount", "currency", "sender", "receiver", "status", "created_at"}).
			AddRow(expectedTransactions[0].ID, expectedTransactions[0].Amount, expectedTransactions[0].Currency,
				expectedTransactions[0].Sender, expectedTransactions[0].Receiver, expectedTransactions[0].Status, time.Time{}).
			AddRow(expectedTransactions[1].ID, expectedTransactions[1].Amount, expectedTransactions[1].Currency,
				expectedTransactions[1].Sender, expectedTransactions[1].Receiver, expectedTransactions[1].Status, time.Time{})

		mock.ExpectQuery("SELECT id, amount, currency, sender, receiver, status, created_at FROM transactions ORDER BY id LIMIT \\? OFFSET \\?").
			WithArgs(limit, offset).
			WillReturnRows(rows)

		transactions, err := mockDB.GetAllTransactions(limit, offset)
		assert.NoError(t, err)
		assert.Equal(t, expectedTransactions, transactions)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		limit, offset := 10, 100

		rows := sqlmock.NewRows([]string{"id", "amount", "currency", "sender", "receiver", "status", "created_at"})

		mock.ExpectQuery("SELECT id, amount, currency, sender, receiver, status, created_at FROM transactions ORDER BY id LIMIT \\? OFFSET \\?").
			WithArgs(limit, offset).
			WillReturnRows(rows)

		transactions, err := mockDB.GetAllTransactions(limit, offset)
		assert.NoError(t, err)
		assert.Empty(t, transactions)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		limit, offset := 10, 0

		expectedErr := errors.New("query error")
		mock.ExpectQuery("SELECT id, amount, currency, sender, receiver, status, created_at FROM transactions ORDER BY id LIMIT \\? OFFSET \\?").
			WithArgs(limit, offset).
			WillReturnError(expectedErr)

		transactions, err := mockDB.GetAllTransactions(limit, offset)
		assert.Error(t, err)
		assert.Nil(t, transactions)
		assert.Equal(t, expectedErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		limit, offset := 10, 0

		// Return rows with wrong data type for amount to cause scan error
		rows := sqlmock.NewRows([]string{"id", "amount", "currency", "sender", "receiver", "status", "created_at"}).
			AddRow("txn-1", "not-a-float", "USD", "user-1", "user-2", models.StatusPending, time.Time{})

		mock.ExpectQuery("SELECT id, amount, currency, sender, receiver, status, created_at FROM transactions ORDER BY id LIMIT \\? OFFSET \\?").
			WithArgs(limit, offset).
			WillReturnRows(rows)

		transactions, err := mockDB.GetAllTransactions(limit, offset)
		assert.Error(t, err)
		assert.Nil(t, transactions)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetTransaction(t *testing.T) {
	t.Run("successful get transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		id := "txn-123"

		expectedTransaction := &models.Transaction{
			ID:       id,
			Amount:   100.50,
			Currency: "USD",
			Sender:   "user-1",
			Receiver: "user-2",
			Status:   models.StatusCompleted,
		}

		row := sqlmock.NewRows([]string{"id", "amount", "currency", "sender", "receiver", "status", "created_at"}).
			AddRow(expectedTransaction.ID, expectedTransaction.Amount, expectedTransaction.Currency,
				expectedTransaction.Sender, expectedTransaction.Receiver, expectedTransaction.Status, time.Time{})

		mock.ExpectQuery("SELECT id, amount, currency, sender, receiver, status, created_at FROM transactions WHERE id = \\?").
			WithArgs(id).
			WillReturnRows(row)

		transaction, err := mockDB.GetTransaction(id)
		assert.NoError(t, err)
		assert.Equal(t, expectedTransaction, transaction)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("transaction not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		id := "non-existent-id"

		mock.ExpectQuery("SELECT id, amount, currency, sender, receiver, status, created_at FROM transactions WHERE id = \\?").
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		transaction, err := mockDB.GetTransaction(id)
		assert.NoError(t, err)
		assert.Nil(t, transaction)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query fails with non-ErrNoRows error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		id := "txn-123"

		expectedErr := errors.New("database error")
		mock.ExpectQuery("SELECT id, amount, currency, sender, receiver, status, created_at FROM transactions WHERE id = \\?").
			WithArgs(id).
			WillReturnError(expectedErr)

		transaction, err := mockDB.GetTransaction(id)
		assert.Error(t, err)
		assert.Nil(t, transaction)
		assert.Equal(t, expectedErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mockDB := &DBImpl{DB: db}
		id := "txn-123"

		// Return row with wrong data type for amount to cause scan error
		row := sqlmock.NewRows([]string{"id", "amount", "currency", "sender", "receiver", "status", "created_at"}).
			AddRow(id, "not-a-float", "USD", "user-1", "user-2", models.StatusPending, time.Time{})

		mock.ExpectQuery("SELECT id, amount, currency, sender, receiver, status, created_at FROM transactions WHERE id = \\?").
			WithArgs(id).
			WillReturnRows(row)

		transaction, err := mockDB.GetTransaction(id)
		assert.Error(t, err)
		assert.Nil(t, transaction)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
