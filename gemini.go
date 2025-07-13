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

// ClassifyQuestion uses AI to intelligently determine what type of question the user is asking
func (g *GeminiClient) ClassifyQuestion(question string) QuestionType {
	ctx := context.Background()
	model := g.Client.GenerativeModel(g.Model)

	prompt := fmt.Sprintf(`You are an expert financial AI assistant. Classify this user question into one of these categories:

Question: "%s"

Categories:
1. SpendingAmount - Questions about how much was spent, totals, specific amounts
2. SpendingAdvice - Requests for advice, suggestions, recommendations for improvement
3. SpendingTrends - Questions about trends, patterns, comparisons over time
4. CategoryAnalysis - Questions about specific categories like food, transport, shopping
5. BudgetSuggestions - Questions about budgets, saving money, reducing expenses
6. General - Any other financial questions

Instructions:
- Analyze the intent and context of the question
- Consider the user's underlying need (amounts, advice, trends, categories, budgets)
- Return ONLY the category name (e.g., "SpendingAmount", "SpendingAdvice")
- Do not include any explanation or additional text

Examples:
- "How much did I spend on food?" → "CategoryAnalysis"
- "What's my total spending this month?" → "SpendingAmount"
- "Can you suggest ways to save money?" → "SpendingAdvice"
- "Are my expenses increasing?" → "SpendingTrends"
- "Help me create a budget" → "BudgetSuggestions"
- "What's my financial situation?" → "General"

Classification:`, question)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		fmt.Printf("⚠️ AI question classification failed, using fallback: %v\n", err)
		return g.classifyQuestionFallback(question)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		fmt.Printf("⚠️ No AI response for question classification, using fallback\n")
		return g.classifyQuestionFallback(question)
	}

	// Extract the classification from the response
	classification := strings.TrimSpace(string(resp.Candidates[0].Content.Parts[0].(genai.Text)))

	// Map AI response to QuestionType
	switch classification {
	case "SpendingAmount":
		fmt.Printf("🤖 AI classified question as: SpendingAmount\n")
		return SpendingAmount
	case "SpendingAdvice":
		fmt.Printf("🤖 AI classified question as: SpendingAdvice\n")
		return SpendingAdvice
	case "SpendingTrends":
		fmt.Printf("🤖 AI classified question as: SpendingTrends\n")
		return SpendingTrends
	case "CategoryAnalysis":
		fmt.Printf("🤖 AI classified question as: CategoryAnalysis\n")
		return CategoryAnalysis
	case "BudgetSuggestions":
		fmt.Printf("🤖 AI classified question as: BudgetSuggestions\n")
		return BudgetSuggestions
	case "General":
		fmt.Printf("🤖 AI classified question as: General\n")
		return General
	default:
		fmt.Printf("⚠️ Unknown AI classification '%s', using fallback\n", classification)
		return g.classifyQuestionFallback(question)
	}
}

