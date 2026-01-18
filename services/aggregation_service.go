package services

import (
	"context"
	"fmt"
	"log"
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
	log.Printf("[GetTotalSpend] Starting aggregation for period: %s", period)

	start, end, err := utils.ResolvePeriod(period)
	if err != nil {
		log.Printf("[GetTotalSpend] Error resolving period %s: %v", period, err)
		return SpendResult{}, err
	}
	log.Printf("[GetTotalSpend] Resolved period range: %s to %s", start.Format(time.RFC3339), end.Format(time.RFC3339))

	transactions, err := s.fetchTransactionsInRange(ctx, start, end)
	if err != nil {
		log.Printf("[GetTotalSpend] Error fetching transactions: %v", err)
		return SpendResult{}, err
	}
	log.Printf("[GetTotalSpend] Fetched %d transactions in range", len(transactions))

	var totalSpent float64
	for _, tx := range transactions {
		totalSpent += tx.Amount
	}

	log.Printf("[GetTotalSpend] Calculated total spend: ₹%.2f for period %s", totalSpent, period)

	return SpendResult{
		Period:     period,
		TotalSpent: totalSpent,
	}, nil
}

// GetCategorySpend calculates spending for a specific category in a period
func (s *AggregationService) GetCategorySpend(ctx context.Context, category, period string) (CategorySpendResult, error) {
	log.Printf("[GetCategorySpend] Starting aggregation for category: %s, period: %s", category, period)

	if category == "" {
		log.Printf("[GetCategorySpend] Error: category is required")
		return CategorySpendResult{}, fmt.Errorf("category is required")
	}

	start, end, err := utils.ResolvePeriod(period)
	if err != nil {
		log.Printf("[GetCategorySpend] Error resolving period %s: %v", period, err)
		return CategorySpendResult{}, err
	}
	log.Printf("[GetCategorySpend] Resolved period range: %s to %s", start.Format(time.RFC3339), end.Format(time.RFC3339))

	transactions, err := s.fetchTransactionsInRange(ctx, start, end)
	if err != nil {
		log.Printf("[GetCategorySpend] Error fetching transactions: %v", err)
		return CategorySpendResult{}, err
	}
	log.Printf("[GetCategorySpend] Fetched %d total transactions in range", len(transactions))

	var totalSpent float64
	var count int
	for _, tx := range transactions {
		if tx.Category == category {
			totalSpent += tx.Amount
			count++
		}
	}

	log.Printf("[GetCategorySpend] Found %d transactions in category '%s'", count, category)

	var average float64
	if count > 0 {
		average = totalSpent / float64(count)
		log.Printf("[GetCategorySpend] Calculated average: ₹%.2f per transaction", average)
	} else {
		log.Printf("[GetCategorySpend] No transactions found for category '%s'", category)
	}

	log.Printf("[GetCategorySpend] Total spent in category '%s': ₹%.2f", category, totalSpent)

	return CategorySpendResult{
		Category:   category,
		Period:     period,
		TotalSpent: totalSpent,
		Average:    average,
	}, nil
}

// CompareCategories compares spending between two categories in a period
func (s *AggregationService) CompareCategories(ctx context.Context, c1, c2, period string) (map[string]float64, error) {
	log.Printf("[CompareCategories] Comparing categories '%s' vs '%s' for period: %s", c1, c2, period)

	if c1 == "" || c2 == "" {
		log.Printf("[CompareCategories] Error: both categories are required")
		return nil, fmt.Errorf("both categories are required")
	}

	start, end, err := utils.ResolvePeriod(period)
	if err != nil {
		log.Printf("[CompareCategories] Error resolving period %s: %v", period, err)
		return nil, err
	}
	log.Printf("[CompareCategories] Resolved period range: %s to %s", start.Format(time.RFC3339), end.Format(time.RFC3339))

	transactions, err := s.fetchTransactionsInRange(ctx, start, end)
	if err != nil {
		log.Printf("[CompareCategories] Error fetching transactions: %v", err)
		return nil, err
	}
	log.Printf("[CompareCategories] Fetched %d transactions in range", len(transactions))

	result := make(map[string]float64)
	var cat1Total, cat2Total float64
	var cat1Count, cat2Count int

	for _, tx := range transactions {
		if tx.Category == c1 {
			cat1Total += tx.Amount
			cat1Count++
		} else if tx.Category == c2 {
			cat2Total += tx.Amount
			cat2Count++
		}
	}

	result[c1] = cat1Total
	result[c2] = cat2Total

	log.Printf("[CompareCategories] Category '%s': ₹%.2f (%d transactions)", c1, cat1Total, cat1Count)
	log.Printf("[CompareCategories] Category '%s': ₹%.2f (%d transactions)", c2, cat2Total, cat2Count)

	return result, nil
}

