package main

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
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

	fmt.Printf("🔗 Connected to Firestore project: %s\n", projectID)
	return &FirestoreClient{Client: client, Ctx: ctx}, nil
}

// SaveTransaction stores a Transaction document
func (f *FirestoreClient) SaveTransaction(txn Transaction) error {
	docRef, _, err := f.Client.Collection("transactions").Add(f.Ctx, txn)
	if err != nil {
		return err
	}
	fmt.Printf("📝 Saved transaction to Firestore with ID: %s\n", docRef.ID)
	return nil
}
