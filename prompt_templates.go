package main

const (
	// CategorySummaryTemplate is the prompt template for CATEGORY_SUMMARY intent
	CategorySummaryTemplate = `You are a personal finance assistant.

Facts:
- Category: {{Category}}
- Period: {{Period}}
- Total spent: ₹{{TotalSpent}}
- Average spend: ₹{{AverageSpent}}
- Budget: ₹{{Budget}}
- Budget exceeded: {{BudgetExceeded}}
- Change vs average: {{DeltaPercent}}%

User question (for context only):
"{{UserQuestion}}"

Task:
1. Explain the spending in simple language.
2. Identify 1 likely reason.
3. Suggest 1 practical improvement.

Rules:
- Do NOT calculate numbers.
- Do NOT invent facts.
- Use only the provided data.`

	// TotalSpendTemplate is the prompt template for TOTAL_SPEND intent
	TotalSpendTemplate = `You are a personal finance assistant.

Facts:
- Period: {{Period}}
- Total spent: ₹{{TotalSpent}}
- Average spend: ₹{{Average}}

User question (for context only):
"{{UserQuestion}}"

Task:
Explain what this spending means and whether it looks normal.

Rules:
- Do NOT calculate numbers.
- Do NOT invent facts.
- Use only the provided data.`

	// ComparisonTemplate is the prompt template for PERIOD_COMPARISON and CATEGORY_COMPARISON intents
	ComparisonTemplate = `You are a personal finance assistant.

Facts:
- Base period: {{BasePeriod}}
- Compare period: {{ComparePeriod}}
- Base amount: ₹{{BaseAmount}}
- Compare amount: ₹{{CompareAmount}}
- Change: {{DeltaPercent}}%

User question (for context only):
"{{UserQuestion}}"

Task:
1. Explain the comparison in simple language.
2. Identify 1 likely reason for the change.
3. Suggest 1 practical action.

Rules:
- Do NOT calculate numbers.
- Do NOT invent facts.
- Use only the provided data.`

	// TopMerchantsTemplate is the prompt template for TOP_MERCHANTS intent
	TopMerchantsTemplate = `You are a personal finance assistant.

Facts:
- Period: {{Period}}
- Top merchants and spending:
{{MerchantsList}}

User question (for context only):
"{{UserQuestion}}"

Task:
1. Explain the spending patterns at these merchants.
2. Identify 1 insight about merchant preferences.
3. Suggest 1 practical optimization.

Rules:
- Do NOT calculate numbers.
- Do NOT invent facts.
- Use only the provided data.`

	// TrendTemplate is the prompt template for DAILY_TREND and MONTHLY_TREND intents
	TrendTemplate = `You are a personal finance assistant.

Facts:
- Period: {{Period}}
- Trend data:
{{TrendDataList}}

User question (for context only):
"{{UserQuestion}}"

Task:
1. Explain the spending trend in simple language.
2. Identify 1 pattern or insight.
3. Suggest 1 practical recommendation.

Rules:
- Do NOT calculate numbers.
- Do NOT invent facts.
- Use only the provided data.`

	// GeneralInsightTemplate is the prompt template for GENERAL_INSIGHT and other intents
	GeneralInsightTemplate = `You are a personal finance coach.

Facts:
{{FactsSummary}}

User question:
"{{UserQuestion}}"

Task:
- Identify patterns
- Explain possible causes
- Suggest one behavioral change

Rules:
- Do NOT calculate numbers.
- Do NOT invent facts.
- Use only the provided data.`
)
