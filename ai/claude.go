package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const systemPrompt = `You are a personal finance assistant. You have tools to query the user's transaction data.

Today: ` + "TODAY_PLACEHOLDER" + `

Rules:
- Always call a tool to fetch real data before answering. Never guess amounts.
- Use ₹ for all amounts.
- Be concise. Lead with the numbers, follow with brief insight.
- If a question spans multiple periods or categories, call the relevant tool for each.`

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

func (c *ClaudeClient) Chat(question string, history []ChatMessage, memories string, executor ToolExecutor) (string, error) {
	memoryBlock := ""
	if memories != "" {
		memoryBlock = "\n\nWhat you know about this user:\n" + memories
	}

	prompt := fmt.Sprintf(`You are a personal finance assistant. You have tools to query the user's transaction data.

Today: %s%s

Rules:
- Always call a tool to fetch real data before answering. Never guess amounts.
- Use ₹ for all amounts.
- Be concise. Lead with the numbers, follow with brief insight.
- If a question spans multiple periods or categories, call the relevant tool for each.

Memory rules (call save_memory proactively):
- User states a financial goal, e.g. "I want to spend less on food"
- User corrects an assumption you made
- User reveals life context, e.g. income, job, family situation, city
- You notice a strong spending pattern worth remembering
- User expresses a preference for how they want analysis presented`,
		time.Now().Format("2 Jan 2006"), memoryBlock)

	// Seed messages from conversation history
	messages := make([]anthropic.MessageParam, 0, len(history)+1)
	for _, h := range history {
		if h.Role == "user" {
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(h.Content)))
		} else {
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(h.Content)))
		}
	}
	messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(question)))

	log.Printf("claude chat start: question=%q history_len=%d", question, len(history))

	// ── Agentic loop ──────────────────────────────────────────────────────────
	// Claude may call tools multiple times before producing a final text answer.
	// Each iteration: call API → if tool_use, execute tools and append results → repeat.
	for iteration := 0; iteration < 10; iteration++ {
		msg, err := c.client.Messages.New(context.Background(), anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_6,
			MaxTokens: 1024,
			System:    []anthropic.TextBlockParam{{Text: prompt}},
			Tools:     toolSchemas,
			Messages:  messages,
		})
		if err != nil {
			return "", fmt.Errorf("claude API error: %w", err)
		}

		log.Printf("claude iteration=%d stop_reason=%s content_blocks=%d", iteration, msg.StopReason, len(msg.Content))

		// ── end_turn: Claude is done, extract the text answer ─────────────────
		if msg.StopReason == "end_turn" {
			for _, block := range msg.Content {
				if block.Type == "text" {
					return block.Text, nil
				}
			}
			return "", fmt.Errorf("end_turn but no text block in response")
		}

		// ── tool_use: execute each requested tool, collect results ─────────────
		if msg.StopReason == "tool_use" {
			// Convert response content blocks to param blocks for the next request
			var assistantBlocks []anthropic.ContentBlockParamUnion
			for _, b := range msg.Content {
				switch b.Type {
				case "text":
					assistantBlocks = append(assistantBlocks, anthropic.NewTextBlock(b.AsText().Text))
				case "tool_use":
					tu := b.AsToolUse()
					assistantBlocks = append(assistantBlocks, anthropic.NewToolUseBlock(tu.ID, tu.Input, tu.Name))
				}
			}
			messages = append(messages, anthropic.NewAssistantMessage(assistantBlocks...))

			// Build tool_result blocks for every tool Claude called
			var toolResults []anthropic.ContentBlockParamUnion
			for _, block := range msg.Content {
				if block.Type != "tool_use" {
					continue
				}

				// Unmarshal the tool input JSON into a plain map
				var input map[string]any
				if err := json.Unmarshal([]byte(block.JSON.Input.Raw()), &input); err != nil {
					input = map[string]any{}
				}

				log.Printf("claude tool_call: tool=%s input=%v", block.Name, input)

				result, toolErr := executor(block.Name, input)
				if toolErr != nil {
					result = fmt.Sprintf("error: %v", toolErr)
				}

				log.Printf("claude tool_result: tool=%s result=%s", block.Name, result)

				toolResults = append(toolResults, anthropic.NewToolResultBlock(block.ID, result, toolErr != nil))
			}

			// Append tool results as a user turn so Claude can reason over them
			messages = append(messages, anthropic.NewUserMessage(toolResults...))
			continue
		}

		// Unexpected stop reason
		return "", fmt.Errorf("unexpected stop_reason: %s", msg.StopReason)
	}

	return "", fmt.Errorf("agentic loop exceeded max iterations")
}
