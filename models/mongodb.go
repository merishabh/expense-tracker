package models

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yourusername/expense-tracker/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient wraps MongoDB client
type MongoClient struct {
	Client   *mongo.Client
	Database *mongo.Database
	Ctx      context.Context
}

// NewMongoClient creates a new MongoDB client
func NewMongoClient() (*MongoClient, error) {
	ctx := context.Background()

	// Get MongoDB connection string from environment
	mongoURI := os.Getenv("MONGODB_URI")

	var client *mongo.Client
	var err error

	if mongoURI != "" {
		// Use provided MONGODB_URI (may include auth or not)
		fmt.Printf("🔌 Connecting to MongoDB using MONGODB_URI: %s\n", maskPassword(mongoURI))
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
		}
	} else {
		// No MONGODB_URI set - try with Docker Compose credentials first
		fmt.Printf("🔌 Connecting to MongoDB (trying with auth)...\n")
		authURI := "mongodb://admin:password@localhost:27017/expense_tracker?authSource=admin"
		fmt.Printf("🔌 Using connection string: mongodb://admin:***@localhost:27017/expense_tracker?authSource=admin\n")
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(authURI))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
		}
		mongoURI = authURI
	}

	// Get database name from environment or use default
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "expense_tracker"
	}

	// Create database object with authentication options
	database := client.Database(dbName)

	// Try to authenticate by running a simple command
	// This ensures authentication is working before we return
	authCmd := bson.D{{Key: "ping", Value: 1}}
	var result bson.M
	if err := database.RunCommand(ctx, authCmd).Decode(&result); err != nil {
		// If authentication fails, try without auth as fallback
		if strings.Contains(err.Error(), "authentication") || strings.Contains(err.Error(), "auth") || strings.Contains(err.Error(), "Unauthorized") {
			if mongoURI == "" || strings.Contains(mongoURI, "@") {
				// We tried with auth, now try without
				fmt.Printf("⚠️  Authentication failed. Trying connection without authentication...\n")
				client.Disconnect(ctx)
				simpleURI := "mongodb://localhost:27017"
				client, err = mongo.Connect(ctx, options.Client().ApplyURI(simpleURI))
				if err != nil {
					return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
				}
				database = client.Database(dbName)
				// Test the connection
				if err := database.RunCommand(ctx, authCmd).Decode(&result); err != nil {
					return nil, fmt.Errorf("failed to authenticate MongoDB (tried with and without auth): %v", err)
				}
				fmt.Printf("✓ Connected to MongoDB without authentication\n")
			} else {
				return nil, fmt.Errorf("failed to authenticate MongoDB: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to authenticate MongoDB: %v", err)
		}
	} else {
		fmt.Printf("✓ Connected and authenticated to MongoDB successfully\n")
	}

	fmt.Printf("🍃 Using database: %s\n", dbName)
	return &MongoClient{
		Client:   client,
		Database: database,
		Ctx:      ctx,
	}, nil
}

// maskPassword masks password in connection string for logging
func maskPassword(uri string) string {
	if strings.Contains(uri, "@") {
		parts := strings.Split(uri, "@")
		if len(parts) == 2 {
			authPart := parts[0]
			if strings.Contains(authPart, ":") {
				authParts := strings.Split(authPart, ":")
				if len(authParts) >= 3 {
					// mongodb://user:pass@host
					return fmt.Sprintf("%s:***@%s", strings.Join(authParts[:len(authParts)-1], ":"), parts[1])
				}
			}
		}
	}
	return uri
}

// SaveTransaction stores a Transaction document in MongoDB
func (m *MongoClient) SaveTransaction(txn Transaction) error {
	collection := m.Database.Collection("transactions")

	result, err := collection.InsertOne(m.Ctx, txn)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %v", err)
	}

	fmt.Printf("Saved transaction to MongoDB with ID: %v\n", result.InsertedID)
	return nil
}

// FetchAllTransactions retrieves all transactions from MongoDB
func (m *MongoClient) FetchAllTransactions() ([]Transaction, error) {
	collection := m.Database.Collection("transactions")

	cursor, err := collection.Find(m.Ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}
	defer cursor.Close(m.Ctx)

	var transactions []Transaction
	if err := cursor.All(m.Ctx, &transactions); err != nil {
		return nil, fmt.Errorf("failed to decode transactions: %v", err)
	}

	fmt.Printf("📊 Fetched %d transactions from MongoDB\n", len(transactions))
	return transactions, nil
}

// SaveUnparsedEmail stores unparsed email data in MongoDB
func (m *MongoClient) SaveUnparsedEmail(body string, headers map[string]string) error {
	collection := m.Database.Collection("unparsed_emails")

	doc := bson.M{
		"body":      body,
		"body_text": utils.StripHTMLTags(body),
		"headers":   headers,
		"timestamp": time.Now(),
	}

	_, err := collection.InsertOne(m.Ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to insert unparsed email: %v", err)
	}

	return nil
}

// Close closes the MongoDB connection
func (m *MongoClient) Close() error {
	return m.Client.Disconnect(m.Ctx)
}

// GetCategoryMapping retrieves a vendor-to-category mapping from MongoDB
func (m *MongoClient) GetCategoryMapping(vendor string) (*CategoryMapping, error) {
	collection := m.Database.Collection("category_mappings")

	var mapping CategoryMapping
	err := collection.FindOne(m.Ctx, bson.M{"vendor": strings.ToLower(vendor)}).Decode(&mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to find category mapping: %v", err)
	}

	return &mapping, nil
}

// SaveCategoryMapping stores a vendor-to-category mapping in MongoDB
func (m *MongoClient) SaveCategoryMapping(mapping *CategoryMapping) error {
	collection := m.Database.Collection("category_mappings")

	// Use upsert to prevent duplicates
	filter := bson.M{"vendor": mapping.Vendor}
	update := bson.M{"$set": mapping}
	opts := options.UpdateOptions{}
	opts.SetUpsert(true)

	_, err := collection.UpdateOne(m.Ctx, filter, update, &opts)
	if err != nil {
		return fmt.Errorf("failed to save category mapping: %v", err)
	}

	fmt.Printf("💾 Saved category mapping: %s -> %s (%s)\n", mapping.Vendor, mapping.Category, mapping.Source)
	return nil
}
