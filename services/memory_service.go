package services

import (
	"fmt"
	"strings"

	"github.com/yourusername/expense-tracker/models"
)

type MemoryService struct {
	db models.DatabaseClient
}

func NewMemoryService(db models.DatabaseClient) *MemoryService {
	return &MemoryService{db: db}
}

func (s *MemoryService) SaveMemory(memType, content string) error {
	return s.db.SaveMemory(models.Memory{
		Type:    memType,
		Content: content,
	})
}

// LoadMemories returns a formatted block ready to inject into the system prompt.
// Returns empty string if there are no memories.
func (s *MemoryService) LoadMemories() string {
	memories, err := s.db.GetAllMemories()
	if err != nil || len(memories) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, m := range memories {
		fmt.Fprintf(&sb, "- [%s] %s\n", m.Type, m.Content)
	}
	return sb.String()
}

// LoadMemoriesRaw returns the raw memory slice. Used by integration tests to
// assert on individual memory fields after a Chat call.
func (s *MemoryService) LoadMemoriesRaw() ([]models.Memory, error) {
	return s.db.GetAllMemories()
}
