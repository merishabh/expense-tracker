package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/yourusername/expense-tracker/models"
	"github.com/yourusername/expense-tracker/services"
)

func transactionsHandler(w http.ResponseWriter, r *http.Request) {
	reporting, cleanup, ok := newReportingService(w)
	if !ok {
		return
	}
	defer cleanup()

	limit := 50
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}

	transactions, err := reporting.ListTransactions(r.URL.Query().Get("period"), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"transactions": transactions})
}

func totalSummaryHandler(w http.ResponseWriter, r *http.Request) {
	reporting, cleanup, ok := newReportingService(w)
	if !ok {
		return
	}
	defer cleanup()

	summary, err := reporting.GetTotalSummary(r.URL.Query().Get("period"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func categorySummaryHandler(w http.ResponseWriter, r *http.Request) {
	reporting, cleanup, ok := newReportingService(w)
	if !ok {
		return
	}
	defer cleanup()

	items, err := reporting.GetCategoryBreakdown(r.URL.Query().Get("period"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func sourceSummaryHandler(w http.ResponseWriter, r *http.Request) {
	reporting, cleanup, ok := newReportingService(w)
	if !ok {
		return
	}
	defer cleanup()

	items, err := reporting.GetSourceBreakdown(r.URL.Query().Get("period"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func trendSummaryHandler(w http.ResponseWriter, r *http.Request) {
	reporting, cleanup, ok := newReportingService(w)
	if !ok {
		return
	}
	defer cleanup()

	points, err := reporting.GetDailyTrend(r.URL.Query().Get("period"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": points})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"mode":   "phase1-mvp",
	})
}

// serveStaticFiles handles serving the frontend files
func serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "frontend/index.html")
		return
	}

	path := r.URL.Path[1:]
	fullPath := filepath.Join("frontend", path)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.ServeFile(w, r, "frontend/index.html")
		return
	}

	http.ServeFile(w, r, fullPath)
}

func StartAPIServer() {
	http.HandleFunc("/api/health", healthHandler)
	http.HandleFunc("/api/transactions", transactionsHandler)
	http.HandleFunc("/api/summary/total", totalSummaryHandler)
	http.HandleFunc("/api/summary/category", categorySummaryHandler)
	http.HandleFunc("/api/summary/source", sourceSummaryHandler)
	http.HandleFunc("/api/summary/trend", trendSummaryHandler)

	http.HandleFunc("/", serveStaticFiles)

	log.Println("🚀 API Server starting on :8080")
	log.Println("🌐 Frontend available at: http://localhost:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func newReportingService(w http.ResponseWriter) (*services.ReportingService, func(), bool) {
	dbClient, err := models.NewDatabaseClient()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return nil, nil, false
	}

	return services.NewReportingService(dbClient), func() {
		dbClient.Close()
	}, true
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
