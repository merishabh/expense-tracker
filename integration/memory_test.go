//go:build integration

package integration

import (
	"os"
	"strings"
	"testing"

	"github.com/yourusername/expense-tracker/ai"
	"github.com/yourusername/expense-tracker/models"
	"github.com/yourusername/expense-tracker/services"
)

const testCollection = "chat_memories_test"

// testExecutor wires save_memory to the real MemoryService.
// All other tools return a stub — memory creation tests don't need transaction data.
func testExecutor(memorySvc *services.MemoryService) ai.ToolExecutor {
	return func(name string, input map[string]any) (string, error) {
		if name != "save_memory" {
			return "no transaction data available in test environment", nil
		}
		memType, _ := input["type"].(string)
		content, _ := input["content"].(string)
		if err := memorySvc.SaveMemory(memType, content); err != nil {
			return "", err
		}
		return "memory saved", nil
	}
}

func setupTest(t *testing.T) (*models.FirestoreClient, *services.MemoryService, *ai.ClaudeClient) {
	t.Helper()

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	db, err := models.NewFirestoreClient()
	if err != nil {
		t.Fatalf("firestore connect: %v", err)
	}
	db.WithMemoriesCollection(testCollection)

	// Clean slate before test, and guarantee cleanup after.
	if err := db.DeleteAllMemories(); err != nil {
		t.Fatalf("pre-test cleanup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.DeleteAllMemories(); err != nil {
			t.Logf("post-test cleanup failed: %v", err)
		}
		db.Close()
	})

	return db, services.NewMemoryService(db), ai.NewClaudeClient(apiKey)
}

// TestMemoryCreation_GoalStatement verifies that when a user states a financial
// goal, Claude calls save_memory with type="goal" and content referencing the goal.
func TestMemoryCreation_GoalStatement(t *testing.T) {
	db, memorySvc, claudeClient := setupTest(t)
	_ = db

	_, _, err := claudeClient.Chat(
		"I want to save ₹3 lakh for a Europe trip by June. Can you help me plan?",
		nil, "", testExecutor(memorySvc),
	)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	memories, err := memorySvc.LoadMemoriesRaw()
	if err != nil {
		t.Fatalf("LoadMemoriesRaw failed: %v", err)
	}

	// Check 1: at least one memory was saved.
	if len(memories) == 0 {
		t.Fatal("expected save_memory to be called, but no memories were saved")
	}

	// Check 2: one of them has type="goal".
	var goal *models.Memory
	for i, m := range memories {
		if m.Type == "goal" {
			goal = &memories[i]
			break
		}
	}
	if goal == nil {
		t.Fatalf("expected a memory with type=goal, got types: %v", memoryTypes(memories))
	}

	// Check 3: content is specific — not just "user wants to travel".
	lower := strings.ToLower(goal.Content)
	if !strings.Contains(lower, "europe") {
		t.Errorf("goal memory missing 'europe', got: %q", goal.Content)
	}
	if !strings.Contains(lower, "3") {
		t.Errorf("goal memory missing the ₹3L amount, got: %q", goal.Content)
	}

	t.Logf("✅ [%s] %s", goal.Type, goal.Content)
}

// TestMemoryCreation_SpendingPattern verifies that when a user describes a
// recurring behaviour, Claude saves a memory capturing that pattern.
func TestMemoryCreation_SpendingPattern(t *testing.T) {
	db, memorySvc, claudeClient := setupTest(t)
	_ = db

	_, _, err := claudeClient.Chat(
		"I always order Swiggy late at night on weekdays — it's become a habit I want to break.",
		nil, "", testExecutor(memorySvc),
	)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	memories, err := memorySvc.LoadMemoriesRaw()
	if err != nil {
		t.Fatalf("LoadMemoriesRaw failed: %v", err)
	}

	if len(memories) == 0 {
		t.Fatal("expected save_memory to be called, but no memories were saved")
	}

	// The content must reference the actual pattern — not a generic "user orders food".
	var found bool
	for _, m := range memories {
		lower := strings.ToLower(m.Content)
		if strings.Contains(lower, "swiggy") && strings.Contains(lower, "night") {
			found = true
			t.Logf("✅ [%s] %s", m.Type, m.Content)
			break
		}
	}
	if !found {
		t.Errorf("no memory captured the Swiggy late-night pattern, got: %+v", memories)
	}
}

func memoryTypes(memories []models.Memory) []string {
	types := make([]string, len(memories))
	for i, m := range memories {
		types[i] = m.Type
	}
	return types
}
