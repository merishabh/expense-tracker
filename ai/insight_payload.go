package ai

// CategoryInsightPayload represents facts about category spending for Gemini explanation
type CategoryInsightPayload struct {
	Category       string  `json:"category"`
	Period         string  `json:"period"`
	TotalSpent     float64 `json:"total_spent"`
	AverageSpent   float64 `json:"average_spent"`
	Budget         float64 `json:"budget"`
	BudgetExceeded bool    `json:"budget_exceeded"`
	DeltaPercent   float64 `json:"delta_percent"`
	UserQuestion   string  `json:"user_question"`
}

// TotalSpendPayload represents facts about total spending for Gemini explanation
type TotalSpendPayload struct {
	Period       string  `json:"period"`
	TotalSpent   float64 `json:"total_spent"`
	Average      float64 `json:"average"`
	UserQuestion string  `json:"user_question"`
}

// ComparisonPayload represents facts about period/category comparison for Gemini explanation
type ComparisonPayload struct {
	BasePeriod    string  `json:"base_period"`
	ComparePeriod string  `json:"compare_period"`
	BaseAmount    float64 `json:"base_amount"`
	CompareAmount float64 `json:"compare_amount"`
	DeltaPercent  float64 `json:"delta_percent"`
	UserQuestion  string  `json:"user_question"`
}

// TopMerchantsPayload represents facts about top merchants for Gemini explanation
type TopMerchantsPayload struct {
	Period       string             `json:"period"`
	Merchants    map[string]float64 `json:"merchants"`
	UserQuestion string             `json:"user_question"`
}

// TrendPayload represents facts about spending trends for Gemini explanation
type TrendPayload struct {
	Period       string             `json:"period"`
	TrendData    map[string]float64 `json:"trend_data"`
	UserQuestion string             `json:"user_question"`
}

// GeneralInsightPayload represents general facts for Gemini explanation
type GeneralInsightPayload struct {
	FactsSummary string `json:"facts_summary"`
	UserQuestion string `json:"user_question"`
}
