package main

import (
	"fmt"
	"strings"
)

// BuildPrompt builds a prompt string from an intent type and payload
func BuildPrompt(intent IntentType, payload interface{}) (string, error) {
	switch intent {
	case CATEGORY_SUMMARY:
		p, ok := payload.(CategoryInsightPayload)
		if !ok {
			return "", fmt.Errorf("invalid payload type for CATEGORY_SUMMARY: expected CategoryInsightPayload")
		}
		if err := validateCategoryPayload(p); err != nil {
			return "", err
		}
		return buildCategoryPrompt(p), nil

	case TOTAL_SPEND:
		p, ok := payload.(TotalSpendPayload)
		if !ok {
			return "", fmt.Errorf("invalid payload type for TOTAL_SPEND: expected TotalSpendPayload")
		}
		if err := validateTotalSpendPayload(p); err != nil {
			return "", err
		}
		return buildTotalSpendPrompt(p), nil

	case PERIOD_COMPARISON, CATEGORY_COMPARISON:
		p, ok := payload.(ComparisonPayload)
		if !ok {
			return "", fmt.Errorf("invalid payload type for comparison: expected ComparisonPayload")
		}
		if err := validateComparisonPayload(p); err != nil {
			return "", err
		}
		return buildComparisonPrompt(p), nil

	case TOP_MERCHANTS:
		p, ok := payload.(TopMerchantsPayload)
		if !ok {
			return "", fmt.Errorf("invalid payload type for TOP_MERCHANTS: expected TopMerchantsPayload")
		}
		if err := validateTopMerchantsPayload(p); err != nil {
			return "", err
		}
		return buildTopMerchantsPrompt(p), nil

	case DAILY_TREND, MONTHLY_TREND:
		p, ok := payload.(TrendPayload)
		if !ok {
			return "", fmt.Errorf("invalid payload type for trend: expected TrendPayload")
		}
		if err := validateTrendPayload(p); err != nil {
			return "", err
		}
		return buildTrendPrompt(p), nil

	case GENERAL_INSIGHT, BUDGET_STATUS, ANOMALY_EXPLANATION:
		p, ok := payload.(GeneralInsightPayload)
		if !ok {
			return "", fmt.Errorf("invalid payload type for general insight: expected GeneralInsightPayload")
		}
		if err := validateGeneralPayload(p); err != nil {
			return "", err
		}
		return buildGeneralPrompt(p), nil

	default:
		return "", fmt.Errorf("unsupported intent type: %s", intent.String())
	}
}

// Validation functions
func validateCategoryPayload(p CategoryInsightPayload) error {
	if p.Category == "" {
		return fmt.Errorf("category is required")
	}
	if p.Period == "" {
		return fmt.Errorf("period is required")
	}
	return nil
}

func validateTotalSpendPayload(p TotalSpendPayload) error {
	if p.Period == "" {
		return fmt.Errorf("period is required")
	}
	return nil
}

func validateComparisonPayload(p ComparisonPayload) error {
	if p.BasePeriod == "" {
		return fmt.Errorf("base_period is required")
	}
	if p.ComparePeriod == "" {
		return fmt.Errorf("compare_period is required")
	}
	return nil
}

func validateTopMerchantsPayload(p TopMerchantsPayload) error {
	if p.Period == "" {
		return fmt.Errorf("period is required")
	}
	if len(p.Merchants) == 0 {
		return fmt.Errorf("merchants data is required")
	}
	return nil
}

func validateTrendPayload(p TrendPayload) error {
	if p.Period == "" {
		return fmt.Errorf("period is required")
	}
	if len(p.TrendData) == 0 {
		return fmt.Errorf("trend_data is required")
	}
	return nil
}

func validateGeneralPayload(p GeneralInsightPayload) error {
	if p.FactsSummary == "" {
		return fmt.Errorf("facts_summary is required")
	}
	return nil
}

