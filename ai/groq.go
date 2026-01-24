package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/yourusername/expense-tracker/models"
)

const groqURL = "https://api.groq.com/openai/v1/chat/completions"

// GroqClient handles intent classification using Groq API
type GroqClient struct {
	APIKey string
	Model  string
}

// NewGroqClient creates a new Groq client for intent classification
func NewGroqClient(apiKey string) *GroqClient {
	return &GroqClient{
		APIKey: apiKey,
		Model:  "llama-3.1-8b-instant", // Groq's fast model, can be changed to llama-3.3-70b-versatile or mixtral-8x7b-32768
	}
}

// GroqRequest represents the OpenAI-compatible request structure
type GroqRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GroqResponse represents the OpenAI-compatible response structure
type GroqResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice represents a choice in the response
type Choice struct {
	Message Message `json:"message"`
}

// ClassifyIntent uses Groq API to convert a free-text user question into a validated intent JSON.
// This function ONLY classifies intent and extracts entities - it does NOT query the database or compute data.
func (g *GroqClient) ClassifyIntent(question string) (*ExpenseIntent, error) {
	log.Printf("[ClassifyIntent] Starting intent classification for question: %q", question)

	prompt := fmt.Sprintf(`You are an expense tracker intent classifier. Analyze the user's question and extract structured intent information.

User Question: "%s"

Valid Intent Types (use exactly these strings):
- "TOTAL_SPEND" - Questions about total spending, overall amounts, summary
- "CATEGORY_SUMMARY" - Questions about spending in a specific category
- "CATEGORY_COMPARISON" - Questions comparing spending across categories
- "PERIOD_COMPARISON" - Questions comparing spending across time periods
- "TOP_MERCHANTS" - Questions about top vendors/merchants, where money is spent
- "DAILY_TREND" - Questions about daily spending patterns, daily trends
- "MONTHLY_TREND" - Questions about monthly spending patterns, monthly trends
- "ANOMALY_EXPLANATION" - Questions about unusual spending, anomalies, outliers
- "BUDGET_STATUS" - Questions about budget, budget remaining, budget limits
- "GENERAL_INSIGHT" - Any other financial questions, general insights

Valid Categories (if mentioned): Food, Shopping, Travel, Entertainment, Bills, Healthcare, Amazon, Other

Valid Period Values (if mentioned, use exactly these strings):
- "TODAY" - Today, today's spending, spent today, today only
- "YESTERDAY" - Yesterday, yesterday's spending, spent yesterday
- "THIS_WEEK" - This week, current week, week so far
- "LAST_WEEK" - Last week, previous week
- "THIS_MONTH" - This month, current month, month so far
- "LAST_MONTH" - Last month, previous month

IMPORTANT: Pay close attention to temporal words in the question:
- "today", "today's", "spent today" → use "TODAY"
- "yesterday", "yesterday's" → use "YESTERDAY"
- "this week" → use "THIS_WEEK"
- "this month" → use "THIS_MONTH"
- "last month" → use "LAST_MONTH"
- If no time period is mentioned, DO NOT include a period field

Instructions:
1. Determine the primary intent type based on the question (must be one of the valid Intent Types above)
2. CAREFULLY extract any mentioned category, period (using the valid Period values), vendor, or amount
   - For period: Look for words like "today", "yesterday", "this week", "this month", etc.
   - If the question says "today" or "today's", you MUST set period to "TODAY"
   - If the question says "this month" or "month", set period to "THIS_MONTH"
   - If no time period is mentioned, omit the period field entirely
3. Return a valid JSON object with the following structure:
{
  "intent_type": "<one of the valid intent types>",
  "category": "<category if mentioned, otherwise omit>",
  "period": "<period enum value if mentioned, otherwise omit>",
  "vendor": "<vendor name if mentioned, otherwise omit>",
  "amount": <numeric value if mentioned, otherwise omit>,
  "confidence": <0.0-1.0 confidence score>
}

Requirements:
- Return ONLY valid JSON, no explanation or additional text
- Omit optional fields entirely if not mentioned (don't use null)
- Set confidence between 0.0 and 1.0 based on how clear the intent is
- Category must be one of: Food, Shopping, Travel, Entertainment, Bills, Healthcare, Amazon, Other
- Intent type must be exactly one of the valid types listed above (use the exact string)
- Period must be exactly one of: TODAY, YESTERDAY, THIS_WEEK, LAST_WEEK, THIS_MONTH, LAST_MONTH (if mentioned)

Examples:
Question: "How much did I spend on food this month?"
Response: {"intent_type": "CATEGORY_SUMMARY", "category": "Food", "period": "THIS_MONTH", "confidence": 0.95}

Question: "How much did I spend today?"
Response: {"intent_type": "TOTAL_SPEND", "period": "TODAY", "confidence": 0.95}

Question: "What did I spend today?"
Response: {"intent_type": "TOTAL_SPEND", "period": "TODAY", "confidence": 0.95}

Question: "Show me my total spending"
Response: {"intent_type": "TOTAL_SPEND", "confidence": 0.9}

Question: "How much did I spend yesterday?"
Response: {"intent_type": "TOTAL_SPEND", "period": "YESTERDAY", "confidence": 0.95}

Question: "How much did I spend in kims?"
Response: {"intent_type": "CATEGORY_SUMMARY", "vendor": "kims", "confidence": 0.9}

Question: "What did I spend at zomato this month?"
Response: {"intent_type": "CATEGORY_SUMMARY", "vendor": "zomato", "period": "THIS_MONTH", "confidence": 0.95}

Question: "Compare my spending last month to this month"
Response: {"intent_type": "PERIOD_COMPARISON", "parameters": {"period1": "LAST_MONTH", "period2": "THIS_MONTH"}, "confidence": 0.9}

Question: "What are my top merchants?"
Response: {"intent_type": "TOP_MERCHANTS", "confidence": 0.95}

Question: "Show me my daily spending trend"
Response: {"intent_type": "DAILY_TREND", "confidence": 0.9}

Question: "How is my budget looking?"
Response: {"intent_type": "BUDGET_STATUS", "confidence": 0.85}

Now classify this question and return the JSON:`, question)

	// Create the request
	req := GroqRequest{
		Model: g.Model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.1, // Lower temperature for more consistent classification
		MaxTokens:   500, // Enough for JSON response
	}

	log.Printf("[ClassifyIntent] Sending request to Groq API with model: %s", g.Model)

	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		log.Printf("[ClassifyIntent] Error marshaling request: %v", err)
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", groqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.APIKey))

	// Send request
	client := &http.Client{}
	log.Printf("[ClassifyIntent] Sending HTTP POST request to Groq API")
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[ClassifyIntent] Error sending request to Groq: %v", err)
		return nil, fmt.Errorf("failed to send request to Groq: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("[ClassifyIntent] Received response with status code: %d", resp.StatusCode)

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ClassifyIntent] Error reading response body: %v", err)
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("[ClassifyIntent] Groq API returned error status %d: %s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("groq API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var groqResp GroqResponse
	if err := json.Unmarshal(respBody, &groqResp); err != nil {
		log.Printf("[ClassifyIntent] Error parsing Groq response JSON: %v", err)
		return nil, fmt.Errorf("failed to parse Groq response: %v", err)
	}

	log.Printf("[ClassifyIntent] Parsed Groq response, number of choices: %d", len(groqResp.Choices))

	if len(groqResp.Choices) == 0 || groqResp.Choices[0].Message.Content == "" {
		log.Printf("[ClassifyIntent] No response content from Groq")
		return nil, fmt.Errorf("no response from Groq")
	}

	// Extract the JSON response
	responseText := strings.TrimSpace(groqResp.Choices[0].Message.Content)
	log.Printf("[ClassifyIntent] Raw response from Groq (before cleaning): %q", responseText)

	// Clean up the response - remove markdown code blocks if present
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSpace(responseText)
	responseText = strings.TrimSuffix(responseText, "```")
	log.Printf("[ClassifyIntent] Cleaned response text: %q", responseText)

	// Parse the JSON response into a temporary struct to handle enum conversion
	type tempIntent struct {
		IntentType string            `json:"intent_type"`
		Category   string            `json:"category,omitempty"`
		Period     string            `json:"period,omitempty"`
		Vendor     string            `json:"vendor,omitempty"`
		Amount     *float64          `json:"amount,omitempty"`
		Parameters map[string]string `json:"parameters,omitempty"`
		Confidence float64           `json:"confidence"`
	}

	var temp tempIntent
	if err := json.Unmarshal([]byte(responseText), &temp); err != nil {
		log.Printf("[ClassifyIntent] Error parsing intent JSON: %v, response text: %q", err, responseText)
		return nil, fmt.Errorf("failed to parse intent JSON: %v (response: %s)", err, responseText)
	}

	log.Printf("[ClassifyIntent] Parsed intent JSON - IntentType: %s, Category: %s, Period: %s, Confidence: %.2f",
		temp.IntentType, temp.Category, temp.Period, temp.Confidence)

	// Convert intent type string to enum
	var intentType IntentType
	switch temp.IntentType {
	case "TOTAL_SPEND":
		intentType = TOTAL_SPEND
	case "CATEGORY_SUMMARY":
		intentType = CATEGORY_SUMMARY
	case "CATEGORY_COMPARISON":
		intentType = CATEGORY_COMPARISON
	case "PERIOD_COMPARISON":
		intentType = PERIOD_COMPARISON
	case "TOP_MERCHANTS":
		intentType = TOP_MERCHANTS
	case "DAILY_TREND":
		intentType = DAILY_TREND
	case "MONTHLY_TREND":
		intentType = MONTHLY_TREND
	case "ANOMALY_EXPLANATION":
		intentType = ANOMALY_EXPLANATION
	case "BUDGET_STATUS":
		intentType = BUDGET_STATUS
	case "GENERAL_INSIGHT":
		intentType = GENERAL_INSIGHT
	default:
		return nil, fmt.Errorf("invalid intent type: %s", temp.IntentType)
	}

	// Convert period string to enum if provided
	var period *Period
	if temp.Period != "" {
		var p Period
		switch temp.Period {
		case "TODAY":
			p = TODAY
		case "YESTERDAY":
			p = YESTERDAY
		case "THIS_WEEK":
			p = THIS_WEEK
		case "LAST_WEEK":
			p = LAST_WEEK
		case "THIS_MONTH":
			p = THIS_MONTH
		case "LAST_MONTH":
			p = LAST_MONTH
		default:
			return nil, fmt.Errorf("invalid period: %s", temp.Period)
		}
		period = &p
	}

	// Validate category if provided
	category := temp.Category
	if category != "" {
		validCategories := map[string]bool{
			"Food":          true,
			"Shopping":      true,
			"Travel":        true,
			"Entertainment": true,
			"Bills":         true,
			"Healthcare":    true,
			"Amazon":        true,
			"Other":         true,
		}
		if !validCategories[category] {
			// Invalid category, set to empty
			category = ""
		}
	}

	// Ensure confidence is between 0.0 and 1.0
	confidence := temp.Confidence
	if confidence < 0.0 {
		confidence = 0.0
	}
	if confidence > 1.0 {
		confidence = 1.0
	}

	intent := &ExpenseIntent{
		IntentType: intentType,
		Category:   category,
		Period:     period,
		Vendor:     temp.Vendor,
		Amount:     temp.Amount,
		Parameters: temp.Parameters,
		Confidence: confidence,
	}

	log.Printf("[ClassifyIntent] Successfully classified intent - Type: %s, Category: %s, Period: %v, Vendor: %s, Confidence: %.2f",
		intent.IntentType.String(), intent.Category, intent.Period, intent.Vendor, intent.Confidence)

	return intent, nil
}

// GenerateExplanation sends a prompt to Groq and returns the explanation text
func (g *GroqClient) GenerateExplanation(prompt string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	// Create the request with higher temperature for natural language generation
	req := GroqRequest{
		Model: g.Model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,  // Higher temperature for more natural explanations
		MaxTokens:   2000, // More tokens for explanations
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", groqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.APIKey))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Groq: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var groqResp GroqResponse
	if err := json.Unmarshal(respBody, &groqResp); err != nil {
		return "", fmt.Errorf("failed to parse Groq response: %v", err)
	}

	if len(groqResp.Choices) == 0 || groqResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no response from Groq")
	}

	explanation := strings.TrimSpace(groqResp.Choices[0].Message.Content)
	if explanation == "" {
		return "", fmt.Errorf("empty response from Groq")
	}

	return explanation, nil
}

// ClassifyVendor uses Groq AI to classify a vendor into predefined categories
func (g *GroqClient) ClassifyVendor(vendor string) (string, error) {
	prompt := fmt.Sprintf(`Classify this vendor into one of these categories:
["Food", "Shopping", "Travel", "Entertainment", "Bills", "Healthcare", "Other"]

Vendor: "%s"

Instructions:
- Return ONLY the category name (e.g., "Food", "Shopping", etc.)
- Do not include any explanation or additional text
- Use "Other" if the vendor doesn't clearly fit any category
- Consider common vendor patterns and business types

Category:`, vendor)

	// Create the request
	req := GroqRequest{
		Model: g.Model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.1, // Lower temperature for more consistent classification
		MaxTokens:   100, // Short response expected
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", groqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.APIKey))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Groq: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var groqResp GroqResponse
	if err := json.Unmarshal(respBody, &groqResp); err != nil {
		return "", fmt.Errorf("failed to parse Groq response: %v", err)
	}

	if len(groqResp.Choices) == 0 || groqResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no response from Groq")
	}

	// Extract the category from the response
	category := strings.TrimSpace(groqResp.Choices[0].Message.Content)

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

	fmt.Printf("🤖 Groq classified vendor '%s' as '%s'\n", vendor, category)
	return category, nil
}

// AnalyzeTransactions creates comprehensive analytics from raw transactions
// This is a pure data processing function (no AI involved)
func (g *GroqClient) AnalyzeTransactions(transactions []models.Transaction) *SpendingAnalytics {
	analytics := &SpendingAnalytics{
		SpendingByCategory: make(map[string]float64),
		SpendingByMonth:    make(map[string]float64),
		CategoryInsights:   make(map[string]CategoryInsight),
	}

	categoryTransactions := make(map[string][]models.Transaction)
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
