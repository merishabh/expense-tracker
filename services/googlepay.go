package services

import (
	"fmt"
	"html"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/expense-tracker/models"
)

const GooglePayTransactionType = "GooglePay"

type GooglePayImportSummary struct {
	LatestStoredAt       *time.Time `json:"latest_stored_at,omitempty"`
	TotalBlocks          int        `json:"total_blocks"`
	ProcessedCount       int        `json:"processed_count"`
	ImportedCount        int        `json:"imported_count"`
	SkippedOldCount      int        `json:"skipped_old_count"`
	SkippedStatusCount   int        `json:"skipped_status_count"`
	SkippedInvalidCount  int        `json:"skipped_invalid_count"`
	BatchCount           int        `json:"batch_count"`
	PendingCount         int        `json:"pending_count"`
	StoppedAtExistingRow bool       `json:"stopped_at_existing_row"`
	LatestImportedAt     *time.Time `json:"latest_imported_at,omitempty"`
}

var (
	googlePayOuterCellSplitter = regexp.MustCompile(`(?i)<div class="outer-cell[^"]*"`)
	googlePayContentCellRe     = regexp.MustCompile(`(?is)<div class="content-cell([^"]*)">(.*?)</div>`)
	googlePayCaptionCellRe     = regexp.MustCompile(`(?is)<div class="content-cell[^"]*mdl-typography--caption[^"]*">(.*?)</div>`)
	googlePayBreakRe           = regexp.MustCompile(`(?i)<br\s*/?>`)
	googlePayTagRe             = regexp.MustCompile(`(?s)<[^>]*>`)
	googlePaySpaceRe           = regexp.MustCompile(`\s+`)
	googlePayAmountRe          = regexp.MustCompile(`₹\s*([\d,]+(?:\.\d+)?)`)
	googlePayAccountRe         = regexp.MustCompile(`using Bank Account ([A-Z0-9X]+)`)
)

const googlePayBatchSize = 250

type googlePayBatchSaver interface {
	SaveTransactions(txns []models.Transaction) error
}

type GooglePayImportProgress func(summary GooglePayImportSummary)

func ImportGooglePayHTML(r io.Reader, dbClient models.DatabaseClient) (GooglePayImportSummary, error) {
	return ImportGooglePayHTMLWithProgress(r, dbClient, nil)
}

func ImportGooglePayHTMLWithProgress(r io.Reader, dbClient models.DatabaseClient, progress GooglePayImportProgress) (GooglePayImportSummary, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return GooglePayImportSummary{}, fmt.Errorf("failed to read uploaded file: %w", err)
	}

	latestStoredAt, err := dbClient.GetLatestTransactionTimeByType(GooglePayTransactionType)
	if err != nil {
		return GooglePayImportSummary{}, err
	}

	summary := GooglePayImportSummary{
		LatestStoredAt: latestStoredAt,
	}

	blocks := googlePayOuterCellSplitter.Split(string(raw), -1)
	summary.TotalBlocks = max(len(blocks)-1, 0)

	pending := make([]models.Transaction, 0, googlePayBatchSize)
	flushPending := func() error {
		if len(pending) == 0 {
			return nil
		}

		if err := saveGooglePayBatch(dbClient, pending); err != nil {
			return err
		}

		summary.ImportedCount += len(pending)
		summary.BatchCount++
		summary.PendingCount = 0
		pending = pending[:0]
		notifyGooglePayProgress(progress, summary)
		return nil
	}

	for _, block := range blocks[1:] {
		summary.ProcessedCount++

		tx, status, err := parseGooglePayTransactionBlock(block, dbClient)
		if err != nil {
			summary.SkippedInvalidCount++
			notifyGooglePayProgress(progress, summary)
			continue
		}

		if !strings.EqualFold(status, "Completed") {
			summary.SkippedStatusCount++
			notifyGooglePayProgress(progress, summary)
			continue
		}

		if latestStoredAt != nil && !tx.DateTime.After(*latestStoredAt) {
			summary.SkippedOldCount++
			summary.StoppedAtExistingRow = true
			notifyGooglePayProgress(progress, summary)
			break
		}

		pending = append(pending, *tx)
		summary.PendingCount = len(pending)

		if summary.LatestImportedAt == nil || tx.DateTime.After(*summary.LatestImportedAt) {
			importedAt := tx.DateTime
			summary.LatestImportedAt = &importedAt
		}

		if len(pending) >= googlePayBatchSize {
			if err := flushPending(); err != nil {
				return summary, fmt.Errorf("failed to save Google Pay transaction batch ending at %s: %w", tx.DateTime.Format(time.RFC3339), err)
			}
			continue
		}

		notifyGooglePayProgress(progress, summary)
	}

	if err := flushPending(); err != nil {
		return summary, fmt.Errorf("failed to save final Google Pay transaction batch: %w", err)
	}

	notifyGooglePayProgress(progress, summary)
	return summary, nil
}

func notifyGooglePayProgress(progress GooglePayImportProgress, summary GooglePayImportSummary) {
	if progress != nil {
		progress(summary)
	}
}