// Template builders
func buildCategoryPrompt(p CategoryInsightPayload) string {
	prompt := CategorySummaryTemplate
	prompt = strings.ReplaceAll(prompt, "{{Category}}", p.Category)
	prompt = strings.ReplaceAll(prompt, "{{Period}}", p.Period)
	prompt = strings.ReplaceAll(prompt, "{{TotalSpent}}", formatFloat(p.TotalSpent))
	prompt = strings.ReplaceAll(prompt, "{{AverageSpent}}", formatFloat(p.AverageSpent))
	prompt = strings.ReplaceAll(prompt, "{{Budget}}", formatFloat(p.Budget))
	prompt = strings.ReplaceAll(prompt, "{{BudgetExceeded}}", formatBool(p.BudgetExceeded))
	prompt = strings.ReplaceAll(prompt, "{{DeltaPercent}}", formatFloat(p.DeltaPercent))
	prompt = strings.ReplaceAll(prompt, "{{UserQuestion}}", p.UserQuestion)
	return prompt
}

func buildTotalSpendPrompt(p TotalSpendPayload) string {
	prompt := TotalSpendTemplate
	prompt = strings.ReplaceAll(prompt, "{{Period}}", p.Period)
	prompt = strings.ReplaceAll(prompt, "{{TotalSpent}}", formatFloat(p.TotalSpent))
	prompt = strings.ReplaceAll(prompt, "{{Average}}", formatFloat(p.Average))
	prompt = strings.ReplaceAll(prompt, "{{UserQuestion}}", p.UserQuestion)
	return prompt
}

func buildComparisonPrompt(p ComparisonPayload) string {
	prompt := ComparisonTemplate
	prompt = strings.ReplaceAll(prompt, "{{BasePeriod}}", p.BasePeriod)
	prompt = strings.ReplaceAll(prompt, "{{ComparePeriod}}", p.ComparePeriod)
	prompt = strings.ReplaceAll(prompt, "{{BaseAmount}}", formatFloat(p.BaseAmount))
	prompt = strings.ReplaceAll(prompt, "{{CompareAmount}}", formatFloat(p.CompareAmount))
	prompt = strings.ReplaceAll(prompt, "{{DeltaPercent}}", formatFloat(p.DeltaPercent))
	prompt = strings.ReplaceAll(prompt, "{{UserQuestion}}", p.UserQuestion)
	return prompt
}

func buildTopMerchantsPrompt(p TopMerchantsPayload) string {
	prompt := TopMerchantsTemplate
	prompt = strings.ReplaceAll(prompt, "{{Period}}", p.Period)

	// Build merchants list
	var merchantsList strings.Builder
	for merchant, amount := range p.Merchants {
		fmt.Fprintf(&merchantsList, "- %s: ₹%s\n", merchant, formatFloat(amount))
	}
	prompt = strings.ReplaceAll(prompt, "{{MerchantsList}}", merchantsList.String())
	prompt = strings.ReplaceAll(prompt, "{{UserQuestion}}", p.UserQuestion)
	return prompt
}

func buildTrendPrompt(p TrendPayload) string {
	prompt := TrendTemplate
	prompt = strings.ReplaceAll(prompt, "{{Period}}", p.Period)

	// Build trend data list
	var trendList strings.Builder
	for key, value := range p.TrendData {
		fmt.Fprintf(&trendList, "- %s: ₹%s\n", key, formatFloat(value))
	}
	prompt = strings.ReplaceAll(prompt, "{{TrendDataList}}", trendList.String())
	prompt = strings.ReplaceAll(prompt, "{{UserQuestion}}", p.UserQuestion)
	return prompt
}

func buildGeneralPrompt(p GeneralInsightPayload) string {
	prompt := GeneralInsightTemplate
	prompt = strings.ReplaceAll(prompt, "{{FactsSummary}}", p.FactsSummary)
	prompt = strings.ReplaceAll(prompt, "{{UserQuestion}}", p.UserQuestion)
	return prompt
}

// Helper functions
func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
