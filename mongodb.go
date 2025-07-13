package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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
	fmt.Println("MongoDB URI:", mongoURI)
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017" // Default for local development
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	// Get database name from environment or use default
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "expense_tracker"
	}

	database := client.Database(dbName)

	fmt.Printf("🍃 Connected to MongoDB at %s, database: %s\n", mongoURI, dbName)
	return &MongoClient{
		Client:   client,
		Database: database,
		Ctx:      ctx,
	}, nil
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
		"body_text": stripHTMLTags(body),
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
