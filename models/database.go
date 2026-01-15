package models

import (
	"os"
	"strings"
)

// DatabaseClient interface defines the common operations for database access
type DatabaseClient interface {
	SaveTransaction(txn Transaction) error
	FetchAllTransactions() ([]Transaction, error)
	SaveUnparsedEmail(body string, headers map[string]string) error
	GetCategoryMapping(vendor string) (*CategoryMapping, error)
	SaveCategoryMapping(mapping *CategoryMapping) error
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