func saveGooglePayBatch(dbClient models.DatabaseClient, txns []models.Transaction) error {
	if len(txns) == 0 {
		return nil
	}

	if saver, ok := dbClient.(googlePayBatchSaver); ok {
		return saver.SaveTransactions(txns)
	}

	for _, txn := range txns {
		if err := dbClient.SaveTransaction(txn); err != nil {
			return err
		}
	}

	return nil
}

func parseGooglePayTransactionBlock(block string, dbClient models.DatabaseClient) (*models.Transaction, string, error) {
	bodyHTML := ""
	bodyMatches := googlePayContentCellRe.FindAllStringSubmatch(block, -1)
	for _, match := range bodyMatches {
		if len(match) < 3 {
			continue
		}
		className := match[1]
		if strings.Contains(className, "mdl-typography--body-1") && !strings.Contains(className, "mdl-typography--text-right") {
			bodyHTML = match[2]
			break
		}
	}
	if bodyHTML == "" {
		return nil, "", fmt.Errorf("transaction body not found")
	}

	captionMatch := googlePayCaptionCellRe.FindStringSubmatch(block)
	if len(captionMatch) < 2 {
		return nil, "", fmt.Errorf("transaction details not found")
	}

	bodyLines := extractGooglePayLines(bodyHTML)
	if len(bodyLines) < 2 {
		return nil, "", fmt.Errorf("unexpected transaction body format")
	}

	tx, err := parseGooglePayBodyLines(bodyLines, dbClient)
	if err != nil {
		return nil, "", err
	}

	detailLines := extractGooglePayLines(captionMatch[1])
	status := ""
	if len(detailLines) > 0 {
		status = detailLines[len(detailLines)-1]
	}

	return tx, status, nil
}

func parseGooglePayBodyLines(lines []string, dbClient models.DatabaseClient) (*models.Transaction, error) {
	description := lines[0]
	timestampLine := lines[1]

	amountMatch := googlePayAmountRe.FindStringSubmatch(description)
	if len(amountMatch) < 2 {
		return nil, fmt.Errorf("amount not found")
	}

	amount, err := strconv.ParseFloat(strings.ReplaceAll(amountMatch[1], ",", ""), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	dateTime, err := parseGooglePayTime(timestampLine)
	if err != nil {
		return nil, err
	}

	account := ""
	if accountMatch := googlePayAccountRe.FindStringSubmatch(description); len(accountMatch) >= 2 {
		account = accountMatch[1]
	}

	tx := &models.Transaction{
		Type:           GooglePayTransactionType,
		Amount:         amount,
		DateTime:       dateTime,
		DebitedAccount: account,
	}

	switch {
	case strings.HasPrefix(description, "Paid "):
		tx.Vendor = betweenGooglePayTokens(description, " to ", " using Bank Account ")
		if tx.Vendor == "" {
			tx.Vendor = "Google Pay"
		}
	case strings.HasPrefix(description, "Sent "):
		tx.Vendor = betweenGooglePayTokens(description, " to ", " using Bank Account ")
		if tx.Vendor == "" {
			tx.Vendor = "Google Pay"
		}
	case strings.HasPrefix(description, "Received "):
		tx.Amount = -tx.Amount
		tx.DebitedAccount = ""
		tx.CreditedAccount = "Google Pay"
		tx.Vendor = afterGooglePayToken(description, " from ")
		if tx.Vendor == "" {
			tx.Vendor = "Google Pay"
		}
	default:
		return nil, fmt.Errorf("unsupported Google Pay transaction description: %s", description)
	}

	tx.Category = CategorizeTransaction(tx.Vendor, dbClient)
	return tx, nil
}

func parseGooglePayTime(value string) (time.Time, error) {
	normalized := strings.ReplaceAll(value, "\u202f", " ")
	normalized = strings.ReplaceAll(normalized, "\u00a0", " ")
	normalized = googlePaySpaceRe.ReplaceAllString(strings.TrimSpace(normalized), " ")

	layouts := []string{
		"Jan 2, 2006, 3:04:05 PM GMT-07:00",
		"Jan 2, 2006, 3:04:05 PM MST",
	}

	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, normalized)
		if err == nil {
			return parsed.UTC(), nil
		}
		lastErr = err
	}

	return time.Time{}, fmt.Errorf("invalid Google Pay timestamp %q: %w", value, lastErr)
}

func extractGooglePayLines(fragment string) []string {
	text := googlePayBreakRe.ReplaceAllString(fragment, "\n")
	text = html.UnescapeString(text)
	text = strings.ReplaceAll(text, "\u202f", " ")
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = googlePayTagRe.ReplaceAllString(text, "")

	rawLines := strings.Split(text, "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = googlePaySpaceRe.ReplaceAllString(strings.TrimSpace(line), " ")
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}

func betweenGooglePayTokens(value, startToken, endToken string) string {
	start := strings.Index(value, startToken)
	if start == -1 {
		return ""
	}
	start += len(startToken)

	end := strings.Index(value[start:], endToken)
	if end == -1 {
		return strings.TrimSpace(value[start:])
	}

	return strings.TrimSpace(value[start : start+end])
}

func afterGooglePayToken(value, token string) string {
	start := strings.Index(value, token)
	if start == -1 {
		return ""
	}
	return strings.TrimSpace(value[start+len(token):])
}