// ComparePeriods compares spending between two periods
func (s *AggregationService) ComparePeriods(ctx context.Context, p1, p2 string) (ComparisonResult, error) {
	log.Printf("[ComparePeriods] Comparing periods: %s vs %s", p1, p2)

	start1, end1, err := utils.ResolvePeriod(p1)
	if err != nil {
		log.Printf("[ComparePeriods] Error resolving period %s: %v", p1, err)
		return ComparisonResult{}, err
	}
	log.Printf("[ComparePeriods] Period 1 (%s) range: %s to %s", p1, start1.Format(time.RFC3339), end1.Format(time.RFC3339))

	start2, end2, err := utils.ResolvePeriod(p2)
	if err != nil {
		log.Printf("[ComparePeriods] Error resolving period %s: %v", p2, err)
		return ComparisonResult{}, err
	}
	log.Printf("[ComparePeriods] Period 2 (%s) range: %s to %s", p2, start2.Format(time.RFC3339), end2.Format(time.RFC3339))

	transactions1, err := s.fetchTransactionsInRange(ctx, start1, end1)
	if err != nil {
		log.Printf("[ComparePeriods] Error fetching transactions for period 1: %v", err)
		return ComparisonResult{}, err
	}
	log.Printf("[ComparePeriods] Fetched %d transactions for period 1", len(transactions1))

	transactions2, err := s.fetchTransactionsInRange(ctx, start2, end2)
	if err != nil {
		log.Printf("[ComparePeriods] Error fetching transactions for period 2: %v", err)
		return ComparisonResult{}, err
	}
	log.Printf("[ComparePeriods] Fetched %d transactions for period 2", len(transactions2))

	var amount1, amount2 float64
	for _, tx := range transactions1 {
		amount1 += tx.Amount
	}
	for _, tx := range transactions2 {
		amount2 += tx.Amount
	}

	log.Printf("[ComparePeriods] Period 1 (%s) total: ₹%.2f", p1, amount1)
	log.Printf("[ComparePeriods] Period 2 (%s) total: ₹%.2f", p2, amount2)

	var deltaPercent float64
	if amount1 > 0 {
		deltaPercent = ((amount2 - amount1) / amount1) * 100
		log.Printf("[ComparePeriods] Calculated delta: %.2f%%", deltaPercent)
	} else if amount2 > 0 {
		deltaPercent = 100.0
		log.Printf("[ComparePeriods] Period 1 had no spending, delta set to 100%%")
	} else {
		log.Printf("[ComparePeriods] Both periods had no spending")
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
	log.Printf("[GetTopMerchants] Fetching top merchants for period: %s, limit: %d", period, limit)

	if limit <= 0 {
		limit = 10 // default limit
		log.Printf("[GetTopMerchants] Limit was 0 or negative, using default: %d", limit)
	}

	start, end, err := utils.ResolvePeriod(period)
	if err != nil {
		log.Printf("[GetTopMerchants] Error resolving period %s: %v", period, err)
		return TopMerchantsResult{}, err
	}
	log.Printf("[GetTopMerchants] Resolved period range: %s to %s", start.Format(time.RFC3339), end.Format(time.RFC3339))

	transactions, err := s.fetchTransactionsInRange(ctx, start, end)
	if err != nil {
		log.Printf("[GetTopMerchants] Error fetching transactions: %v", err)
		return TopMerchantsResult{}, err
	}
	log.Printf("[GetTopMerchants] Fetched %d transactions in range", len(transactions))

	merchantTotals := make(map[string]float64)
	var merchantsWithTransactions int
	for _, tx := range transactions {
		if tx.Vendor != "" {
			merchantTotals[tx.Vendor] += tx.Amount
			merchantsWithTransactions++
		}
	}

	log.Printf("[GetTopMerchants] Found %d unique merchants from %d transactions", len(merchantTotals), merchantsWithTransactions)

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
		log.Printf("[GetTopMerchants] Only %d merchants available, returning all", maxCount)
	}
	for i := 0; i < maxCount; i++ {
		topMerchants[sorted[i].merchant] = sorted[i].amount
		log.Printf("[GetTopMerchants] Top merchant %d: %s - ₹%.2f", i+1, sorted[i].merchant, sorted[i].amount)
	}

	log.Printf("[GetTopMerchants] Returning top %d merchants", len(topMerchants))

	return TopMerchantsResult{
		Period:    period,
		Merchants: topMerchants,
	}, nil
}

// GetDailyTrend gets daily spending trend for a period
func (s *AggregationService) GetDailyTrend(ctx context.Context, period string) (map[string]float64, error) {
	log.Printf("[GetDailyTrend] Fetching daily trend for period: %s", period)

	start, end, err := utils.ResolvePeriod(period)
	if err != nil {
		log.Printf("[GetDailyTrend] Error resolving period %s: %v", period, err)
		return nil, err
	}
	log.Printf("[GetDailyTrend] Resolved period range: %s to %s", start.Format(time.RFC3339), end.Format(time.RFC3339))

	transactions, err := s.fetchTransactionsInRange(ctx, start, end)
	if err != nil {
		log.Printf("[GetDailyTrend] Error fetching transactions: %v", err)
		return nil, err
	}
	log.Printf("[GetDailyTrend] Fetched %d transactions in range", len(transactions))

	dailyTotals := make(map[string]float64)
	for _, tx := range transactions {
		dateKey := tx.DateTime.Format("2006-01-02")
		dailyTotals[dateKey] += tx.Amount
	}

	log.Printf("[GetDailyTrend] Aggregated spending into %d unique days", len(dailyTotals))

	return dailyTotals, nil
}

