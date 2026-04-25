package services

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/expense-tracker/models"
)

type googlePayTestDB struct {
	latest         *time.Time
	saved          []models.Transaction
	saveCalls      int
	batchSaveCalls int
}

func (d *googlePayTestDB) SaveTransaction(txn models.Transaction) error {
	d.saveCalls++
	d.saved = append(d.saved, txn)
	return nil
}

func (d *googlePayTestDB) SaveTransactions(txns []models.Transaction) error {
	d.batchSaveCalls++
	d.saved = append(d.saved, txns...)
	return nil
}

func (d *googlePayTestDB) FetchAllTransactions() ([]models.Transaction, error) {
	return nil, nil
}

func (d *googlePayTestDB) UpdateTransaction(id string, txn models.Transaction) error {
	return nil
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
<div class="content-cell mdl-cell mdl-cell--6-col mdl-typography--body-1">Paid ₹1,100.00 to RAMESHWARAM ENTERPRISES using Bank Account XXXXXXXX0000<br>Apr 19, 2026, 8:31:30 AM GMT+05:30<br></div>
<div class="content-cell mdl-cell mdl-cell--12-col mdl-typography--caption"><b>Details:</b><br>&emsp;abc123<br>&emsp;Completed<br></div>
</div></div>
<div class="outer-cell mdl-cell mdl-cell--12-col mdl-shadow--2dp"><div class="mdl-grid">
<div class="content-cell mdl-cell mdl-cell--6-col mdl-typography--body-1">Paid ₹399.00 to sonyliv using Bank Account XXXXXXXX0000<br>Apr 16, 2026, 5:12:42 PM GMT+05:30<br></div>
<div class="content-cell mdl-cell mdl-cell--12-col mdl-typography--caption"><b>Details:</b><br>&emsp;def456<br>&emsp;Cancelled<br></div>
</div></div>
<div class="outer-cell mdl-cell mdl-cell--12-col mdl-shadow--2dp"><div class="mdl-grid">
<div class="content-cell mdl-cell mdl-cell--6-col mdl-typography--body-1">Paid ₹900.90 to Airtel Prepaid using Bank Account XXXXXXXX0000<br>Apr 17, 2026, 7:44:01 AM GMT+05:30<br></div>
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

	if db.saved[0].Category != "Food" {
		t.Fatalf("expected category Food, got %q", db.saved[0].Category)
	}
}

func TestImportGooglePayHTMLCategorizesRecognizedMerchants(t *testing.T) {
	db := &googlePayTestDB{}

	html := `
<html><body>
<div class="outer-cell mdl-cell mdl-cell--12-col mdl-shadow--2dp"><div class="mdl-grid">
<div class="content-cell mdl-cell mdl-cell--6-col mdl-typography--body-1">Paid ₹900.90 to Airtel Prepaid using Bank Account XXXXXXXX0000<br>Apr 15, 2026, 6:05:07 PM GMT+05:30<br></div>
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

func TestImportGooglePayHTMLMarksReceivedTransactionsAsCredits(t *testing.T) {
	db := &googlePayTestDB{}

	html := `
<html><body>
<div class="outer-cell mdl-cell mdl-cell--12-col mdl-shadow--2dp"><div class="mdl-grid">
<div class="content-cell mdl-cell mdl-cell--6-col mdl-typography--body-1">Received ₹80,000.00 from Rishabh<br>Feb 16, 2026, 9:57:40 PM GMT+05:30<br></div>
<div class="content-cell mdl-cell mdl-cell--12-col mdl-typography--caption"><b>Details:</b><br>&emsp;abc123<br>&emsp;Completed<br></div>
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

	if db.saved[0].CreditedAccount != "Google Pay" {
		t.Fatalf("expected credited account to be Google Pay, got %q", db.saved[0].CreditedAccount)
	}

	if db.saved[0].DebitedAccount != "" {
		t.Fatalf("expected debited account to be empty, got %q", db.saved[0].DebitedAccount)
	}

	if db.saved[0].Amount != -80000 {
		t.Fatalf("expected amount -80000, got %v", db.saved[0].Amount)
	}

	if !db.saved[0].IsCredit() {
		t.Fatalf("expected received transaction to be treated as credit")
	}
}

func TestImportGooglePayHTMLBatchesWritesAndReportsProgress(t *testing.T) {
	db := &googlePayTestDB{}

	var builder strings.Builder
	builder.WriteString("<html><body>")
	for i := 0; i < googlePayBatchSize+1; i++ {
		builder.WriteString(`
<div class="outer-cell mdl-cell mdl-cell--12-col mdl-shadow--2dp"><div class="mdl-grid">
<div class="content-cell mdl-cell mdl-cell--6-col mdl-typography--body-1">Paid ₹100.00 to Merchant `)
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString(` using Bank Account XXXXXXXX0000<br>Apr 20, 2026, 8:31:30 AM GMT+05:30<br></div>
<div class="content-cell mdl-cell mdl-cell--12-col mdl-typography--caption"><b>Details:</b><br>&emsp;`)
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString(`<br>&emsp;Completed<br></div>
</div></div>`)
	}
	builder.WriteString("</body></html>")

	var progress []GooglePayImportSummary
	summary, err := ImportGooglePayHTMLWithProgress(strings.NewReader(builder.String()), db, func(summary GooglePayImportSummary) {
		progress = append(progress, summary)
	})
	if err != nil {
		t.Fatalf("ImportGooglePayHTMLWithProgress returned error: %v", err)
	}

	if summary.ImportedCount != googlePayBatchSize+1 {
		t.Fatalf("expected %d imported transactions, got %d", googlePayBatchSize+1, summary.ImportedCount)
	}

	if summary.BatchCount != 2 {
		t.Fatalf("expected 2 batch writes, got %d", summary.BatchCount)
	}

	if summary.PendingCount != 0 {
		t.Fatalf("expected no pending transactions after import, got %d", summary.PendingCount)
	}

	if db.batchSaveCalls != 2 {
		t.Fatalf("expected 2 SaveTransactions calls, got %d", db.batchSaveCalls)
	}

	if db.saveCalls != 0 {
		t.Fatalf("expected SaveTransaction fallback to be unused, got %d calls", db.saveCalls)
	}

	if len(progress) == 0 {
		t.Fatalf("expected progress updates to be emitted")
	}

	last := progress[len(progress)-1]
	if last.ImportedCount != summary.ImportedCount || last.ProcessedCount != summary.ProcessedCount {
		t.Fatalf("expected final progress update to match summary, got %+v want %+v", last, summary)
	}
}
