package main

import (
	context "context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	genai "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	Client *genai.Client
	Model  string
}

// SpendingAnalytics contains structured data for better AI analysis
type SpendingAnalytics struct {
	TotalSpending      float64                    `json:"total_spending"`
	SpendingByCategory map[string]float64         `json:"spending_by_category"`
	SpendingByMonth    map[string]float64         `json:"spending_by_month"`
	TopVendors         []VendorSpending           `json:"top_vendors"`
	TransactionCount   int                        `json:"transaction_count"`
	AverageTransaction float64                    `json:"average_transaction"`
	MonthlyTrends      []MonthlyTrend             `json:"monthly_trends"`
	CategoryInsights   map[string]CategoryInsight `json:"category_insights"`
	RecentTransactions []Transaction              `json:"recent_transactions"`
}

type VendorSpending struct {
	Vendor string  `json:"vendor"`
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
}

type MonthlyTrend struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
	Trend  string  `json:"trend"` // "increase", "decrease", "stable"
}

type CategoryInsight struct {
	TotalAmount      float64 `json:"total_amount"`
	TransactionCount int     `json:"transaction_count"`
	AverageAmount    float64 `json:"average_amount"`
	Percentage       float64 `json:"percentage"`
	TopVendor        string  `json:"top_vendor"`
}

// QuestionType represents different types of questions users might ask
type QuestionType int

const (
	SpendingAmount QuestionType = iota
	SpendingAdvice
	SpendingTrends
	CategoryAnalysis
	BudgetSuggestions
	General
)

func NewGeminiClient(apiKey string) *GeminiClient {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		panic(fmt.Sprintf("Failed to create Gemini client: %v", err))
	}
	return &GeminiClient{
		Client: client,
		Model:  "models/gemini-2.0-flash",
	}
}

// AnalyzeTransactions creates comprehensive analytics from raw transactions
func (g *GeminiClient) AnalyzeTransactions(transactions []Transaction) *SpendingAnalytics {
	analytics := &SpendingAnalytics{
		SpendingByCategory: make(map[string]float64),
		SpendingByMonth:    make(map[string]float64),
		CategoryInsights:   make(map[string]CategoryInsight),
	}

	categoryTransactions := make(map[string][]Transaction)
	vendorSpending := make(map[string]*VendorSpending)
	monthlyData := make(map[string]*MonthlyTrend)

	for _, tx := range transactions {
		// Total spending
		analytics.TotalSpending += tx.Amount
		analytics.TransactionCount++

		// Category analysis
		analytics.SpendingByCategory[tx.Category] += tx.Amount
		categoryTransactions[tx.Category] = append(categoryTransactions[tx.Category], tx)

		// Vendor analysis
		if vs, exists := vendorSpending[tx.Vendor]; exists {
			vs.Amount += tx.Amount
			vs.Count++
		} else {
			vendorSpending[tx.Vendor] = &VendorSpending{
				Vendor: tx.Vendor,
				Amount: tx.Amount,
				Count:  1,
			}
		}

		// Monthly analysis
		monthKey := tx.DateTime.Format("2006-01")
		analytics.SpendingByMonth[monthKey] += tx.Amount
		if mt, exists := monthlyData[monthKey]; exists {
			mt.Amount += tx.Amount
			mt.Count++
		} else {
			monthlyData[monthKey] = &MonthlyTrend{
				Month:  monthKey,
				Amount: tx.Amount,
				Count:  1,
			}
		}
	}

	// Calculate averages and insights
	if analytics.TransactionCount > 0 {
		analytics.AverageTransaction = analytics.TotalSpending / float64(analytics.TransactionCount)
	}

	// Top vendors
	for _, vs := range vendorSpending {
		analytics.TopVendors = append(analytics.TopVendors, *vs)
	}
	sort.Slice(analytics.TopVendors, func(i, j int) bool {
		return analytics.TopVendors[i].Amount > analytics.TopVendors[j].Amount
	})
	if len(analytics.TopVendors) > 10 {
		analytics.TopVendors = analytics.TopVendors[:10]
	}

	// Category insights
	for category, amount := range analytics.SpendingByCategory {
		txs := categoryTransactions[category]
		insight := CategoryInsight{
			TotalAmount:      amount,
			TransactionCount: len(txs),
			AverageAmount:    amount / float64(len(txs)),
			Percentage:       (amount / analytics.TotalSpending) * 100,
		}

		// Find top vendor in category
		categoryVendors := make(map[string]float64)
		for _, tx := range txs {
			categoryVendors[tx.Vendor] += tx.Amount
		}
		maxAmount := 0.0
		for vendor, vendorAmount := range categoryVendors {
			if vendorAmount > maxAmount {
				maxAmount = vendorAmount
				insight.TopVendor = vendor
			}
		}
		analytics.CategoryInsights[category] = insight
	}

	// Monthly trends
	for _, mt := range monthlyData {
		analytics.MonthlyTrends = append(analytics.MonthlyTrends, *mt)
	}
	sort.Slice(analytics.MonthlyTrends, func(i, j int) bool {
		return analytics.MonthlyTrends[i].Month < analytics.MonthlyTrends[j].Month
	})

	// Calculate trends
	for i := 1; i < len(analytics.MonthlyTrends); i++ {
		current := analytics.MonthlyTrends[i].Amount
		previous := analytics.MonthlyTrends[i-1].Amount
		if current > previous*1.1 {
			analytics.MonthlyTrends[i].Trend = "increase"
		} else if current < previous*0.9 {
			analytics.MonthlyTrends[i].Trend = "decrease"
		} else {
			analytics.MonthlyTrends[i].Trend = "stable"
		}
	}

	// Recent transactions (last 10)
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].DateTime.After(transactions[j].DateTime)
	})
	recentCount := 10
	if len(transactions) < recentCount {
		recentCount = len(transactions)
	}
	analytics.RecentTransactions = transactions[:recentCount]

	return analytics
}

