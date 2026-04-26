package services

import (
	"testing"
	"time"

	"github.com/yourusername/expense-tracker/models"
)

type reportingTestDB struct {
	transactions []models.Transaction
}

func (d *reportingTestDB) SaveTransaction(txn models.Transaction) error { return nil }
func (d *reportingTestDB) UpdateTransaction(id string, txn models.Transaction) error {
	return nil
}
func (d *reportingTestDB) FetchAllTransactions() ([]models.Transaction, error) {
	return d.transactions, nil
}
func (d *reportingTestDB) GetLatestTransactionTimeByType(txType string) (*time.Time, error) {
	return nil, nil
}
func (d *reportingTestDB) SaveUnparsedEmail(body string, headers map[string]string) error {
	return nil
}
func (d *reportingTestDB) GetCategoryMapping(vendor string) (*models.CategoryMapping, error) {
	return nil, nil
}
func (d *reportingTestDB) SaveCategoryMapping(mapping *models.CategoryMapping) error { return nil }
func (d *reportingTestDB) SaveMemory(mem models.Memory) error                        { return nil }
func (d *reportingTestDB) GetAllMemories() ([]models.Memory, error)                  { return nil, nil }
func (d *reportingTestDB) Close() error                                              { return nil }

func TestGetTotalSummaryDeductsCredits(t *testing.T) {
	now := time.Now().UTC()
	db := &reportingTestDB{
		transactions: []models.Transaction{
			{
				Type:           GooglePayTransactionType,
				Amount:         1000,
				Vendor:         "Store",
				Category:       "Shopping",
				DateTime:       now,
				DebitedAccount: "XXXX1234",
			},
			{
				Type:            GooglePayTransactionType,
				Amount:          -800,
				Vendor:          "Friend",
				Category:        "Other",
				DateTime:        now.Add(-time.Minute),
				CreditedAccount: "Google Pay",
			},
		},
	}

	reporting := NewReportingService(db)
	summary, err := reporting.GetTotalSummary("THIS_MONTH")
	if err != nil {
		t.Fatalf("GetTotalSummary returned error: %v", err)
	}

	if summary.TotalAmount != 200 {
		t.Fatalf("expected total amount 200, got %v", summary.TotalAmount)
	}

	if summary.GrossExpense != 1000 {
		t.Fatalf("expected gross expense 1000, got %v", summary.GrossExpense)
	}

	if summary.CreditAmount != 800 {
		t.Fatalf("expected credit amount 800, got %v", summary.CreditAmount)
	}

	if summary.AverageAmount != 500 {
		t.Fatalf("expected average amount 500, got %v", summary.AverageAmount)
	}
}
