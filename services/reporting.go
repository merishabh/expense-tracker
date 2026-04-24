package services

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/yourusername/expense-tracker/models"
	"github.com/yourusername/expense-tracker/utils"
)

type ReportingService struct {
	dbClient models.DatabaseClient
}

type TotalSummary struct {
	Period             string  `json:"period"`
	TransactionCount   int     `json:"transaction_count"`
	TotalAmount        float64 `json:"total_amount"`
	GrossExpense       float64 `json:"gross_expense"`
	CreditAmount       float64 `json:"credit_amount"`
	AverageAmount      float64 `json:"average_amount"`
	UncategorizedCount int     `json:"uncategorized_count"`
}

type BreakdownItem struct {
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
}

type TrendPoint struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
}

type MonthlyComparison struct {
	CurrentMonthAmount   float64 `json:"current_month_amount"`
	LastMonthAmount      float64 `json:"last_month_amount"`
	CurrentMonthCount    int     `json:"current_month_count"`
	LastMonthCount       int     `json:"last_month_count"`
	DeltaAmount          float64 `json:"delta_amount"`
	DeltaPercent         float64 `json:"delta_percent"`
	TopMerchantThisMonth string  `json:"top_merchant_this_month"`
	TopMerchantSpend     float64 `json:"top_merchant_spend"`
}

func NewReportingService(dbClient models.DatabaseClient) *ReportingService {
	return &ReportingService{dbClient: dbClient}
}

func (s *ReportingService) ListTransactions(period string, category string, limit int) ([]models.Transaction, error) {
	txs, err := s.filteredTransactions(period)
	if err != nil {
		return nil, err
	}

	category = strings.TrimSpace(category)
	if category != "" {
		filteredByCategory := make([]models.Transaction, 0, len(txs))
		for _, tx := range txs {
			txCategory := strings.TrimSpace(tx.Category)
			if txCategory == "" {
				txCategory = "Other"
			}
			if strings.EqualFold(txCategory, category) {
				filteredByCategory = append(filteredByCategory, tx)
			}
		}
		txs = filteredByCategory
	}

	sort.Slice(txs, func(i, j int) bool {
		return txs[i].DateTime.After(txs[j].DateTime)
	})

	if limit > 0 && len(txs) > limit {
		txs = txs[:limit]
	}

	return txs, nil
}

func (s *ReportingService) GetTotalSummary(period string) (TotalSummary, error) {
	txs, err := s.filteredTransactions(period)
	if err != nil {
		return TotalSummary{}, err
	}

	summary := TotalSummary{Period: normalizePeriod(period)}
	for _, tx := range txs {
		summary.TransactionCount++
		if tx.Amount < 0 {
			summary.CreditAmount += -tx.Amount
		} else {
			summary.GrossExpense += tx.Amount
		}
		if strings.TrimSpace(tx.Category) == "" || strings.EqualFold(tx.Category, "Other") {
			summary.UncategorizedCount++
		}
	}
	summary.TotalAmount = summary.GrossExpense - summary.CreditAmount
	if summary.TransactionCount > 0 {
		summary.AverageAmount = summary.GrossExpense / float64(summary.TransactionCount)
	}

	return summary, nil
}

func (s *ReportingService) GetCategoryBreakdown(period string) ([]BreakdownItem, error) {
	return s.groupBreakdown(period, func(tx models.Transaction) string {
		if tx.Category == "" {
			return "Other"
		}
		return tx.Category
	})
}

func (s *ReportingService) GetSourceBreakdown(period string) ([]BreakdownItem, error) {
	return s.groupBreakdown(period, func(tx models.Transaction) string {
		if strings.HasPrefix(tx.Type, "HDFC") {
			return "HDFC"
		}
		if tx.Type == "" {
			return "Unknown"
		}
		return tx.Type
	})
}

