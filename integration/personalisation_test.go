//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/yourusername/expense-tracker/models"
	"github.com/yourusername/expense-tracker/services"
)

const judgeSystemPrompt = `You score a personal finance agent's answer on personalisation quality.
Return ONLY valid JSON, no other text: {"score": 0.0, "reason": "one sentence"}

Scoring guide:
1.0 — uses specific numbers, named merchants (Swiggy/Amazon etc), references this person's actual patterns. Feels like personal advice.
0.7 — somewhat specific but contains some generic elements
0.5 — half personalised, half generic
0.3 — mostly generic with minor specific elements
0.0 — completely generic advice OR invented facts not present in the profile`

type judgeResult struct {
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

// judge calls a separate Claude instance to score the agent's answer.
// It never shares a client with the agent under test.
func judge(t *testing.T, question, answer string, memories []models.Memory, profileEmpty bool) judgeResult {
	t.Helper()

	apiKey := getEnv(t, "ANTHROPIC_API_KEY")
	c := anthropic.NewClient(option.WithAPIKey(apiKey))

	profileSection := "Profile available to the agent: EMPTY"
	if !profileEmpty && len(memories) > 0 {
		var sb strings.Builder
		for _, m := range memories {
			fmt.Fprintf(&sb, "- [%s] %s\n", m.Type, m.Content)
		}
		profileSection = "Profile available to the agent:\n" + sb.String()
	}

	userMsg := fmt.Sprintf("Question asked: %s\n\n%s\n\nAgent's answer:\n%s", question, profileSection, answer)

	resp, err := c.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 256,
		System:    []anthropic.TextBlockParam{{Text: judgeSystemPrompt}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(userMsg))},
	})
	if err != nil {
		t.Fatalf("judge call failed: %v", err)
	}

	raw := strings.TrimSpace(resp.Content[0].AsText().Text)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var result judgeResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Logf("judge parse failed on: %s", raw)
		return judgeResult{Score: 0.5, Reason: "judge parse failed"}
	}
	return result
}

func getEnv(t *testing.T, key string) string {
	t.Helper()
	val := os.Getenv(key)
	if val == "" {
		t.Skipf("%s not set", key)
	}
	return val
}

// seedMemories writes a set of memories into the test collection and returns them.
func seedMemories(t *testing.T, memorySvc *services.MemoryService, mems []models.Memory) {
	t.Helper()
	for _, m := range mems {
		if err := memorySvc.SaveMemory(m.Type, m.Content); err != nil {
			t.Fatalf("seed memory failed: %v", err)
		}
	}
}

// TestPersonalisation_ProfiledUser seeds goal + pattern memories and asserts
// the agent's answer is personalised enough to score >= 0.75.
func TestPersonalisation_ProfiledUser(t *testing.T) {
	db, memorySvc, claudeClient := setupTest(t)
	_ = db

	memories := []models.Memory{
		{Type: "goal", Content: "User wants to save ₹3L for a Europe trip by June"},
		{Type: "pattern", Content: "User orders Swiggy late at night on weekdays"},
		{Type: "life_context", Content: "User's top spending categories are Food, Shopping, and Travel"},
	}
	seedMemories(t, memorySvc, memories)

	memBlock := memorySvc.LoadMemories()
	answer, _, err := claudeClient.Chat("What should I watch out for this month?", nil, memBlock, testExecutor(memorySvc))
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	t.Logf("agent answer:\n%s", answer)

	result := judge(t, "What should I watch out for this month?", answer, memories, false)
	t.Logf("judge score=%.2f reason=%s", result.Score, result.Reason)

	if result.Score < 0.75 {
		t.Errorf("expected personalisation score >= 0.75, got %.2f: %s", result.Score, result.Reason)
	}
}

// TestPersonalisation_NoMemory_NoHallucination seeds nothing and asserts the
// agent does not invent specific rupee amounts it couldn't know.
func TestPersonalisation_NoMemory_NoHallucination(t *testing.T) {
	db, memorySvc, claudeClient := setupTest(t)
	_ = db

	// No seed — collection is empty.
	answer, _, err := claudeClient.Chat("Am I a foodie?", nil, "", testExecutor(memorySvc))
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	t.Logf("agent answer:\n%s", answer)

	// Hard check: must not contain specific rupee amounts it couldn't know.
	lower := strings.ToLower(answer)
	for _, forbidden := range []string{"₹1,91", "₹191", "₹3,", "₹50,000"} {
		if strings.Contains(lower, strings.ToLower(forbidden)) {
			t.Errorf("answer hallucinated %q which was not in the profile", forbidden)
		}
	}

	// Judge check: profile was empty — a good answer acknowledges lack of data.
	result := judge(t, "Am I a foodie?", answer, nil, true)
	t.Logf("judge score=%.2f reason=%s", result.Score, result.Reason)

	if result.Score < 0.75 {
		t.Errorf("expected score >= 0.75 for honest no-data response, got %.2f: %s", result.Score, result.Reason)
	}
}

// TestPersonalisation_GoalReferenced seeds only a savings goal and asserts the
// agent references it directly. Higher bar (0.8) since the goal is directly relevant.
func TestPersonalisation_GoalReferenced(t *testing.T) {
	db, memorySvc, claudeClient := setupTest(t)
	_ = db

	memories := []models.Memory{
		{Type: "goal", Content: "User wants to save ₹3L for a Europe trip by June"},
	}
	seedMemories(t, memorySvc, memories)

	memBlock := memorySvc.LoadMemories()
	answer, _, err := claudeClient.Chat("How should I plan my savings?", nil, memBlock, testExecutor(memorySvc))
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	t.Logf("agent answer:\n%s", answer)

	// Hard check before calling the judge.
	lower := strings.ToLower(answer)
	if !strings.Contains(lower, "europe") && !strings.Contains(lower, "₹3") && !strings.Contains(lower, "3 lakh") {
		t.Errorf("answer did not reference the Europe goal at all: %q", answer)
	}

	result := judge(t, "How should I plan my savings?", answer, memories, false)
	t.Logf("judge score=%.2f reason=%s", result.Score, result.Reason)

	if result.Score < 0.80 {
		t.Errorf("expected personalisation score >= 0.80, got %.2f: %s", result.Score, result.Reason)
	}
}
