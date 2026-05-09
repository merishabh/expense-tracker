package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/yourusername/expense-tracker/ai"
	"github.com/yourusername/expense-tracker/models"
	"github.com/yourusername/expense-tracker/services"
)

const evalCollection = "chat_memories_test"

const evalJudgePrompt = `You score a personal finance agent's answer on personalisation quality.
Return ONLY valid JSON, no other text: {"score": 0.0, "reason": "one sentence"}

Scoring guide:
1.0 — uses specific numbers, named merchants (Swiggy/Amazon etc), references this person's actual patterns. Feels like personal advice.
0.7 — somewhat specific but contains some generic elements
0.5 — half personalised, half generic
0.3 — mostly generic with minor specific elements
0.0 — completely generic advice OR invented facts not present in the profile`

type evalResult struct {
	Name    string
	Passed  bool
	Score   float64 // -1 means no judge (memory creation tests)
	Reason  string
	Details string
	Elapsed time.Duration
}

type judgeScore struct {
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

func evalJudge(apiKey, question, answer string, memories []models.Memory, profileEmpty bool) (judgeScore, error) {
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
		System:    []anthropic.TextBlockParam{{Text: evalJudgePrompt}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(userMsg))},
	})
	if err != nil {
		return judgeScore{Score: 0.5, Reason: "judge call failed"}, err
	}

	raw := strings.TrimSpace(resp.Content[0].AsText().Text)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var result judgeScore
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return judgeScore{Score: 0.5, Reason: "judge parse failed"}, nil
	}
	return result, nil
}

func evalMemoryOnlyExecutor(memorySvc *services.MemoryService) ai.ToolExecutor {
	return func(name string, input map[string]any) (string, error) {
		if name != "save_memory" {
			return "no transaction data available in eval environment", nil
		}
		memType, _ := input["type"].(string)
		content, _ := input["content"].(string)
		if err := memorySvc.SaveMemory(memType, content); err != nil {
			return "", err
		}
		return "memory saved", nil
	}
}

func runMemoryCreationEval(apiKey string, db *models.FirestoreClient, claudeClient *ai.ClaudeClient, name, question, wantType string, wantContent []string) evalResult {
	start := time.Now()
	memorySvc := services.NewMemoryService(db)

	if err := db.DeleteAllMemories(); err != nil {
		return evalResult{Name: name, Passed: false, Score: -1, Details: "cleanup failed: " + err.Error(), Elapsed: time.Since(start)}
	}

	_, _, err := claudeClient.Chat(question, nil, "", evalMemoryOnlyExecutor(memorySvc))
	if err != nil {
		return evalResult{Name: name, Passed: false, Score: -1, Details: "chat failed: " + err.Error(), Elapsed: time.Since(start)}
	}

	memories, err := memorySvc.LoadMemoriesRaw()
	if err != nil || len(memories) == 0 {
		return evalResult{Name: name, Passed: false, Score: -1, Details: "no memory saved", Elapsed: time.Since(start)}
	}

	var matched *models.Memory
	for i, m := range memories {
		if m.Type == wantType {
			matched = &memories[i]
			break
		}
	}
	if matched == nil {
		types := make([]string, len(memories))
		for i, m := range memories {
			types[i] = m.Type
		}
		return evalResult{Name: name, Passed: false, Score: -1,
			Details: fmt.Sprintf("no memory with type=%s, got types: %v", wantType, types),
			Elapsed: time.Since(start)}
	}

	lower := strings.ToLower(matched.Content)
	for _, want := range wantContent {
		if !strings.Contains(lower, strings.ToLower(want)) {
			return evalResult{Name: name, Passed: false, Score: -1,
				Details: fmt.Sprintf("memory content missing %q: got %q", want, matched.Content),
				Elapsed: time.Since(start)}
		}
	}

	return evalResult{
		Name:    name,
		Passed:  true,
		Score:   -1,
		Details: fmt.Sprintf("[%s] %s", matched.Type, matched.Content),
		Elapsed: time.Since(start),
	}
}

func runPersonalisationEval(apiKey string, db *models.FirestoreClient, claudeClient *ai.ClaudeClient, name, question string, seeds []models.Memory, minScore float64, forbidden []string) evalResult {
	start := time.Now()
	memorySvc := services.NewMemoryService(db)

	if err := db.DeleteAllMemories(); err != nil {
		return evalResult{Name: name, Passed: false, Score: 0, Details: "cleanup failed: " + err.Error(), Elapsed: time.Since(start)}
	}

	for _, m := range seeds {
		if err := memorySvc.SaveMemory(m.Type, m.Content); err != nil {
			return evalResult{Name: name, Passed: false, Score: 0, Details: "seed failed: " + err.Error(), Elapsed: time.Since(start)}
		}
	}

	memBlock := memorySvc.LoadMemories()
	answer, _, err := claudeClient.Chat(question, nil, memBlock, evalMemoryOnlyExecutor(memorySvc))
	if err != nil {
		return evalResult{Name: name, Passed: false, Score: 0, Details: "chat failed: " + err.Error(), Elapsed: time.Since(start)}
	}

	lower := strings.ToLower(answer)
	for _, f := range forbidden {
		if strings.Contains(lower, strings.ToLower(f)) {
			return evalResult{Name: name, Passed: false, Score: 0,
				Details: fmt.Sprintf("hallucinated forbidden term %q", f),
				Elapsed: time.Since(start)}
		}
	}

	judgment, err := evalJudge(apiKey, question, answer, seeds, len(seeds) == 0)
	if err != nil {
		return evalResult{Name: name, Passed: false, Score: judgment.Score, Details: "judge error: " + err.Error(), Elapsed: time.Since(start)}
	}

	passed := judgment.Score >= minScore
	details := judgment.Reason
	if !passed {
		details = fmt.Sprintf("score %.2f below threshold %.2f — %s", judgment.Score, minScore, judgment.Reason)
	}

	return evalResult{Name: name, Passed: passed, Score: judgment.Score, Details: details, Elapsed: time.Since(start)}
}

