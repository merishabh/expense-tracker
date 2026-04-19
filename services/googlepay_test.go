package services

import (
	"strings"
	"testing"
	"time"

	"github.com/yourusername/expense-tracker/models"
)

type googlePayTestDB struct {
	latest *time.Time
	saved  []models.Transaction
}

func (d *googlePayTestDB) SaveTransaction(txn models.Transaction) error {
	d.saved = append(d.saved, txn)
	return nil
}

func (d *googlePayTestDB) FetchAllTransactions() ([]models.Transaction, error) {
	return nil, nil
}

func (d *googlePayTestDB) GetLatestTransactionTimeByType(txType string) (*time.Time, error) {
	return d.latest, nil
}

func (d *googlePayTestDB) SaveUnparsedEmail(body string, headers map[string]string) error {
	return nil
}

func (d *googlePayTestDB) GetCategoryMapping(vendor string) (*models.CategoryMapping, error) {
	return nil, nil
}

func (d *googlePayTestDB) SaveCategoryMapping(mapping *models.CategoryMapping) error {
	return nil
}

func (d *googlePayTestDB) Close() error {
	return nil
}

func TestImportGooglePayHTMLStopsAtLatestStoredTransaction(t *testing.T) {
	latestStored := time.Date(2026, 4, 17, 2, 14, 1, 0, time.UTC)
	db := &googlePayTestDB{latest: &latestStored}

	html := `
<html><body>
<div class="outer-cell mdl-cell mdl-cell--12-col mdl-shadow--2dp"><div class="mdl-grid">
<div class="content-cell mdl-cell mdl-cell--6-col mdl-typography--body-1">Paid ₹1,100.00 to RAMESHWARAM ENTERPRISES using Bank Account XXXXXXXX5483<br>Apr 19, 2026, 8:31:30 AM GMT+05:30<br></div>
<div class="content-cell mdl-cell mdl-cell--12-col mdl-typography--caption"><b>Details:</b><br>&emsp;abc123<br>&emsp;Completed<br></div>
</div></div>
<div class="outer-cell mdl-cell mdl-cell--12-col mdl-shadow--2dp"><div class="mdl-grid">
<div class="content-cell mdl-cell mdl-cell--6-col mdl-typography--body-1">Paid ₹399.00 to sonyliv using Bank Account XXXXXXXX5483<br>Apr 16, 2026, 5:12:42 PM GMT+05:30<br></div>
<div class="content-cell mdl-cell mdl-cell--12-col mdl-typography--caption"><b>Details:</b><br>&emsp;def456<br>&emsp;Cancelled<br></div>
</div></div>
<div class="outer-cell mdl-cell mdl-cell--12-col mdl-shadow--2dp"><div class="mdl-grid">
<div class="content-cell mdl-cell mdl-cell--6-col mdl-typography--body-1">Paid ₹900.90 to Airtel Prepaid using Bank Account XXXXXXXX5483<br>Apr 17, 2026, 7:44:01 AM GMT+05:30<br></div>
<div class="content-cell mdl-cell mdl-cell--12-col mdl-typography--caption"><b>Details:</b><br>&emsp;ghi789<br>&emsp;Completed<br></div>
</div></div>
</body></html>`

	summary, err := ImportGooglePayHTML(strings.NewReader(html), db)
	if err != nil {
		t.Fatalf("ImportGooglePayHTML returned error: %v", err)
	}

	if summary.ImportedCount != 1 {
		t.Fatalf("expected 1 imported transaction, got %d", summary.ImportedCount)
	}

	if summary.SkippedStatusCount != 1 {
		t.Fatalf("expected 1 skipped status transaction, got %d", summary.SkippedStatusCount)
	}

	if !summary.StoppedAtExistingRow {
		t.Fatalf("expected importer to stop at existing row")
	}

	if len(db.saved) != 1 {
		t.Fatalf("expected 1 saved transaction, got %d", len(db.saved))
	}

	if db.saved[0].Type != GooglePayTransactionType {
		t.Fatalf("expected type %q, got %q", GooglePayTransactionType, db.saved[0].Type)
	}

	if db.saved[0].Vendor != "RAMESHWARAM ENTERPRISES" {
		t.Fatalf("unexpected vendor %q", db.saved[0].Vendor)
	}

	if db.saved[0].Category != "Other" {
		t.Fatalf("expected category Other, got %q", db.saved[0].Category)
	}
}

func TestImportGooglePayHTMLCategorizesRecognizedMerchants(t *testing.T) {
	db := &googlePayTestDB{}

	html := `
<html><body>
<div class="outer-cell mdl-cell mdl-cell--12-col mdl-shadow--2dp"><div class="mdl-grid">
<div class="content-cell mdl-cell mdl-cell--6-col mdl-typography--body-1">Paid ₹900.90 to Airtel Prepaid using Bank Account XXXXXXXX5483<br>Apr 15, 2026, 6:05:07 PM GMT+05:30<br></div>
<div class="content-cell mdl-cell mdl-cell--12-col mdl-typography--caption"><b>Details:</b><br>&emsp;xyz123<br>&emsp;Completed<br></div>
</div></div>
</body></html>`

	summary, err := ImportGooglePayHTML(strings.NewReader(html), db)
	if err != nil {
		t.Fatalf("ImportGooglePayHTML returned error: %v", err)
	}

	if summary.ImportedCount != 1 {
		t.Fatalf("expected 1 imported transaction, got %d", summary.ImportedCount)
	}

	if len(db.saved) != 1 {
		t.Fatalf("expected 1 saved transaction, got %d", len(db.saved))
	}

	if db.saved[0].Category != "Bills" {
		t.Fatalf("expected category Bills, got %q", db.saved[0].Category)
	}
}
