package services

import (
	"sort"
	"strings"

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

func NewReportingService(dbClient models.DatabaseClient) *ReportingService {
	return &ReportingService{dbClient: dbClient}
}

func (s *ReportingService) ListTransactions(period string, limit int) ([]models.Transaction, error) {
	txs, err := s.filteredTransactions(period)
	if err != nil {
		return nil, err
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
		summary.TotalAmount += tx.Amount
		summary.TransactionCount++
		if strings.TrimSpace(tx.Category) == "" || strings.EqualFold(tx.Category, "Other") {
			summary.UncategorizedCount++
		}
	}
	if summary.TransactionCount > 0 {
		summary.AverageAmount = summary.TotalAmount / float64(summary.TransactionCount)
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
