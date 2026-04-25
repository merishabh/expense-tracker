package ai

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/yourusername/expense-tracker/models"
)

type ClaudeClient struct {
	client *anthropic.Client
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewClaudeClient(apiKey string) *ClaudeClient {
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &ClaudeClient{client: &c}
}

func (c *ClaudeClient) Chat(question string, history []ChatMessage, transactions []models.Transaction) (string, error) {
	systemPrompt := buildFinanceSystemPrompt(transactions)
	log.Printf("claude chat: question=%q history_len=%d system_prompt=%s", question, len(history), systemPrompt)

	messages := make([]anthropic.MessageParam, 0, len(history)+1)
	for _, h := range history {
		if h.Role == "user" {
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(h.Content)))
		} else {
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(h.Content)))
		}
	}
	messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(question)))

	msg, err := c.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: messages,
	})
	if err != nil {
		return "", fmt.Errorf("claude API error: %w", err)
	}

	if len(msg.Content) == 0 {
		return "", fmt.Errorf("empty response from claude")
	}
	return msg.Content[0].Text, nil
}

func buildFinanceSystemPrompt(transactions []models.Transaction) string {
	if len(transactions) == 0 {
		return "You are a personal finance assistant. No transaction data is available for the last 30 days."
	}

	var totalSpend float64
	categoryTotals := map[string]float64{}
	vendorTotals := map[string]float64{}
	vendorCounts := map[string]int{}

	for _, tx := range transactions {
		if tx.IsCredit() {
			continue
		}
		totalSpend += tx.Amount
		cat := tx.Category
		if cat == "" {
			cat = "Other"
		}
		categoryTotals[cat] += tx.Amount
		if tx.Vendor != "" {
			vendorTotals[tx.Vendor] += tx.Amount
			vendorCounts[tx.Vendor]++
		}
	}

	// Sort categories by amount descending
	type kv struct {
		Key string
		Val float64
	}
	var cats []kv
	for k, v := range categoryTotals {
		cats = append(cats, kv{k, v})
	}
	sort.Slice(cats, func(i, j int) bool { return cats[i].Val > cats[j].Val })

	// Top 10 vendors by amount
	var vendors []kv
	for k, v := range vendorTotals {
		vendors = append(vendors, kv{k, v})
	}
	sort.Slice(vendors, func(i, j int) bool { return vendors[i].Val > vendors[j].Val })
	if len(vendors) > 10 {
		vendors = vendors[:10]
	}

	var sb strings.Builder
	sb.WriteString("You are a personal finance assistant. Answer questions about the user's spending concisely.\n")
	sb.WriteString("Use ₹ for amounts. If data is insufficient to answer, say so.\n\n")
	fmt.Fprintf(&sb, "Today: %s\n\n", time.Now().Format("2 Jan 2006"))
	fmt.Fprintf(&sb, "Last 30 days summary (%d transactions):\n", len(transactions))
	fmt.Fprintf(&sb, "Total spend: ₹%.2f\n\n", totalSpend)

	sb.WriteString("By category:\n")
	for _, c := range cats {
		fmt.Fprintf(&sb, "  %s: ₹%.2f\n", c.Key, c.Val)
	}

	sb.WriteString("\nTop vendors:\n")
	for _, v := range vendors {
		fmt.Fprintf(&sb, "  %s: ₹%.2f (%d txns)\n", v.Key, v.Val, vendorCounts[v.Key])
	}

	return sb.String()
}
