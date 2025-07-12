package main

import (
	"fmt"
	"math"
	"sort"
)

// AnalyticsService provides advanced analytics for expense data
type AnalyticsService struct{}

// SpendingInsight represents a key insight about spending patterns
type SpendingInsight struct {
	Type     string  `json:"type"` // "warning", "tip", "trend"
	Category string  `json:"category"`
	Message  string  `json:"message"`
	Impact   float64 `json:"impact"`   // Monetary impact
	Severity string  `json:"severity"` // "low", "medium", "high"
}

// BudgetRecommendation provides budget suggestions
type BudgetRecommendation struct {
	Category          string  `json:"category"`
	CurrentSpending   float64 `json:"current_spending"`
	RecommendedBudget float64 `json:"recommended_budget"`
	PotentialSavings  float64 `json:"potential_savings"`
	Justification     string  `json:"justification"`
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{}
}

// GenerateInsights analyzes spending patterns and generates actionable insights
func (a *AnalyticsService) GenerateInsights(analytics *SpendingAnalytics) []SpendingInsight {
	var insights []SpendingInsight

	// Check for high food spending
	if foodSpending, exists := analytics.SpendingByCategory["Food"]; exists {
		foodPercentage := (foodSpending / analytics.TotalSpending) * 100
		if foodPercentage > 30 {
			insights = append(insights, SpendingInsight{
				Type:     "warning",
				Category: "Food",
				Message:  fmt.Sprintf("Food spending is %.1f%% of total expenses (₹%.2f). Consider meal planning or cooking at home more often.", foodPercentage, foodSpending),
				Impact:   foodSpending * 0.2, // Potential 20% savings
				Severity: "medium",
			})
		}
	}

	// Check for subscription spending patterns
	subscriptionCategories := []string{"Entertainment", "Utilities"}
	var totalSubscriptions float64
	for _, cat := range subscriptionCategories {
		if amount, exists := analytics.SpendingByCategory[cat]; exists {
			totalSubscriptions += amount
		}
	}
	if totalSubscriptions > 0 {
		subPercentage := (totalSubscriptions / analytics.TotalSpending) * 100
		if subPercentage > 15 {
			insights = append(insights, SpendingInsight{
				Type:     "tip",
				Category: "Subscriptions",
				Message:  fmt.Sprintf("Subscription-like expenses are %.1f%% of spending (₹%.2f). Review and cancel unused subscriptions.", subPercentage, totalSubscriptions),
				Impact:   totalSubscriptions * 0.3,
				Severity: "low",
			})
		}
	}

	// Check for spending trend increases
	if len(analytics.MonthlyTrends) >= 3 {
		recentTrends := analytics.MonthlyTrends[len(analytics.MonthlyTrends)-3:]
		increasingMonths := 0
		for _, trend := range recentTrends {
			if trend.Trend == "increase" {
				increasingMonths++
			}
		}
		if increasingMonths >= 2 {
			lastMonth := recentTrends[len(recentTrends)-1]
			insights = append(insights, SpendingInsight{
				Type:     "warning",
				Category: "General",
				Message:  fmt.Sprintf("Spending has been increasing for %d months. Last month: ₹%.2f", increasingMonths, lastMonth.Amount),
				Impact:   lastMonth.Amount * 0.1,
				Severity: "high",
			})
		}
	}

	// Check for high transportation costs
	if transportSpending, exists := analytics.SpendingByCategory["Transportation"]; exists {
		transportPercentage := (transportSpending / analytics.TotalSpending) * 100
		if transportPercentage > 20 {
			insights = append(insights, SpendingInsight{
				Type:     "tip",
				Category: "Transportation",
				Message:  fmt.Sprintf("Transportation is %.1f%% of expenses (₹%.2f). Consider carpooling, public transport, or route optimization.", transportPercentage, transportSpending),
				Impact:   transportSpending * 0.25,
				Severity: "medium",
			})
		}
	}

	// Check for impulse shopping patterns (high frequency, low amounts)
	if shoppingSpending, exists := analytics.SpendingByCategory["Shopping"]; exists {
		if insight, exists := analytics.CategoryInsights["Shopping"]; exists {
			if insight.TransactionCount > 10 && insight.AverageAmount < 500 {
				insights = append(insights, SpendingInsight{
					Type:     "tip",
					Category: "Shopping",
					Message:  fmt.Sprintf("High frequency shopping detected (%d transactions, avg ₹%.2f). Consider making shopping lists and batching purchases.", insight.TransactionCount, insight.AverageAmount),
					Impact:   shoppingSpending * 0.15,
					Severity: "low",
				})
			}
		}
	}

	return insights
}

