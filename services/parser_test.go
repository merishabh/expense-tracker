package services

import (
	"testing"
	"time"

	"github.com/yourusername/expense-tracker/models"
)

type parserTestDB struct{}

func (d *parserTestDB) SaveTransaction(txn models.Transaction) error { return nil }
func (d *parserTestDB) UpdateTransaction(id string, txn models.Transaction) error {
	return nil
}
func (d *parserTestDB) FetchAllTransactions() ([]models.Transaction, error) {
	return nil, nil
}
func (d *parserTestDB) GetLatestTransactionTimeByType(txType string) (*time.Time, error) {
	return nil, nil
}
func (d *parserTestDB) SaveUnparsedEmail(body string, headers map[string]string) error { return nil }
func (d *parserTestDB) GetCategoryMapping(vendor string) (*models.CategoryMapping, error) {
	return nil, nil
}
func (d *parserTestDB) SaveCategoryMapping(mapping *models.CategoryMapping) error { return nil }
func (d *parserTestDB) SaveMemory(mem models.Memory) error                        { return nil }
func (d *parserTestDB) GetAllMemories() ([]models.Memory, error)                  { return nil, nil }
func (d *parserTestDB) Close() error                                              { return nil }

func TestParseICICICreditCardTransaction_AmazonPayAlert(t *testing.T) {
	db := &parserTestDB{}
	text := "Your ICICI Bank Credit Card XX3013 has been used for a transaction of INR 241.00 on Jan 23, 2026 at 04:52:40. Info: AMAZON PAY IN E COMMERCE."

	tx := ParseICICICreditCardTransaction(text, db)
	if tx == nil {
		t.Fatalf("expected transaction to be parsed")
	}

	if tx.Type != "ICICICreditCard" {
		t.Fatalf("unexpected type %q", tx.Type)
	}
	if tx.CardEnding != "XX3013" {
		t.Fatalf("unexpected card ending %q", tx.CardEnding)
	}
	if tx.Amount != 241.00 {
		t.Fatalf("unexpected amount %v", tx.Amount)
	}
	if tx.Vendor != "AMAZON PAY IN E COMMERCE" {
		t.Fatalf("unexpected vendor %q", tx.Vendor)
	}
	if tx.Category != "Amazon" {
		t.Fatalf("unexpected category %q", tx.Category)
	}
}

func TestParseICICICreditCardTransaction_WithTrailingSentence(t *testing.T) {
	db := &parserTestDB{}
	text := "Your ICICI Bank Credit Card XX3013 has been used for a transaction of INR 241.00 on Jan 23, 2026 at 04:52:40. Info: AMAZON PAY IN E COMMERCE. The Available Credit Limit on your card is INR 1,00,000.00."

	tx := ParseICICICreditCardTransaction(text, db)
	if tx == nil {
		t.Fatalf("expected transaction to be parsed")
	}

	if tx.Vendor != "AMAZON PAY IN E COMMERCE" {
		t.Fatalf("unexpected vendor %q", tx.Vendor)
	}
}
