package handlers

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/yourusername/expense-tracker/ai"
	"github.com/yourusername/expense-tracker/services"
)

// NewToolExecutor builds a ToolExecutor backed by the given ReportingService.
func NewToolExecutor(reporting *services.ReportingService) ai.ToolExecutor {
	return func(name string, input map[string]any) (string, error) {
		switch name {
		case "get_category_spend":
			return executeCategorySpend(reporting, input)
		case "get_monthly_summary":
			return executeMonthlySum(reporting, input)
		case "get_top_merchants":
			return executeTopMerchants(reporting, input)
		case "get_transactions":
			return executeGetTransactions(reporting, input)
		default:
			return "", fmt.Errorf("unknown tool: %s", name)
		}
	}
}

func executeCategorySpend(r *services.ReportingService, input map[string]any) (string, error) {
	category, _ := input["category"].(string)
	from, to, err := parseDateRange(input)
	if err != nil {
		return "", err
	}

	txs, err := r.ListTransactionsByDateRange(from, to, category, 1000)
	if err != nil {
		return "", err
	}

	var total float64
	count := 0
	for _, tx := range txs {
		if !tx.IsCredit() {
			total += tx.Amount
			count++
		}
	}

	return fmt.Sprintf("Category: %s | Period: %s to %s | Total spend: ₹%.2f | Transactions: %d",
		category, input["from_date"], input["to_date"], total, count), nil
}

func executeMonthlySum(r *services.ReportingService, input map[string]any) (string, error) {
	from, to, err := parseDateRange(input)
	if err != nil {
		return "", err
	}

	txs, err := r.ListTransactionsByDateRange(from, to, "", 2000)
	if err != nil {
		return "", err
	}

	var total float64
	categoryTotals := map[string]float64{}
	count := 0

	for _, tx := range txs {
		if tx.IsCredit() {
			continue
		}
		total += tx.Amount
		count++
		cat := tx.Category
		if cat == "" {
			cat = "Other"
		}
		categoryTotals[cat] += tx.Amount
	}

	type kv struct {
		k string
		v float64
	}
	var cats []kv
	for k, v := range categoryTotals {
		cats = append(cats, kv{k, v})
	}
	sort.Slice(cats, func(i, j int) bool { return cats[i].v > cats[j].v })

	var sb strings.Builder
	fmt.Fprintf(&sb, "Period: %s to %s | Total: ₹%.2f | Transactions: %d\nBy category:\n",
		input["from_date"], input["to_date"], total, count)
	for _, c := range cats {
		fmt.Fprintf(&sb, "  %s: ₹%.2f\n", c.k, c.v)
	}

	return sb.String(), nil
}

func executeTopMerchants(r *services.ReportingService, input map[string]any) (string, error) {
	from, to, err := parseDateRange(input)
	if err != nil {
		return "", err
	}

	limit := 10
	if v, ok := input["limit"]; ok {
		switch n := v.(type) {
		case float64:
			limit = int(n)
		case int:
			limit = n
		}
	}

	txs, err := r.ListTransactionsByDateRange(from, to, "", 2000)
	if err != nil {
		return "", err
	}

	totals := map[string]float64{}
	counts := map[string]int{}
	for _, tx := range txs {
		if tx.IsCredit() || tx.Vendor == "" {
			continue
		}
		totals[tx.Vendor] += tx.Amount
		counts[tx.Vendor]++
	}

	type kv struct {
		k string
		v float64
	}
	var vendors []kv
	for k, v := range totals {
		vendors = append(vendors, kv{k, v})
	}
	sort.Slice(vendors, func(i, j int) bool { return vendors[i].v > vendors[j].v })
	if len(vendors) > limit {
		vendors = vendors[:limit]
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Top merchants %s to %s:\n", input["from_date"], input["to_date"])
	for i, v := range vendors {
		fmt.Fprintf(&sb, "  %d. %s: ₹%.2f (%d txns)\n", i+1, v.k, v.v, counts[v.k])
	}

	return sb.String(), nil
}

func executeGetTransactions(r *services.ReportingService, input map[string]any) (string, error) {
	category, _ := input["category"].(string)
	from, to, err := parseDateRange(input)
	if err != nil {
		return "", err
	}

	txs, err := r.ListTransactionsByDateRange(from, to, category, 1000)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Transactions (%s to %s", input["from_date"], input["to_date"])
	if category != "" {
		fmt.Fprintf(&sb, ", category: %s", category)
	}
	fmt.Fprintf(&sb, "):\n")

	for _, tx := range txs {
		if tx.IsCredit() {
			continue
		}
		fmt.Fprintf(&sb, "  %s | %s | %s | ₹%.2f\n",
			tx.DateTime.Format("2006-01-02 15:04 Mon"),
			tx.Category,
			tx.Vendor,
			tx.Amount,
		)
	}
	return sb.String(), nil
}

func parseDateRange(input map[string]any) (from, to time.Time, err error) {
	fromStr, _ := input["from_date"].(string)
	toStr, _ := input["to_date"].(string)

	from, err = time.Parse("2006-01-02", fromStr)
	if err != nil {
		return from, to, fmt.Errorf("invalid from_date %q: %w", fromStr, err)
	}
	to, err = time.Parse("2006-01-02", toStr)
	if err != nil {
		return from, to, fmt.Errorf("invalid to_date %q: %w", toStr, err)
	}
	to = to.Add(24*time.Hour - time.Nanosecond)
	return from, to, nil
}