// GenerateBudgetRecommendations creates personalized budget recommendations
func (a *AnalyticsService) GenerateBudgetRecommendations(analytics *SpendingAnalytics) []BudgetRecommendation {
	var recommendations []BudgetRecommendation

	// Calculate average monthly spending
	var avgMonthlySpending float64
	if len(analytics.MonthlyTrends) > 0 {
		var total float64
		for _, trend := range analytics.MonthlyTrends {
			total += trend.Amount
		}
		avgMonthlySpending = total / float64(len(analytics.MonthlyTrends))
	} else {
		avgMonthlySpending = analytics.TotalSpending
	}

	// Budget recommendations for each category
	for category, currentSpending := range analytics.SpendingByCategory {
		var recommendedBudget float64
		var justification string

		currentPercentage := (currentSpending / analytics.TotalSpending) * 100

		switch category {
		case "Food":
			if currentPercentage > 25 {
				recommendedBudget = avgMonthlySpending * 0.25
				justification = "Food should ideally be 20-25% of total spending. Consider meal planning and home cooking."
			} else {
				recommendedBudget = currentSpending * 1.05
				justification = "Your food spending is within healthy limits. Small buffer for inflation."
			}

		case "Transportation":
			if currentPercentage > 15 {
				recommendedBudget = avgMonthlySpending * 0.15
				justification = "Transportation should be around 10-15% of spending. Look for cost-effective alternatives."
			} else {
				recommendedBudget = currentSpending * 1.1
				justification = "Transportation spending is reasonable. Small increase for fuel price changes."
			}

		case "Entertainment":
			if currentPercentage > 10 {
				recommendedBudget = avgMonthlySpending * 0.08
				justification = "Entertainment should be 5-10% of spending. Consider free or low-cost activities."
			} else {
				recommendedBudget = currentSpending
				justification = "Entertainment spending is balanced. Maintain current levels."
			}

		case "Shopping":
			if currentPercentage > 20 {
				recommendedBudget = avgMonthlySpending * 0.15
				justification = "Shopping should be 10-15% of spending. Focus on needs vs wants."
			} else {
				recommendedBudget = currentSpending * 0.95
				justification = "Shopping spending is acceptable. Slight reduction for better savings."
			}

		case "Healthcare":
			recommendedBudget = math.Max(currentSpending, avgMonthlySpending*0.05)
			justification = "Healthcare is essential. Budget at least 5% for medical expenses."

		case "Utilities":
			recommendedBudget = currentSpending * 1.02
			justification = "Utilities are fixed costs. Small buffer for rate increases."

		default:
			recommendedBudget = currentSpending * 0.9
			justification = "General recommendation to optimize spending in this category."
		}

		potentialSavings := math.Max(0, currentSpending-recommendedBudget)

		recommendations = append(recommendations, BudgetRecommendation{
			Category:          category,
			CurrentSpending:   currentSpending,
			RecommendedBudget: recommendedBudget,
			PotentialSavings:  potentialSavings,
			Justification:     justification,
		})
	}

	// Sort by potential savings (highest first)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].PotentialSavings > recommendations[j].PotentialSavings
	})

	return recommendations
}

