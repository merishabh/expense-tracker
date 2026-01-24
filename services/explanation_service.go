package services

import (
	"fmt"
	"os"

	"github.com/yourusername/expense-tracker/ai"
)

// ExplanationService handles converting aggregation results to explanations
type ExplanationService struct {
	groqClient *ai.GroqClient
}

// NewExplanationService creates a new explanation service using Groq
func NewExplanationService(apiKey string) (*ExplanationService, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY is required")
	}

	groqClient := ai.NewGroqClient(apiKey)

	return &ExplanationService{
		groqClient: groqClient,
	}, nil
}

// ExplainAggregation converts an aggregation result to a human-readable explanation
func (s *ExplanationService) ExplainAggregation(
	intent ai.IntentType,
	aggregationResult interface{},
	userQuestion string,
) (string, error) {
	// Convert aggregation result to payload
	payload, err := convertToPayload(intent, aggregationResult, userQuestion)
	if err != nil {
		return "", fmt.Errorf("failed to convert aggregation result: %v", err)
	}

	// Build prompt
	prompt, err := ai.BuildPrompt(intent, payload)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt: %v", err)
	}

	// Generate explanation using Groq
	explanation, err := s.groqClient.GenerateExplanation(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate explanation: %v", err)
	}

	return explanation, nil
}

// convertToPayload converts aggregation results to payload structs
func convertToPayload(intent ai.IntentType, result interface{}, userQuestion string) (interface{}, error) {
	switch intent {
	case ai.TOTAL_SPEND:
		spendResult, ok := result.(SpendResult)
		if !ok {
			return nil, fmt.Errorf("invalid result type for TOTAL_SPEND: expected SpendResult")
		}
		return ai.TotalSpendPayload{
			Period:       spendResult.Period,
			TotalSpent:   spendResult.TotalSpent,
			Average:      spendResult.TotalSpent, // For total spend, average equals total
			UserQuestion: userQuestion,
		}, nil

	case ai.CATEGORY_SUMMARY:
		// Check if result is VendorSpendResult (vendor query) or CategorySpendResult (category query)
		if vendorResult, ok := result.(VendorSpendResult); ok {
			// Vendor-specific query
			return ai.CategoryInsightPayload{
				Category:       vendorResult.Vendor, // Use vendor name as category for explanation
				Period:         vendorResult.Period,
				TotalSpent:     vendorResult.TotalSpent,
				AverageSpent:   vendorResult.Average,
				Budget:         0, // Budget not available in aggregation result
				BudgetExceeded: false,
				DeltaPercent:   0, // Delta not calculated in aggregation
				UserQuestion:   userQuestion,
			}, nil
		}

		categoryResult, ok := result.(CategorySpendResult)
		if !ok {
			return nil, fmt.Errorf("invalid result type for CATEGORY_SUMMARY: expected CategorySpendResult or VendorSpendResult")
		}
		return ai.CategoryInsightPayload{
			Category:       categoryResult.Category,
			Period:         categoryResult.Period,
			TotalSpent:     categoryResult.TotalSpent,
			AverageSpent:   categoryResult.Average,
			Budget:         0, // Budget not available in aggregation result
			BudgetExceeded: false,
			DeltaPercent:   0, // Delta not calculated in aggregation
			UserQuestion:   userQuestion,
		}, nil

	case ai.PERIOD_COMPARISON, ai.CATEGORY_COMPARISON:
		comparisonResult, ok := result.(ComparisonResult)
		if !ok {
			return nil, fmt.Errorf("invalid result type for comparison: expected ComparisonResult")
		}
		return ai.ComparisonPayload{
			BasePeriod:    comparisonResult.BasePeriod,
			ComparePeriod: comparisonResult.ComparePeriod,
			BaseAmount:    comparisonResult.BaseAmount,
			CompareAmount: comparisonResult.CompareAmount,
			DeltaPercent:  comparisonResult.DeltaPercent,
			UserQuestion:  userQuestion,
		}, nil

	case ai.TOP_MERCHANTS:
		merchantsResult, ok := result.(TopMerchantsResult)
		if !ok {
			return nil, fmt.Errorf("invalid result type for TOP_MERCHANTS: expected TopMerchantsResult")
		}
		return ai.TopMerchantsPayload{
			Period:       merchantsResult.Period,
			Merchants:    merchantsResult.Merchants,
			UserQuestion: userQuestion,
		}, nil

	case ai.DAILY_TREND, ai.MONTHLY_TREND:
		trendData, ok := result.(map[string]float64)
		if !ok {
			return nil, fmt.Errorf("invalid result type for trend: expected map[string]float64")
		}
		// Period information is not included in trend results, use default
		return ai.TrendPayload{
			Period:       "specified period",
			TrendData:    trendData,
			UserQuestion: userQuestion,
		}, nil

	case ai.ANOMALY_EXPLANATION:
		// Anomalies return map[string]interface{}, convert to general insight
		anomalyData, ok := result.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid result type for ANOMALY_EXPLANATION: expected map[string]interface{}")
		}
		factsSummary := formatAnomalyFacts(anomalyData)
		return ai.GeneralInsightPayload{
			FactsSummary: factsSummary,
			UserQuestion: userQuestion,
		}, nil

	case ai.BUDGET_STATUS, ai.GENERAL_INSIGHT:
		// For these, we need to format the result as facts summary
		factsSummary := formatGeneralFacts(result)
		return ai.GeneralInsightPayload{
			FactsSummary: factsSummary,
			UserQuestion: userQuestion,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported intent type for explanation: %s", intent.String())
	}
}

// formatAnomalyFacts formats anomaly data as a facts summary string
func formatAnomalyFacts(data map[string]interface{}) string {
	var summary string
	if period, ok := data["period"].(string); ok {
		summary += fmt.Sprintf("Period: %s\n", period)
	}
	if avg, ok := data["average"].(float64); ok {
		summary += fmt.Sprintf("Average spending: ₹%.2f\n", avg)
	}
	if threshold, ok := data["threshold"].(float64); ok {
		summary += fmt.Sprintf("Anomaly threshold: ₹%.2f\n", threshold)
	}
	if count, ok := data["anomaly_count"].(int); ok {
		summary += fmt.Sprintf("Number of anomalies: %d\n", count)
	}
	return summary
}

// formatGeneralFacts formats general result data as a facts summary string
func formatGeneralFacts(result interface{}) string {
	// Try to format common result types
	switch v := result.(type) {
	case SpendResult:
		return fmt.Sprintf("Period: %s\nTotal spent: ₹%.2f", v.Period, v.TotalSpent)
	case CategorySpendResult:
		return fmt.Sprintf("Category: %s\nPeriod: %s\nTotal spent: ₹%.2f\nAverage: ₹%.2f",
			v.Category, v.Period, v.TotalSpent, v.Average)
	case ComparisonResult:
		return fmt.Sprintf("Base period: %s (₹%.2f)\nCompare period: %s (₹%.2f)\nChange: %.2f%%",
			v.BasePeriod, v.BaseAmount, v.ComparePeriod, v.CompareAmount, v.DeltaPercent)
	default:
		return fmt.Sprintf("Data: %v", result)
	}
}

// Close closes the explanation service
func (s *ExplanationService) Close() error {
	// GroqClient doesn't need explicit closing (uses HTTP client)
	return nil
}

// NewExplanationServiceFromEnv creates an explanation service using API key from environment
func NewExplanationServiceFromEnv() (*ExplanationService, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY environment variable is required")
	}
	return NewExplanationService(apiKey)
}
