package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/yourusername/expense-tracker/models"
	"github.com/yourusername/expense-tracker/utils"
)

// AggregationService handles data aggregation operations
type AggregationService struct {
	dbClient models.DatabaseClient
}

// NewAggregationService creates a new aggregation service
func NewAggregationService(dbClient models.DatabaseClient) *AggregationService {
	return &AggregationService{
		dbClient: dbClient,
	}
}

// GetTotalSpend calculates total spending for a given period
func (s *AggregationService) GetTotalSpend(ctx context.Context, period string) (SpendResult, error) {
	start, end, err := utils.ResolvePeriod(period)
	if err != nil {
		return SpendResult{}, err
	}

	transactions, err := s.fetchTransactionsInRange(ctx, start, end)
	if err != nil {
		return SpendResult{}, err
	}

	var totalSpent float64
	for _, tx := range transactions {
		totalSpent += tx.Amount
	}

	return SpendResult{
		Period:     period,
		TotalSpent: totalSpent,
	}, nil
}

// GetCategorySpend calculates spending for a specific category in a period
func (s *AggregationService) GetCategorySpend(ctx context.Context, category, period string) (CategorySpendResult, error) {
	if category == "" {
		return CategorySpendResult{}, fmt.Errorf("category is required")
	}

	start, end, err := utils.ResolvePeriod(period)
	if err != nil {
		return CategorySpendResult{}, err
	}

	transactions, err := s.fetchTransactionsInRange(ctx, start, end)
	if err != nil {
		return CategorySpendResult{}, err
	}

	var totalSpent float64
	var count int
	for _, tx := range transactions {
		if tx.Category == category {
			totalSpent += tx.Amount
			count++
		}
	}

	var average float64
	if count > 0 {
		average = totalSpent / float64(count)
	}

	return CategorySpendResult{
		Category:   category,
		Period:     period,
		TotalSpent: totalSpent,
		Average:    average,
	}, nil
}

// CompareCategories compares spending between two categories in a period
func (s *AggregationService) CompareCategories(ctx context.Context, c1, c2, period string) (map[string]float64, error) {
	if c1 == "" || c2 == "" {
		return nil, fmt.Errorf("both categories are required")
	}

	start, end, err := utils.ResolvePeriod(period)
	if err != nil {
		return nil, err
	}

	transactions, err := s.fetchTransactionsInRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	result := make(map[string]float64)
	var cat1Total, cat2Total float64

	for _, tx := range transactions {
		if tx.Category == c1 {
			cat1Total += tx.Amount
		} else if tx.Category == c2 {
			cat2Total += tx.Amount
		}
	}

	result[c1] = cat1Total
	result[c2] = cat2Total

	return result, nil
}

// ComparePeriods compares spending between two periods
func (s *AggregationService) ComparePeriods(ctx context.Context, p1, p2 string) (ComparisonResult, error) {
	start1, end1, err := utils.ResolvePeriod(p1)
	if err != nil {
		return ComparisonResult{}, err
	}

	start2, end2, err := utils.ResolvePeriod(p2)
	if err != nil {
		return ComparisonResult{}, err
	}

	transactions1, err := s.fetchTransactionsInRange(ctx, start1, end1)
	if err != nil {
		return ComparisonResult{}, err
	}

	transactions2, err := s.fetchTransactionsInRange(ctx, start2, end2)
	if err != nil {
		return ComparisonResult{}, err
	}

	var amount1, amount2 float64
	for _, tx := range transactions1 {
		amount1 += tx.Amount
	}
	for _, tx := range transactions2 {
		amount2 += tx.Amount
	}

	var deltaPercent float64
	if amount1 > 0 {
		deltaPercent = ((amount2 - amount1) / amount1) * 100
	} else if amount2 > 0 {
		deltaPercent = 100.0
	}

	return ComparisonResult{
		BasePeriod:    p1,
		ComparePeriod: p2,
		BaseAmount:    amount1,
		CompareAmount: amount2,
		DeltaPercent:  deltaPercent,
	}, nil
}