// GetMonthlyTrend gets monthly spending trend
func (s *AggregationService) GetMonthlyTrend(ctx context.Context, months int) (map[string]float64, error) {
	log.Printf("[GetMonthlyTrend] Fetching monthly trend for %d months", months)

	if months <= 0 {
		months = 12 // default to 12 months
		log.Printf("[GetMonthlyTrend] Months was 0 or negative, using default: %d", months)
	}

	now := time.Now().UTC()
	end := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	start := end.AddDate(0, -months, 0)

	log.Printf("[GetMonthlyTrend] Date range: %s to %s", start.Format(time.RFC3339), end.Format(time.RFC3339))

	transactions, err := s.fetchTransactionsInRange(ctx, start, now)
	if err != nil {
		log.Printf("[GetMonthlyTrend] Error fetching transactions: %v", err)
		return nil, err
	}
	log.Printf("[GetMonthlyTrend] Fetched %d transactions in range", len(transactions))

	monthlyTotals := make(map[string]float64)
	for _, tx := range transactions {
		monthKey := tx.DateTime.Format("2006-01")
		monthlyTotals[monthKey] += tx.Amount
	}

	log.Printf("[GetMonthlyTrend] Aggregated spending into %d unique months", len(monthlyTotals))

	return monthlyTotals, nil
}

// GetAnomalies identifies transactions that are significantly above average (anomalies)
func (s *AggregationService) GetAnomalies(ctx context.Context, period string) (map[string]interface{}, error) {
	log.Printf("[GetAnomalies] Identifying anomalies for period: %s", period)

	start, end, err := utils.ResolvePeriod(period)
	if err != nil {
		log.Printf("[GetAnomalies] Error resolving period %s: %v", period, err)
		return nil, err
	}
	log.Printf("[GetAnomalies] Resolved period range: %s to %s", start.Format(time.RFC3339), end.Format(time.RFC3339))

	transactions, err := s.fetchTransactionsInRange(ctx, start, end)
	if err != nil {
		log.Printf("[GetAnomalies] Error fetching transactions: %v", err)
		return nil, err
	}
	log.Printf("[GetAnomalies] Fetched %d transactions in range", len(transactions))

	if len(transactions) == 0 {
		log.Printf("[GetAnomalies] No transactions found, returning empty result")
		return map[string]interface{}{
			"period":        period,
			"average":       0.0,
			"threshold":     0.0,
			"anomaly_count": 0,
			"anomalies":     []models.Transaction{},
		}, nil
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
	log.Printf("[GetAnomalies] Calculated average transaction amount: ₹%.2f (total: ₹%.2f, count: %d)", average, totalAmount, len(transactions))

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
	log.Printf("[GetAnomalies] Calculated standard deviation: ₹%.2f (variance: %.2f)", stdDev, variance)

	// Threshold: transactions > average + 2*stdDev
	threshold := average + (2 * stdDev)
	if threshold < average*2 {
		threshold = average * 2 // Minimum threshold of 2x average
		log.Printf("[GetAnomalies] Threshold adjusted to minimum 2x average: ₹%.2f", threshold)
	} else {
		log.Printf("[GetAnomalies] Calculated threshold (average + 2*stdDev): ₹%.2f", threshold)
	}

	// Find anomalies
	anomalies := make([]models.Transaction, 0)
	for _, tx := range transactions {
		if tx.Amount > threshold {
			anomalies = append(anomalies, tx)
			log.Printf("[GetAnomalies] Found anomaly: ₹%.2f at %s (vendor: %s)", tx.Amount, tx.DateTime.Format(time.RFC3339), tx.Vendor)
		}
	}

	log.Printf("[GetAnomalies] Identified %d anomalies out of %d transactions", len(anomalies), len(transactions))

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
	log.Printf("[fetchTransactionsInRange] Fetching transactions from database, filtering range: %s to %s", start.Format(time.RFC3339), end.Format(time.RFC3339))

	allTransactions, err := s.dbClient.FetchAllTransactions()
	if err != nil {
		log.Printf("[fetchTransactionsInRange] Error fetching all transactions: %v", err)
		return nil, err
	}
	log.Printf("[fetchTransactionsInRange] Fetched %d total transactions from database", len(allTransactions))

	var filtered []models.Transaction
	for _, tx := range allTransactions {
		// Convert to UTC for comparison
		txTime := tx.DateTime.UTC()
		if (txTime.After(start) || txTime.Equal(start)) && (txTime.Before(end) || txTime.Equal(end)) {
			filtered = append(filtered, tx)
		}
	}

	log.Printf("[fetchTransactionsInRange] Filtered to %d transactions in date range", len(filtered))

	return filtered, nil
}
