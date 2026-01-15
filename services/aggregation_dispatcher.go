package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/yourusername/expense-tracker/ai"
	"github.com/yourusername/expense-tracker/models"
)

// ExecuteAggregation routes an intent to the appropriate aggregation function
func ExecuteAggregation(ctx context.Context, intent ai.ExpenseIntent, dbClient models.DatabaseClient) (interface{}, error) {
	service := NewAggregationService(dbClient)

	// Validate required parameters before querying
	if err := validateIntent(intent); err != nil {
		return nil, err
	}

	// Get period string (default to THIS_MONTH if not specified)
	period := "THIS_MONTH"
	if intent.Period != nil {
		period = intent.Period.String()
	}

	switch intent.IntentType {
	case ai.TOTAL_SPEND:
		return service.GetTotalSpend(ctx, period)

	case ai.CATEGORY_SUMMARY:
		category := intent.Category
		if category == "" {
			// Try to get from parameters
			if intent.Parameters != nil {
				category = intent.Parameters["category"]
			}
		}
		if category == "" {
			return nil, errors.New("category is required for CATEGORY_SUMMARY intent")
		}
		return service.GetCategorySpend(ctx, category, period)

	case ai.CATEGORY_COMPARISON:
		var cat1, cat2 string
		if intent.Parameters != nil {
			cat1 = intent.Parameters["category1"]
			cat2 = intent.Parameters["category2"]
		}
		// Also check if single category is provided and compare with all others
		if cat1 == "" && intent.Category != "" {
			cat1 = intent.Category
			cat2 = intent.Parameters["category2"]
		}
		if cat1 == "" || cat2 == "" {
			return nil, errors.New("both category1 and category2 are required for CATEGORY_COMPARISON intent")
		}
		return service.CompareCategories(ctx, cat1, cat2, period)

	case ai.PERIOD_COMPARISON:
		var period1, period2 string
		if intent.Parameters != nil {
			period1 = intent.Parameters["period1"]
			period2 = intent.Parameters["period2"]
		}
		if period1 == "" || period2 == "" {
			return nil, errors.New("both period1 and period2 are required for PERIOD_COMPARISON intent")
		}
		return service.ComparePeriods(ctx, period1, period2)

	case ai.TOP_MERCHANTS:
		return service.GetTopMerchants(ctx, period, 10)

	case ai.DAILY_TREND:
		return service.GetDailyTrend(ctx, period)

	case ai.MONTHLY_TREND:
		months := 12
		if intent.Parameters != nil {
			if monthsStr := intent.Parameters["months"]; monthsStr != "" {
				fmt.Sscanf(monthsStr, "%d", &months)
			}
		}
		return service.GetMonthlyTrend(ctx, months)

	case ai.ANOMALY_EXPLANATION:
		// Return anomaly data (transactions that are significantly above average)
		return service.GetAnomalies(ctx, period)

	case ai.BUDGET_STATUS:
		// For now, return spending data - budget logic would need budget definitions
		return service.GetTotalSpend(ctx, period)

	case ai.GENERAL_INSIGHT:
		// Return total spend as a general insight
		return service.GetTotalSpend(ctx, period)

	default:
		return nil, fmt.Errorf("unsupported intent type: %s", intent.IntentType.String())
	}
}

// validateIntent validates that required parameters are present
func validateIntent(intent ai.ExpenseIntent) error {
	switch intent.IntentType {
	case ai.CATEGORY_SUMMARY:
		category := intent.Category
		if category == "" && intent.Parameters != nil {
			category = intent.Parameters["category"]
		}
		if category == "" {
			return errors.New("category is required for CATEGORY_SUMMARY intent")
		}

	case ai.CATEGORY_COMPARISON:
		var cat1, cat2 string
		if intent.Parameters != nil {
			cat1 = intent.Parameters["category1"]
			cat2 = intent.Parameters["category2"]
		}
		if cat1 == "" && intent.Category != "" {
			cat1 = intent.Category
		}
		if cat1 == "" || cat2 == "" {
			return errors.New("both category1 and category2 are required for CATEGORY_COMPARISON intent")
		}

	case ai.PERIOD_COMPARISON:
		var period1, period2 string
		if intent.Parameters != nil {
			period1 = intent.Parameters["period1"]
			period2 = intent.Parameters["period2"]
		}
		if period1 == "" || period2 == "" {
			return errors.New("both period1 and period2 are required for PERIOD_COMPARISON intent")
		}
	}

	return nil
}