// GetTopMerchants gets top merchants by spending amount for a period
func (s *AggregationService) GetTopMerchants(ctx context.Context, period string, limit int) (TopMerchantsResult, error) {
	if limit <= 0 {
		limit = 10 // default limit
	}

	start, end, err := utils.ResolvePeriod(period)
	if err != nil {
		return TopMerchantsResult{}, err
	}

	transactions, err := s.fetchTransactionsInRange(ctx, start, end)
	if err != nil {
		return TopMerchantsResult{}, err
	}

	merchantTotals := make(map[string]float64)
	for _, tx := range transactions {
		if tx.Vendor != "" {
			merchantTotals[tx.Vendor] += tx.Amount
		}
	}

	// Sort merchants by total spending
	type merchantAmount struct {
		merchant string
		amount   float64
	}
	var sorted []merchantAmount
	for merchant, amount := range merchantTotals {
		sorted = append(sorted, merchantAmount{merchant: merchant, amount: amount})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].amount > sorted[j].amount
	})

	// Take top N merchants
	topMerchants := make(map[string]float64)
	maxCount := limit
	if len(sorted) < maxCount {
		maxCount = len(sorted)
	}
	for i := 0; i < maxCount; i++ {
		topMerchants[sorted[i].merchant] = sorted[i].amount
	}

	return TopMerchantsResult{
		Period:    period,
		Merchants: topMerchants,
	}, nil
}

// GetDailyTrend gets daily spending trend for a period
func (s *AggregationService) GetDailyTrend(ctx context.Context, period string) (map[string]float64, error) {
	start, end, err := utils.ResolvePeriod(period)
	if err != nil {
		return nil, err
	}

	transactions, err := s.fetchTransactionsInRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	dailyTotals := make(map[string]float64)
	for _, tx := range transactions {
		dateKey := tx.DateTime.Format("2006-01-02")
		dailyTotals[dateKey] += tx.Amount
	}

	return dailyTotals, nil
}

// GetMonthlyTrend gets monthly spending trend
func (s *AggregationService) GetMonthlyTrend(ctx context.Context, months int) (map[string]float64, error) {
	if months <= 0 {
		months = 12 // default to 12 months
	}

	now := time.Now().UTC()
	end := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	start := end.AddDate(0, -months, 0)

	transactions, err := s.fetchTransactionsInRange(ctx, start, now)
	if err != nil {
		return nil, err
	}

	monthlyTotals := make(map[string]float64)
	for _, tx := range transactions {
		monthKey := tx.DateTime.Format("2006-01")
		monthlyTotals[monthKey] += tx.Amount
	}

	return monthlyTotals, nil
}

// GetAnomalies identifies transactions that are significantly above average (anomalies)
func (s *AggregationService) GetAnomalies(ctx context.Context, period string) (map[string]interface{}, error) {
	start, end, err := utils.ResolvePeriod(period)
	if err != nil {
		return nil, err
	}

	transactions, err := s.fetchTransactionsInRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	// Calculate average transaction amount
	var totalAmount float64
	for _, tx := range transactions {
		totalAmount += tx.Amount
	}
	var average float64
	if len(transactions) > 0 {
		average = totalAmount / float64(len(transactions))
	}

	// Standard deviation calculation
	var variance float64
	if len(transactions) > 1 {
		for _, tx := range transactions {
			diff := tx.Amount - average
			variance += diff * diff
		}
		variance = variance / float64(len(transactions))
	}
	stdDev := math.Sqrt(variance)

	// Threshold: transactions > average + 2*stdDev
	threshold := average + (2 * stdDev)
	if threshold < average*2 {
		threshold = average * 2 // Minimum threshold of 2x average
	}

	// Find anomalies
	anomalies := make([]models.Transaction, 0)
	for _, tx := range transactions {
		if tx.Amount > threshold {
			anomalies = append(anomalies, tx)
		}
	}

	result := map[string]interface{}{
		"period":        period,
		"average":       average,
		"threshold":     threshold,
		"anomaly_count": len(anomalies),
		"anomalies":     anomalies,
	}

	return result, nil
}

// fetchTransactionsInRange fetches transactions within a date range
func (s *AggregationService) fetchTransactionsInRange(ctx context.Context, start, end time.Time) ([]models.Transaction, error) {
	allTransactions, err := s.dbClient.FetchAllTransactions()
	if err != nil {
		return nil, err
	}

	var filtered []models.Transaction
	for _, tx := range allTransactions {
		// Convert to UTC for comparison
		txTime := tx.DateTime.UTC()
		if (txTime.After(start) || txTime.Equal(start)) && (txTime.Before(end) || txTime.Equal(end)) {
			filtered = append(filtered, tx)
		}
	}

	return filtered, nil
}
