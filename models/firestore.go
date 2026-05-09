package models

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yourusername/expense-tracker/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// FirestoreClient wraps Firestore client
type FirestoreClient struct {
	Client            *firestore.Client
	Ctx               context.Context
	memoriesCollection string
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
	return &FirestoreClient{Client: client, Ctx: ctx, memoriesCollection: "chat_memories"}, nil
}

// WithMemoriesCollection overrides the Firestore collection used for memories.
// Use in tests to avoid writing into the production chat_memories collection.
func (f *FirestoreClient) WithMemoriesCollection(name string) *FirestoreClient {
	f.memoriesCollection = name
	return f
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

func (f *FirestoreClient) SaveTransactions(txns []Transaction) error {
	if len(txns) == 0 {
		return nil
	}

	batch := f.Client.Batch()
	for _, txn := range txns {
		docRef := f.Client.Collection("transactions").NewDoc()
		batch.Set(docRef, txn)
	}

	if _, err := batch.Commit(f.Ctx); err != nil {
		return fmt.Errorf("failed to insert transactions: %v", err)
	}

	fmt.Printf("Saved %d transactions to Firestore\n", len(txns))
	return nil
}


func (f *FirestoreClient) FetchTransactionsByDateRange(from, to time.Time) ([]Transaction, error) {
	var txs []Transaction
	iter := f.Client.Collection("transactions").
		Where("datetime", ">=", from).
		Where("datetime", "<=", to).
		Documents(f.Ctx)
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
		tx.ID = doc.Ref.ID
		txs = append(txs, tx)
	}
	return txs, nil
}

// UpdateTransaction updates an existing transaction by ID, preserving datetime.
func (f *FirestoreClient) UpdateTransaction(id string, tx Transaction) error {
	_, err := f.Client.Collection("transactions").Doc(id).Update(f.Ctx, []firestore.Update{
		{Path: "type", Value: tx.Type},
		{Path: "vendor", Value: tx.Vendor},
		{Path: "amount", Value: tx.Amount},
		{Path: "category", Value: tx.Category},
	})
	if err != nil {
		return fmt.Errorf("failed to update transaction: %v", err)
	}
	return nil
}

func (f *FirestoreClient) DeleteTransaction(id string) error {
	_, err := f.Client.Collection("transactions").Doc(id).Delete(f.Ctx)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %v", err)
	}
	return nil
}

func (f *FirestoreClient) GetLatestTransactionTimeByType(txType string) (*time.Time, error) {
	iter := f.Client.Collection("transactions").Documents(f.Ctx)
	defer iter.Stop()

	var latest *time.Time
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch latest transaction for type %s: %v", txType, err)
		}

		var txn Transaction
		if err := doc.DataTo(&txn); err != nil {
			return nil, fmt.Errorf("failed to decode latest transaction for type %s: %v", txType, err)
		}

		if txn.Type != txType {
			continue
		}

		if latest == nil || txn.DateTime.After(*latest) {
			copyTime := txn.DateTime
			latest = &copyTime
		}
	}

	return latest, nil
}

// SaveUnparsedEmail stores unparsed email data for future analysis
func (f *FirestoreClient) SaveUnparsedEmail(body string, headers map[string]string) error {
	doc := map[string]interface{}{
		"body":      body,
		"body_text": utils.StripHTMLTags(body),
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

// MigrateFieldNames rewrites old Firestore documents that used Go struct field names
// (e.g. "DateTime") to the lowercase tag names (e.g. "datetime"). This is a one-time
// migration needed because firestore struct tags were added after initial data was written.
func (f *FirestoreClient) MigrateFieldNames() (int, int, error) {
	// mapping from old Pascal-case Firestore field name → new lowercase tag name
	rename := map[string]string{
		"Type":            "type",
		"CardEnding":      "cardending",
		"DebitedAccount":  "debitedaccount",
		"CreditedAccount": "creditedaccount",
		"Amount":          "amount",
		"Vendor":          "vendor",
		"DateTime":        "datetime",
		"Category":        "category",
	}

	iter := f.Client.Collection("transactions").Documents(f.Ctx)
	defer iter.Stop()

	migrated, skipped := 0, 0
	batch := f.Client.Batch()
	batchSize := 0

	flush := func() error {
		if batchSize == 0 {
			return nil
		}
		if _, err := batch.Commit(f.Ctx); err != nil {
			return fmt.Errorf("batch commit failed: %v", err)
		}
		batch = f.Client.Batch()
		batchSize = 0
		return nil
	}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return migrated, skipped, fmt.Errorf("failed to iterate transactions: %v", err)
		}

		raw := doc.Data()
		if _, hasOld := raw["DateTime"]; !hasOld {
			skipped++
			continue
		}

		newData := make(map[string]interface{}, len(raw))
		for k, v := range raw {
			if newKey, ok := rename[k]; ok {
				newData[newKey] = v
			} else {
				newData[k] = v
			}
		}

		batch.Set(doc.Ref, newData)
		batchSize++
		migrated++

		if batchSize >= 400 {
			if err := flush(); err != nil {
				return migrated, skipped, err
			}
		}
	}

	if err := flush(); err != nil {
		return migrated, skipped, err
	}

	return migrated, skipped, nil
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

func (f *FirestoreClient) SaveMemory(mem Memory) error {
	_, _, err := f.Client.Collection(f.memoriesCollection).Add(f.Ctx, mem)
	if err != nil {
		return fmt.Errorf("failed to save memory: %v", err)
	}
	return nil
}

func (f *FirestoreClient) GetAllMemories() ([]Memory, error) {
	iter := f.Client.Collection(f.memoriesCollection).OrderBy("created_at", firestore.Desc).Documents(f.Ctx)
	var memories []Memory
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch memories: %v", err)
		}
		var m Memory
		if err := doc.DataTo(&m); err == nil {
			memories = append(memories, m)
		}
	}
	return memories, nil
}

// DeleteAllMemories deletes every document in the memories collection.
// Only used in integration tests — not on the DatabaseClient interface.
func (f *FirestoreClient) DeleteAllMemories() error {
	iter := f.Client.Collection(f.memoriesCollection).Documents(f.Ctx)
	batch := f.Client.Batch()
	count := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to list memories for deletion: %v", err)
		}
		batch.Delete(doc.Ref)
		count++
	}
	if count == 0 {
		return nil
	}
	if _, err := batch.Commit(f.Ctx); err != nil {
		return fmt.Errorf("failed to delete memories: %v", err)
	}
	return nil
}