// ClassifyQuestion determines what type of question the user is asking
func (g *GeminiClient) ClassifyQuestion(question string) QuestionType {
	question = strings.ToLower(question)

	if strings.Contains(question, "how much") || strings.Contains(question, "total") || strings.Contains(question, "spent") {
		return SpendingAmount
	}
	if strings.Contains(question, "advice") || strings.Contains(question, "suggest") || strings.Contains(question, "recommend") {
		return SpendingAdvice
	}
	if strings.Contains(question, "trend") || strings.Contains(question, "pattern") || strings.Contains(question, "compare") {
		return SpendingTrends
	}
	if strings.Contains(question, "category") || strings.Contains(question, "food") || strings.Contains(question, "transport") {
		return CategoryAnalysis
	}
	if strings.Contains(question, "budget") || strings.Contains(question, "save") || strings.Contains(question, "reduce") {
		return BudgetSuggestions
	}
	return General
}

// BuildPrompt creates specialized prompts based on question type
func (g *GeminiClient) BuildPrompt(analytics *SpendingAnalytics, question string, questionType QuestionType) string {
	analyticsJSON, _ := json.MarshalIndent(analytics, "", "  ")

	// Generate additional insights
	analyticsService := NewAnalyticsService()
	insights := analyticsService.GenerateInsights(analytics)
	recommendations := analyticsService.GenerateBudgetRecommendations(analytics)
	score, scoreExplanation := analyticsService.CalculateSpendingScore(analytics)
	predictions := analyticsService.PredictNextMonthSpending(analytics)

	insightsJSON, _ := json.MarshalIndent(insights, "", "  ")
	recommendationsJSON, _ := json.MarshalIndent(recommendations, "", "  ")
	predictionsJSON, _ := json.MarshalIndent(predictions, "", "  ")

	baseContext := fmt.Sprintf(`You are a personal finance advisor analyzing spending data. Here is the comprehensive analysis:

SPENDING ANALYTICS:
%s

FINANCIAL HEALTH SCORE: %d/100
%s

INSIGHTS & WARNINGS:
%s

BUDGET RECOMMENDATIONS:
%s

NEXT MONTH PREDICTIONS:
%s

Current date: %s
Analysis period: Covers transactions from the data above

`, analyticsJSON, score, scoreExplanation, insightsJSON, recommendationsJSON, predictionsJSON, time.Now().Format("January 2, 2006"))

	switch questionType {
	case SpendingAmount:
		return baseContext + fmt.Sprintf(`
QUESTION TYPE: Spending Amount Query
USER QUESTION: %s

INSTRUCTIONS:
- Provide specific numerical answers with currency amounts
- Break down by categories if relevant
- Compare with other periods if data is available
- Use clear, concise language
- Include percentage breakdowns when helpful

Please answer the question with specific amounts and relevant breakdowns.`, question)

	case SpendingAdvice:
		return baseContext + fmt.Sprintf(`
QUESTION TYPE: Spending Advice Request
USER QUESTION: %s

INSTRUCTIONS:
- Analyze spending patterns and identify areas for improvement
- Provide 3-5 specific, actionable recommendations
- Focus on largest expense categories first
- Suggest realistic budget targets
- Consider spending trends and patterns
- Highlight any concerning spending spikes

Please provide personalized financial advice based on the spending patterns.`, question)

	case SpendingTrends:
		return baseContext + fmt.Sprintf(`
QUESTION TYPE: Spending Trends Analysis
USER QUESTION: %s

INSTRUCTIONS:
- Focus on monthly trends and patterns
- Identify increasing/decreasing spending areas
- Compare different time periods
- Highlight seasonal patterns if visible
- Provide insights on spending velocity and frequency

Please analyze the spending trends and patterns from the data.`, question)

	case CategoryAnalysis:
		return baseContext + fmt.Sprintf(`
QUESTION TYPE: Category Analysis
USER QUESTION: %s

INSTRUCTIONS:
- Deep dive into specific spending categories
- Compare category spending percentages
- Identify top vendors in each category
- Suggest optimizations for specific categories
- Provide category-specific insights

Please provide detailed analysis of the requested spending category.`, question)

	case BudgetSuggestions:
		return baseContext + fmt.Sprintf(`
QUESTION TYPE: Budget Suggestions
USER QUESTION: %s

INSTRUCTIONS:
- Create realistic budget recommendations based on current spending
- Suggest specific amounts for each category
- Identify areas where spending can be reduced
- Provide a structured monthly budget plan
- Consider essential vs. discretionary spending

Please create a practical budget plan based on the spending history.`, question)

	default:
		return baseContext + fmt.Sprintf(`
USER QUESTION: %s

Please provide a comprehensive financial analysis and answer based on the spending data provided.`, question)
	}
}

// AskGemini sends structured analytics and an intelligent prompt to Gemini API
func (g *GeminiClient) AskGemini(transactions []Transaction, question string) (string, error) {
	ctx := context.Background()

	// Analyze transactions to create structured data
	analytics := g.AnalyzeTransactions(transactions)

	// Classify the question type
	questionType := g.ClassifyQuestion(question)

	// Build an intelligent prompt
	prompt := g.BuildPrompt(analytics, question, questionType)

	model := g.Client.GenerativeModel(g.Model)

	// Configure model for better financial analysis
	model.SetTemperature(0.1) // Lower temperature for more factual responses
	model.SetTopK(1)
	model.SetTopP(0.8)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 {
		if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return string(text), nil
		}
	}

	return "", fmt.Errorf("no response from Gemini")
}