// CalculateSpendingScore gives an overall financial health score (0-100)
func (a *AnalyticsService) CalculateSpendingScore(analytics *SpendingAnalytics) (int, string) {
	score := 100.0
	var factors []string

	// Check essential vs non-essential spending balance
	essential := []string{"Food", "Healthcare", "Utilities", "Transportation"}
	var essentialSpending, nonEssentialSpending float64

	for category, amount := range analytics.SpendingByCategory {
		isEssential := false
		for _, ess := range essential {
			if category == ess {
				isEssential = true
				break
			}
		}
		if isEssential {
			essentialSpending += amount
		} else {
			nonEssentialSpending += amount
		}
	}

	essentialRatio := essentialSpending / analytics.TotalSpending
	if essentialRatio < 0.5 {
		score -= 20
		factors = append(factors, "Essential spending is too low")
	} else if essentialRatio > 0.8 {
		score -= 15
		factors = append(factors, "Too much spending on essentials, no room for savings")
	}

	// Check for spending consistency
	if len(analytics.MonthlyTrends) >= 3 {
		var variance float64
		mean := analytics.TotalSpending / float64(len(analytics.MonthlyTrends))

		for _, trend := range analytics.MonthlyTrends {
			variance += math.Pow(trend.Amount-mean, 2)
		}
		variance /= float64(len(analytics.MonthlyTrends))
		stdDev := math.Sqrt(variance)

		// If standard deviation is more than 30% of mean, deduct points
		if stdDev > mean*0.3 {
			score -= 15
			factors = append(factors, "Inconsistent spending patterns")
		}
	}

	// Check high-risk categories
	if foodPercentage := (analytics.SpendingByCategory["Food"] / analytics.TotalSpending) * 100; foodPercentage > 30 {
		score -= 10
		factors = append(factors, "High food spending")
	}

	if shoppingPercentage := (analytics.SpendingByCategory["Shopping"] / analytics.TotalSpending) * 100; shoppingPercentage > 25 {
		score -= 15
		factors = append(factors, "Excessive shopping")
	}

	// Check for positive trends
	if len(analytics.MonthlyTrends) >= 2 {
		lastTrend := analytics.MonthlyTrends[len(analytics.MonthlyTrends)-1]
		if lastTrend.Trend == "decrease" {
			score += 5
			factors = append(factors, "Spending decreased last month")
		}
	}

	// Cap score between 0 and 100
	finalScore := int(math.Max(0, math.Min(100, score)))

	var explanation string
	if finalScore >= 80 {
		explanation = "Excellent financial discipline! Your spending patterns are well-balanced."
	} else if finalScore >= 60 {
		explanation = "Good spending habits with room for improvement in some areas."
	} else if finalScore >= 40 {
		explanation = "Moderate spending control. Consider reviewing high-expense categories."
	} else {
		explanation = "Spending patterns need attention. Focus on budgeting and expense reduction."
	}

	if len(factors) > 0 {
		explanation += " Key factors: " + fmt.Sprintf("%v", factors)
	}

	return finalScore, explanation
}

// PredictNextMonthSpending estimates next month's spending based on trends
func (a *AnalyticsService) PredictNextMonthSpending(analytics *SpendingAnalytics) map[string]float64 {
	predictions := make(map[string]float64)

	if len(analytics.MonthlyTrends) < 2 {
		// Not enough data, return current averages
		for category, amount := range analytics.SpendingByCategory {
			predictions[category] = amount / float64(len(analytics.MonthlyTrends))
		}
		return predictions
	}

	// Simple trend-based prediction
	recentMonths := 3
	if len(analytics.MonthlyTrends) < recentMonths {
		recentMonths = len(analytics.MonthlyTrends)
	}

	// Calculate trend for total spending
	recentTrends := analytics.MonthlyTrends[len(analytics.MonthlyTrends)-recentMonths:]
	var totalRecent, avgGrowth float64

	for i, trend := range recentTrends {
		totalRecent += trend.Amount
		if i > 0 {
			growth := (trend.Amount - recentTrends[i-1].Amount) / recentTrends[i-1].Amount
			avgGrowth += growth
		}
	}

	if recentMonths > 1 {
		avgGrowth /= float64(recentMonths - 1)
	}

	avgMonthly := totalRecent / float64(recentMonths)
	predictedTotal := avgMonthly * (1 + avgGrowth)

	// Distribute predicted total across categories based on historical percentages
	for category, amount := range analytics.SpendingByCategory {
		categoryPercentage := amount / analytics.TotalSpending
		predictions[category] = predictedTotal * categoryPercentage
	}

	return predictions
}
