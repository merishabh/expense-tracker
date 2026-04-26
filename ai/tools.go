package ai

import "github.com/anthropics/anthropic-sdk-go"

// ToolExecutor is called by the agentic loop when Claude requests a tool.
type ToolExecutor func(name string, input map[string]any) (string, error)

// toolSchemas are the tool definitions sent to Claude on every request.
var toolSchemas = func() []anthropic.ToolUnionParam {
	categorySpend := anthropic.ToolParam{
		Name:        "get_category_spend",
		Description: anthropic.String("Get total amount spent in a category between two dates. Use this to answer questions about how much was spent on food, travel, bills, etc. in any time period."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"category": map[string]string{
					"type":        "string",
					"description": "Category name, e.g. Food, Travel, Bills, Grocery, Shopping, Entertainment, Healthcare, SIP, Subscription",
				},
				"from_date": map[string]string{
					"type":        "string",
					"description": "Start date in YYYY-MM-DD format (inclusive)",
				},
				"to_date": map[string]string{
					"type":        "string",
					"description": "End date in YYYY-MM-DD format (inclusive)",
				},
			},
			Required: []string{"category", "from_date", "to_date"},
		},
	}

	monthlySum := anthropic.ToolParam{
		Name:        "get_monthly_summary",
		Description: anthropic.String("Get a full spending breakdown for a specific month: total spend, per-category totals, and transaction count. Use this for month-level overviews or comparisons."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"from_date": map[string]string{
					"type":        "string",
					"description": "First day of the month in YYYY-MM-DD format, e.g. 2026-04-01",
				},
				"to_date": map[string]string{
					"type":        "string",
					"description": "Last day of the month in YYYY-MM-DD format, e.g. 2026-04-30",
				},
			},
			Required: []string{"from_date", "to_date"},
		},
	}

	topMerchants := anthropic.ToolParam{
		Name:        "get_top_merchants",
		Description: anthropic.String("Get the top merchants/vendors ranked by total spend in a date range. Use this to answer questions about where most money is being spent."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"from_date": map[string]string{
					"type":        "string",
					"description": "Start date in YYYY-MM-DD format (inclusive)",
				},
				"to_date": map[string]string{
					"type":        "string",
					"description": "End date in YYYY-MM-DD format (inclusive)",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Number of top merchants to return (default 10)",
				},
			},
			Required: []string{"from_date", "to_date"},
		},
	}

	getTransactions := anthropic.ToolParam{
		Name:        "get_transactions",
		Description: anthropic.String("Fetch individual transactions with full detail (datetime, vendor, amount, category). Use this when the question requires analysis that aggregated tools can't answer — e.g. time of day, day of week, weekend vs weekday, streaks, or any pattern over raw rows."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"category": map[string]string{
					"type":        "string",
					"description": "Filter by category, e.g. Food, Travel, Grocery. Leave empty for all categories.",
				},
				"from_date": map[string]string{
					"type":        "string",
					"description": "Start date in YYYY-MM-DD format (inclusive)",
				},
				"to_date": map[string]string{
					"type":        "string",
					"description": "End date in YYYY-MM-DD format (inclusive)",
				},
			},
			Required: []string{"from_date", "to_date"},
		},
	}

	saveMemory := anthropic.ToolParam{
		Name:        "save_memory",
		Description: anthropic.String("Persist something worth remembering about the user across conversations: a goal, a correction, a life-context fact, a spending pattern, or an analytical preference. Call this proactively when the user reveals information that should influence future responses."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"type": map[string]string{
					"type":        "string",
					"description": "Memory category: life_context | goal | correction | pattern | preference",
				},
				"content": map[string]string{
					"type":        "string",
					"description": "The memory to save, written as a short fact in third person, e.g. 'User wants to cut food spend to ₹5000/month'",
				},
			},
			Required: []string{"type", "content"},
		},
	}

	return []anthropic.ToolUnionParam{
		{OfTool: &categorySpend},
		{OfTool: &monthlySum},
		{OfTool: &topMerchants},
		{OfTool: &getTransactions},
		{OfTool: &saveMemory},
	}
}()
