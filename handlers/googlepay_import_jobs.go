package handlers

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yourusername/expense-tracker/models"
	"github.com/yourusername/expense-tracker/services"
)

type googlePayImportJob struct {
	ID          string                          `json:"id"`
	Status      string                          `json:"status"`
	Error       string                          `json:"error,omitempty"`
	Summary     services.GooglePayImportSummary `json:"summary"`
	CreatedAt   time.Time                       `json:"created_at"`
	StartedAt   *time.Time                      `json:"started_at,omitempty"`
	CompletedAt *time.Time                      `json:"completed_at,omitempty"`
}

type googlePayImportManager struct {
	mu      sync.RWMutex
	seq     uint64
	imports map[string]*googlePayImportJob
}

var googlePayImports = &googlePayImportManager{
	imports: make(map[string]*googlePayImportJob),
}

func (m *googlePayImportManager) Start(content []byte) googlePayImportJob {
	id := fmt.Sprintf("gpay-%d-%d", time.Now().UnixNano(), atomic.AddUint64(&m.seq, 1))
	job := &googlePayImportJob{
		ID:        id,
		Status:    "queued",
		CreatedAt: time.Now().UTC(),
	}

	m.mu.Lock()
	m.imports[id] = job
	m.mu.Unlock()

	go m.run(job, content)

	return m.snapshot(job)
}

func (m *googlePayImportManager) Get(id string) (googlePayImportJob, bool) {
	m.mu.RLock()
	job, ok := m.imports[id]
	m.mu.RUnlock()
	if !ok {
		return googlePayImportJob{}, false
	}

	return m.snapshot(job), true
}

func (m *googlePayImportManager) run(job *googlePayImportJob, content []byte) {
	startedAt := time.Now().UTC()
	m.mu.Lock()
	job.Status = "running"
	job.StartedAt = &startedAt
	m.mu.Unlock()

	dbClient, err := models.NewDatabaseClient()
	if err != nil {
		m.fail(job, fmt.Sprintf("database connection failed: %v", err))
		return
	}
	defer dbClient.Close()

	summary, err := services.ImportGooglePayHTMLWithProgress(bytes.NewReader(content), dbClient, func(summary services.GooglePayImportSummary) {
		m.mu.Lock()
		job.Summary = summary
		m.mu.Unlock()
	})
	if err != nil {
		m.mu.Lock()
		job.Summary = summary
		m.mu.Unlock()
		m.fail(job, err.Error())
		return
	}

	completedAt := time.Now().UTC()
	m.mu.Lock()
	job.Status = "completed"
	job.Summary = summary
	job.CompletedAt = &completedAt
	m.mu.Unlock()

	log.Printf("google pay import completed job_id=%s imported=%d processed=%d", job.ID, summary.ImportedCount, summary.ProcessedCount)
}

func (m *googlePayImportManager) fail(job *googlePayImportJob, message string) {
	completedAt := time.Now().UTC()
	m.mu.Lock()
	job.Status = "failed"
	job.Error = message
	job.CompletedAt = &completedAt
	m.mu.Unlock()

	log.Printf("google pay import failed job_id=%s err=%s", job.ID, message)
}

func (m *googlePayImportManager) snapshot(job *googlePayImportJob) googlePayImportJob {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copy := *job
	return copy
}