func runAllEvals(apiKey string) []evalResult {
	db, err := models.NewFirestoreClient()
	if err != nil {
		return []evalResult{{Name: "setup", Passed: false, Details: "firestore connect failed: " + err.Error()}}
	}
	defer db.Close()
	db.WithMemoriesCollection(evalCollection)

	claudeClient := ai.NewClaudeClient(apiKey)
	var results []evalResult

	results = append(results, runMemoryCreationEval(apiKey, db, claudeClient,
		"Memory creation: goal statement",
		"I want to save ₹3 lakh for a Europe trip by June. Can you help me plan?",
		"goal", []string{"europe", "3"},
	))

	results = append(results, runMemoryCreationEval(apiKey, db, claudeClient,
		"Memory creation: spending pattern",
		"I always order Swiggy late at night on weekdays — it's become a habit I want to break.",
		"pattern", []string{"swiggy", "night"},
	))

	results = append(results, runPersonalisationEval(apiKey, db, claudeClient,
		"Personalisation: profiled user",
		"What should I watch out for this month?",
		[]models.Memory{
			{Type: "goal", Content: "User wants to save ₹3L for a Europe trip by June"},
			{Type: "pattern", Content: "User orders Swiggy late at night on weekdays"},
			{Type: "life_context", Content: "User's top spending categories are Food, Shopping, and Travel"},
		},
		0.75, nil,
	))

	results = append(results, runPersonalisationEval(apiKey, db, claudeClient,
		"Personalisation: no memory — no hallucination",
		"Am I a foodie?",
		nil, 0.75,
		[]string{"₹1,91", "₹191", "₹50,000"},
	))

	results = append(results, runPersonalisationEval(apiKey, db, claudeClient,
		"Personalisation: goal referenced",
		"How should I plan my savings?",
		[]models.Memory{
			{Type: "goal", Content: "User wants to save ₹3L for a Europe trip by June"},
		},
		0.80, nil,
	))

	// Final cleanup
	_ = db.DeleteAllMemories()
	return results
}

var evalPageTmpl = template.Must(template.New("evals").Funcs(template.FuncMap{
	"scoreStr": func(s float64) string {
		if s < 0 {
			return "—"
		}
		return fmt.Sprintf("%.2f", s)
	},
	"statusClass": func(passed bool) string {
		if passed {
			return "pass"
		}
		return "fail"
	},
	"icon": func(passed bool) string {
		if passed {
			return "✅"
		}
		return "❌"
	},
}).Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Memory Evals</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 860px; margin: 48px auto; padding: 0 24px; color: #1a1a1a; }
  h1 { font-size: 1.4rem; margin-bottom: 4px; }
  .meta { color: #666; font-size: 0.85rem; margin-bottom: 24px; }
  table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
  th { text-align: left; padding: 10px 12px; border-bottom: 2px solid #e0e0e0; color: #444; }
  td { padding: 10px 12px; border-bottom: 1px solid #eee; vertical-align: top; }
  tr.pass td:first-child { border-left: 3px solid #22c55e; }
  tr.fail td:first-child { border-left: 3px solid #ef4444; }
  .score { font-variant-numeric: tabular-nums; }
  .details { color: #555; font-size: 0.82rem; }
  .elapsed { color: #999; font-size: 0.8rem; }
  .summary { margin-top: 20px; padding: 14px 16px; background: #f5f5f5; border-radius: 6px; font-size: 0.9rem; }
  .solid { color: #16a34a; font-weight: 600; }
  .needs-work { color: #dc2626; font-weight: 600; }
</style>
</head>
<body>
<h1>Phase 2 Memory Evals</h1>
<div class="meta">Run at {{.RunAt}}</div>
<table>
  <thead><tr><th>Test</th><th>Score</th><th>Details</th><th>Time</th></tr></thead>
  <tbody>
  {{range .Results}}
  <tr class="{{statusClass .Passed}}">
    <td>{{icon .Passed}} {{.Name}}</td>
    <td class="score">{{scoreStr .Score}}</td>
    <td class="details">{{.Details}}</td>
    <td class="elapsed">{{.Elapsed.Round 10000000}}</td>
  </tr>
  {{end}}
  </tbody>
</table>
<div class="summary">
  {{.Passed}}/{{.Total}} passed &nbsp;·&nbsp; avg score {{.AvgScore}}
  &nbsp;&nbsp;
  {{if .Solid}}<span class="solid">Phase 2 memory is solid ✅</span>{{else}}<span class="needs-work">Phase 2 needs work ❌</span>{{end}}
</div>
</body>
</html>`))

func evalRunHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		http.Error(w, "ANTHROPIC_API_KEY not configured", http.StatusInternalServerError)
		return
	}

	results := runAllEvals(apiKey)

	passed := 0
	totalScore := 0.0
	scoreCount := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
		if r.Score >= 0 {
			totalScore += r.Score
			scoreCount++
		}
	}

	avgScore := 0.0
	if scoreCount > 0 {
		avgScore = totalScore / float64(scoreCount)
	}

	data := struct {
		RunAt    string
		Results  []evalResult
		Passed   int
		Total    int
		AvgScore string
		Solid    bool
	}{
		RunAt:    time.Now().Format("2 Jan 2006, 15:04:05"),
		Results:  results,
		Passed:   passed,
		Total:    len(results),
		AvgScore: fmt.Sprintf("%.2f", avgScore),
		Solid:    avgScore >= 0.75,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := evalPageTmpl.Execute(w, data); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}
