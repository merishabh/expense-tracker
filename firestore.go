package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// FirestoreClient wraps Firestore client
type FirestoreClient struct {
	Client *firestore.Client
	Ctx    context.Context
}

// NewFirestoreClient creates the Firestore client
func NewFirestoreClient() (*FirestoreClient, error) {
	ctx := context.Background()

	// Get project ID from environment variable
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_PROJECT environment variable is required")
	}

	// Create Firestore client
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %v", err)
	}

	fmt.Printf("Connected to Firestore project: %s\n", projectID)
	return &FirestoreClient{Client: client, Ctx: ctx}, nil
}

// SaveTransaction stores a Transaction document
func (f *FirestoreClient) SaveTransaction(txn Transaction) error {
	docRef, _, err := f.Client.Collection("transactions").Add(f.Ctx, txn)
	if err != nil {
		return err
	}
	fmt.Printf("Saved transaction to Firestore with ID: %s\n", docRef.ID)
	return nil
}

// FetchAllTransactions retrieves all transactions from Firestore
func (f *FirestoreClient) FetchAllTransactions() ([]Transaction, error) {
	var txs []Transaction
	iter := f.Client.Collection("transactions").Documents(f.Ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var tx Transaction
		if err := doc.DataTo(&tx); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	fmt.Printf("📊 Fetched %d transactions from Firestore\n", len(txs))
	return txs, nil
}

// SaveUnparsedEmail stores unparsed email data for future analysis
func (f *FirestoreClient) SaveUnparsedEmail(body string, headers map[string]string) error {
	doc := map[string]interface{}{
		"body":      body,
		"body_text": stripHTMLTags(body),
		"headers":   headers,
		"timestamp": time.Now(),
	}
	_, _, err := f.Client.Collection("unparsed_emails").Add(f.Ctx, doc)
	return err
}

// Close closes the Firestore connection
func (f *FirestoreClient) Close() error {
	return f.Client.Close()
}

// GetCategoryMapping retrieves a vendor-to-category mapping from Firestore
func (f *FirestoreClient) GetCategoryMapping(vendor string) (*CategoryMapping, error) {
	vendor = strings.ToLower(vendor)
	doc, err := f.Client.Collection("category_mappings").Doc(vendor).Get(f.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find category mapping: %v", err)
	}

	var mapping CategoryMapping
	if err := doc.DataTo(&mapping); err != nil {
		return nil, fmt.Errorf("failed to decode category mapping: %v", err)
	}

	return &mapping, nil
}

// SaveCategoryMapping stores a vendor-to-category mapping in Firestore
func (f *FirestoreClient) SaveCategoryMapping(mapping *CategoryMapping) error {
	vendor := strings.ToLower(mapping.Vendor)
	_, err := f.Client.Collection("category_mappings").Doc(vendor).Set(f.Ctx, mapping)
	if err != nil {
		return fmt.Errorf("failed to save category mapping: %v", err)
	}

	fmt.Printf("Saved category mapping: %s -> %s (%s)\n", mapping.Vendor, mapping.Category, mapping.Source)
	return nil
}