func (s *ReportingService) GetDailyTrend(period string) ([]TrendPoint, error) {
	txs, err := s.filteredTransactions(period)
	if err != nil {
		return nil, err
	}

	byDay := make(map[string]*TrendPoint)
	for _, tx := range txs {
		day := tx.DateTime.Format("2006-01-02")
		point, exists := byDay[day]
		if !exists {
			point = &TrendPoint{Date: day}
			byDay[day] = point
		}
		point.Amount += tx.Amount
		point.Count++
	}

	points := make([]TrendPoint, 0, len(byDay))
	for _, point := range byDay {
		points = append(points, *point)
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points, nil
}

func (s *ReportingService) GetLastNDaysTrend(days int) ([]TrendPoint, error) {
	if days <= 0 {
		days = 10
	}

	txs, err := s.dbClient.FetchAllTransactions()
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -(days - 1))
	cutoff = time.Date(cutoff.Year(), cutoff.Month(), cutoff.Day(), 0, 0, 0, 0, time.UTC)

	byDay := make(map[string]*TrendPoint)
	for _, tx := range txs {
		ts := tx.DateTime.UTC()
		if ts.Before(cutoff) {
			continue
		}

		day := ts.Format("2006-01-02")
		point, exists := byDay[day]
		if !exists {
			point = &TrendPoint{Date: day}
			byDay[day] = point
		}
		point.Amount += tx.Amount
		point.Count++
	}

	points := make([]TrendPoint, 0, len(byDay))
	for _, point := range byDay {
		points = append(points, *point)
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points, nil
}

func (s *ReportingService) GetLastNDaysTransactions(days int, limit int) ([]models.Transaction, error) {
	if days <= 0 {
		days = 10
	}

	fmt.Printf("Fetching last %d days of transactions...\n", days)

	txs, err := s.dbClient.FetchAllTransactions()
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -(days - 1))
	cutoff = time.Date(cutoff.Year(), cutoff.Month(), cutoff.Day(), 0, 0, 0, 0, time.UTC)

	filtered := make([]models.Transaction, 0, len(txs))

	for _, tx := range txs {
		ts := tx.DateTime.UTC()
		if ts.After(cutoff) || ts.Equal(cutoff) {
			filtered = append(filtered, tx)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].DateTime.After(filtered[j].DateTime)
	})

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	for _, tx := range filtered {
		fmt.Printf("Filtered Transaction: %+v\n", tx)
	}
	return filtered, nil
}

func (s *ReportingService) GetMonthlyComparison() (MonthlyComparison, error) {
	currentMonthTxs, err := s.filteredTransactions("THIS_MONTH")
	if err != nil {
		return MonthlyComparison{}, err
	}

	lastMonthTxs, err := s.filteredTransactions("LAST_MONTH")
	if err != nil {
		return MonthlyComparison{}, err
	}

	comparison := MonthlyComparison{}
	topMerchantTotals := make(map[string]float64)

	for _, tx := range currentMonthTxs {
		comparison.CurrentMonthAmount += tx.Amount
		comparison.CurrentMonthCount++
		if tx.Vendor != "" && tx.Amount > 0 {
			topMerchantTotals[tx.Vendor] += tx.Amount
		}
	}

	for _, tx := range lastMonthTxs {
		comparison.LastMonthAmount += tx.Amount
		comparison.LastMonthCount++
	}

	comparison.DeltaAmount = comparison.CurrentMonthAmount - comparison.LastMonthAmount
	if comparison.LastMonthAmount > 0 {
		comparison.DeltaPercent = (comparison.DeltaAmount / comparison.LastMonthAmount) * 100
	}

	for merchant, amount := range topMerchantTotals {
		if amount > comparison.TopMerchantSpend {
			comparison.TopMerchantSpend = amount
			comparison.TopMerchantThisMonth = merchant
		}
	}

	return comparison, nil
}

func (s *ReportingService) ListTransactionsByDateRange(from, to time.Time, category string, limit int) ([]models.Transaction, error) {
	txs, err := s.dbClient.FetchAllTransactions()
	if err != nil {
		return nil, err
	}

	filtered := make([]models.Transaction, 0, len(txs))
	for _, tx := range txs {
		ts := tx.DateTime.UTC()
		if (ts.After(from) || ts.Equal(from)) && (ts.Before(to) || ts.Equal(to)) {
			filtered = append(filtered, tx)
		}
	}

	category = strings.TrimSpace(category)
	if category != "" {
		byCat := filtered[:0]
		for _, tx := range filtered {
			txCategory := strings.TrimSpace(tx.Category)
			if txCategory == "" {
				txCategory = "Other"
			}
			if strings.EqualFold(txCategory, category) {
				byCat = append(byCat, tx)
			}
		}
		filtered = byCat
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].DateTime.After(filtered[j].DateTime)
	})

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

func (s *ReportingService) groupBreakdown(period string, keyFn func(models.Transaction) string) ([]BreakdownItem, error) {
	txs, err := s.filteredTransactions(period)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string]*BreakdownItem)
	for _, tx := range txs {
		key := keyFn(tx)
		item, exists := grouped[key]
		if !exists {
			item = &BreakdownItem{Label: key}
			grouped[key] = item
		}
		item.Amount += tx.Amount
		item.Count++
	}

	items := make([]BreakdownItem, 0, len(grouped))
	for _, item := range grouped {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Amount > items[j].Amount
	})

	return items, nil
}

func (s *ReportingService) filteredTransactions(period string) ([]models.Transaction, error) {
	txs, err := s.dbClient.FetchAllTransactions()
	if err != nil {
		return nil, err
	}

	start, end, err := utils.ResolvePeriod(normalizePeriod(period))
	if err != nil {
		return nil, err
	}

	filtered := make([]models.Transaction, 0, len(txs))
	for _, tx := range txs {
		ts := tx.DateTime.UTC()
		if (ts.After(start) || ts.Equal(start)) && (ts.Before(end) || ts.Equal(end)) {
			filtered = append(filtered, tx)
		}
	}

	return filtered, nil
}

func normalizePeriod(period string) string {
	value := strings.TrimSpace(strings.ToUpper(period))
	if value == "" {
		return "THIS_MONTH"
	}
	return value
}