// ClassifyQuestionWithContext uses AI with spending context for even smarter question classification
func (g *GeminiClient) ClassifyQuestionWithContext(question string, analytics *SpendingAnalytics) QuestionType {
	ctx := context.Background()
	model := g.Client.GenerativeModel(g.Model)

	// Build context information
	contextInfo := ""
	if analytics != nil {
		contextInfo = fmt.Sprintf(`

User's Spending Context:
- Total Spending: ₹%.2f
- Transaction Count: %d
- Average Transaction: ₹%.2f
- Top Categories: %s
- Recent Trends: %s

`, analytics.TotalSpending, analytics.TransactionCount, analytics.AverageTransaction,
			g.getTopCategories(analytics), g.getRecentTrends(analytics))
	}

	prompt := fmt.Sprintf(`You are an expert financial AI assistant. Classify this user question into one of these categories based on the question intent and the user's spending context:

Question: "%s"
%s
Categories:
1. SpendingAmount - Questions about how much was spent, totals, specific amounts
2. SpendingAdvice - Requests for advice, suggestions, recommendations for improvement
3. SpendingTrends - Questions about trends, patterns, comparisons over time
4. CategoryAnalysis - Questions about specific categories like food, transport, shopping
5. BudgetSuggestions - Questions about budgets, saving money, reducing expenses
6. General - Any other financial questions

Context Considerations:
- If user has high spending in a category, questions about that category are likely CategoryAnalysis
- If user has irregular spending patterns, trend questions are likely SpendingTrends
- If user asks vague questions but has concerning spending, they likely want SpendingAdvice
- Questions about amounts with specific context clues should be SpendingAmount

Instructions:
- Analyze the intent and context of the question
- Consider the user's spending patterns and context
- Return ONLY the category name (e.g., "SpendingAmount", "SpendingAdvice")
- Do not include any explanation or additional text

Classification:`, question, contextInfo)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		fmt.Printf("⚠️ AI context-aware classification failed, using standard AI classification: %v\n", err)
		return g.ClassifyQuestion(question)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		fmt.Printf("⚠️ No AI response for context-aware classification, using standard AI classification\n")
		return g.ClassifyQuestion(question)
	}

	// Extract the classification from the response
	classification := strings.TrimSpace(string(resp.Candidates[0].Content.Parts[0].(genai.Text)))

	// Map AI response to QuestionType
	switch classification {
	case "SpendingAmount":
		fmt.Printf("🧠 AI (context-aware) classified question as: SpendingAmount\n")
		return SpendingAmount
	case "SpendingAdvice":
		fmt.Printf("🧠 AI (context-aware) classified question as: SpendingAdvice\n")
		return SpendingAdvice
	case "SpendingTrends":
		fmt.Printf("🧠 AI (context-aware) classified question as: SpendingTrends\n")
		return SpendingTrends
	case "CategoryAnalysis":
		fmt.Printf("🧠 AI (context-aware) classified question as: CategoryAnalysis\n")
		return CategoryAnalysis
	case "BudgetSuggestions":
		fmt.Printf("🧠 AI (context-aware) classified question as: BudgetSuggestions\n")
		return BudgetSuggestions
	case "General":
		fmt.Printf("🧠 AI (context-aware) classified question as: General\n")
		return General
	default:
		fmt.Printf("⚠️ Unknown AI classification '%s', using standard AI classification\n", classification)
		return g.ClassifyQuestion(question)
	}
}

// Helper method to get top spending categories for context
func (g *GeminiClient) getTopCategories(analytics *SpendingAnalytics) string {
	if analytics == nil || len(analytics.SpendingByCategory) == 0 {
		return "No data"
	}

	type categoryAmount struct {
		category string
		amount   float64
	}

	var categories []categoryAmount
	for category, amount := range analytics.SpendingByCategory {
		categories = append(categories, categoryAmount{category, amount})
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].amount > categories[j].amount
	})

	topCount := 3
	if len(categories) < topCount {
		topCount = len(categories)
	}

	var result []string
	for i := 0; i < topCount; i++ {
		result = append(result, fmt.Sprintf("%s (₹%.0f)", categories[i].category, categories[i].amount))
	}

	return strings.Join(result, ", ")
}

// Helper method to get recent spending trends for context
func (g *GeminiClient) getRecentTrends(analytics *SpendingAnalytics) string {
	if analytics == nil || len(analytics.MonthlyTrends) < 2 {
		return "Insufficient data"
	}

	lastMonth := analytics.MonthlyTrends[len(analytics.MonthlyTrends)-1]
	return fmt.Sprintf("Last month: ₹%.0f (%s)", lastMonth.Amount, lastMonth.Trend)
}

// classifyQuestionFallback provides a fallback classification using keyword matching
func (g *GeminiClient) classifyQuestionFallback(question string) QuestionType {
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

	// Classify the question type using context-aware AI classification
	questionType := g.ClassifyQuestionWithContext(question, analytics)

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

// ClassifyVendor uses Gemini AI to classify a vendor into predefined categories
func (g *GeminiClient) ClassifyVendor(vendor string) (string, error) {
	ctx := context.Background()
	model := g.Client.GenerativeModel(g.Model)

	prompt := fmt.Sprintf(`Classify this vendor into one of these categories:
["Food", "Shopping", "Travel", "Entertainment", "Bills", "Healthcare", "Other"]

Vendor: "%s"

Instructions:
- Return ONLY the category name (e.g., "Food", "Shopping", etc.)
- Do not include any explanation or additional text
- Use "Other" if the vendor doesn't clearly fit any category
- Consider common vendor patterns and business types

Category:`, vendor)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate vendor classification: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	// Extract the category from the response
	category := strings.TrimSpace(string(resp.Candidates[0].Content.Parts[0].(genai.Text)))

	// Validate the category is one of our allowed categories
	validCategories := map[string]bool{
		"Food":          true,
		"Shopping":      true,
		"Travel":        true,
		"Entertainment": true,
		"Bills":         true,
		"Healthcare":    true,
		"Other":         true,
	}

	if !validCategories[category] {
		// If AI returns invalid category, default to "Other"
		category = "Other"
	}

	fmt.Printf("🤖 AI classified vendor '%s' as '%s'\n", vendor, category)
	return category, nil
}
