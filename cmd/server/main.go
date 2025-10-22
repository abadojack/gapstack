// Package main provides the entry point for the gapstack transaction service.
// This service exposes a REST API for managing financial transactions backed by MySQL.
package main

import (
	"log"
	"net/http"

	"github.com/abadojack/gapstack/internal/api"
	db "github.com/abadojack/gapstack/internal/db"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize database connection
	database, err := db.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	// Create API handler with database dependency
	handler := api.NewHandler(database)

	// Set up HTTP router with Gorilla Mux
	r := mux.NewRouter()

	// Register all API routes
	handler.RegisterRoutes(r)

	// Start HTTP server on port 8080
	log.Print("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
