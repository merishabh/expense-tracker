package main

import (
	"os"
	"strings"
)

// DatabaseClient interface abstracts database operations
type DatabaseClient interface {
	SaveTransaction(txn Transaction) error
	FetchAllTransactions() ([]Transaction, error)
	SaveUnparsedEmail(body string, headers map[string]string) error
	Close() error
}

// NewDatabaseClient creates a database client: Firestore for prod, MongoDB otherwise
func NewDatabaseClient() (DatabaseClient, error) {
	envVar, exists := os.LookupEnv("ENVIRONMENT")
	env := strings.ToLower(envVar)
	if exists && (env == "production" || env == "prod") {
		return NewFirestoreClient()
	}
	// If ENVIRONMENT is not set or not production, use MongoDB
	return NewMongoClient()
}
