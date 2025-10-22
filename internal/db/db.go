// Package db provides database connectivity and configuration management.
// It handles MySQL connections, environment variable loading, and connection pooling.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/abadojack/gapstack/internal/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// DB defines the interface for database operations.
// This interface allows for easy testing by providing mock implementations.
type DB interface {
	// CreateTransaction inserts a new transaction into the database
	CreateTransaction(transaction models.Transaction) error
	// UpdateTransaction updates the status of an existing transaction
	UpdateTransaction(id string, status models.Status) error
	// GetAllTransactions retrieves a paginated list of all transactions
	GetAllTransactions(limit, offset int) ([]models.Transaction, error)
	// GetTransaction retrieves a single transaction by its ID
	GetTransaction(id string) (*models.Transaction, error)
	// Close closes the database connection
	Close() error
}

// DBImpl is the concrete implementation of the DB interface.
// It wraps a sql.DB instance and provides transaction-specific operations.
type DBImpl struct {
	DB *sql.DB
}

// Ensure DBImpl implements the DB interface at compile time
var _ DB = (*DBImpl)(nil)

// NewDB creates a new database connection and returns the DB interface.
// It loads configuration from environment variables and establishes a connection to MySQL.
func NewDB() (DB, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	sqlDB, err := connectDB(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &DBImpl{DB: sqlDB}, nil
}

// NewDBWithInstance creates a DB instance with an existing sql.DB.
// This is useful for testing with mock databases or existing connections.
func NewDBWithInstance(sqlDB *sql.DB) DB {
	return &DBImpl{DB: sqlDB}
}

// Config holds database connection configuration parameters.
type Config struct {
	// DBUser is the MySQL username
	DBUser string
	// DBPassword is the MySQL password
	DBPassword string
	// DBHost is the MySQL host address
	DBHost string
	// DBPort is the MySQL port number
	DBPort string
	// DBName is the MySQL database name
	DBName string
	// MaxOpenConns is the maximum number of open connections to the database
	MaxOpenConns int
	// MaxIdleConns is the maximum number of idle connections in the pool
	MaxIdleConns int
	// ConnMaxLifetime is the maximum amount of time a connection may be reused
	ConnMaxLifetime time.Duration
}

// loadConfig loads database configuration from environment variables.
// It first tries to load from a .env file, then falls back to system environment variables.
func loadConfig() (*Config, error) {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: No .env file found or error loading .env: %v", err)
		log.Println("Using system environment variables only")
	}

	// Required environment variables
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		return nil, fmt.Errorf("DB_USER environment variable is required")
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}

	// Optional environment variables with defaults
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "transactions_db")

	return &Config{
		DBUser:          dbUser,
		DBPassword:      dbPassword,
		DBHost:          dbHost,
		DBPort:          dbPort,
		DBName:          dbName,
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
		ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME_MINUTES", 5)) * time.Minute,
	}, nil
}

// connectDB establishes a connection to the MySQL database using the provided configuration.
// It sets up connection pooling and verifies the connection is working.
func connectDB(config *Config) (*sql.DB, error) {
	// Build connection string with MySQL-specific parameters
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&timeout=5s",
		config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DBImpl) Close() error {
	if db.DB != nil {
		return db.DB.Close()
	}
	return nil
}

// Helper functions for environment variable handling

// getEnv retrieves an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt retrieves an environment variable as an integer with a default value.
// If the environment variable cannot be parsed as an integer, it returns the default value.
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid value for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}
