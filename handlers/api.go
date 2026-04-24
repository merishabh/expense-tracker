package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

	transactions, err := reporting.ListTransactions(
		r.URL.Query().Get("period"),
		r.URL.Query().Get("category"),
		limit,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"transactions": transactions})
}

func transactionsByRangeHandler(w http.ResponseWriter, r *http.Request) {
	reporting, cleanup, ok := newReportingService(w)
	if !ok {
		return
	}
	defer cleanup()

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		http.Error(w, "query params 'from' and 'to' are required (format: 2006-01-02)", http.StatusBadRequest)
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		http.Error(w, "invalid 'from' date, expected format: 2006-01-02", http.StatusBadRequest)
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		http.Error(w, "invalid 'to' date, expected format: 2006-01-02", http.StatusBadRequest)
		return
	}
	// include the full last day
	to = to.Add(24*time.Hour - time.Nanosecond)

	limit := 500
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	transactions, err := reporting.ListTransactionsByDateRange(from, to, r.URL.Query().Get("category"), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"from":         fromStr,
		"to":           r.URL.Query().Get("to"),
		"count":        len(transactions),
		"transactions": transactions,
	})
}

func lastTenDaysTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	reporting, cleanup, ok := newReportingService(w)
	if !ok {
		return
	}
	defer cleanup()

	transactions, err := reporting.GetLastNDaysTransactions(10, 200)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i := range transactions {
		fmt.Println(transactions[i])
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

func lastTenDaysTrendHandler(w http.ResponseWriter, r *http.Request) {
	reporting, cleanup, ok := newReportingService(w)
	if !ok {
		return
	}
	defer cleanup()

	points, err := reporting.GetLastNDaysTrend(10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": points})
}

func monthlyComparisonHandler(w http.ResponseWriter, r *http.Request) {
	reporting, cleanup, ok := newReportingService(w)
	if !ok {
		return
	}
	defer cleanup()

	comparison, err := reporting.GetMonthlyComparison()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, comparison)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"mode":   "phase1-mvp",
	})
}

func updateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Only PUT method allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		ID       string  `json:"id"`
		Type     string  `json:"type"`
		Vendor   string  `json:"vendor"`
		Amount   float64 `json:"amount"`
		Category string  `json:"category"`
		DateTime string  `json:"date_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.ID == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}
	if body.Vendor == "" || body.Amount == 0 {
		http.Error(w, "vendor and non-zero amount are required", http.StatusBadRequest)
		return
	}

	dt, err := time.Parse("2006-01-02T15:04", body.DateTime)
	if err != nil {
		dt, err = time.Parse("2006-01-02", body.DateTime)
		if err != nil {
			http.Error(w, "invalid date_time format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	tx := models.Transaction{
		Type:     body.Type,
		Vendor:   body.Vendor,
		Amount:   body.Amount,
		Category: body.Category,
		DateTime: dt,
	}

	dbClient, err := models.NewDatabaseClient()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer dbClient.Close()

	if err := dbClient.UpdateTransaction(body.ID, tx); err != nil {
		log.Printf("transaction update failed id=%s err=%v", body.ID, err)
		http.Error(w, "Failed to update transaction", http.StatusInternalServerError)
		return
	}

	log.Printf("transaction updated id=%s vendor=%q amount=%.2f category=%q", body.ID, tx.Vendor, tx.Amount, tx.Category)
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok"})
}

func addManualTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Type     string  `json:"type"`
		Vendor   string  `json:"vendor"`
		Amount   float64 `json:"amount"`
		Category string  `json:"category"`
		DateTime string  `json:"date_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.Vendor == "" || body.Amount == 0 {
		http.Error(w, "vendor and non-zero amount are required", http.StatusBadRequest)
		return
	}

	dt, err := time.Parse("2006-01-02T15:04", body.DateTime)
	if err != nil {
		dt, err = time.Parse("2006-01-02", body.DateTime)
		if err != nil {
			http.Error(w, "invalid date_time format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	txType := body.Type
	if txType == "" {
		txType = "Manual"
	}
	tx := models.Transaction{
		Type:     txType,
		Vendor:   body.Vendor,
		Amount:   body.Amount,
		Category: body.Category,
		DateTime: dt,
	}

	dbClient, err := models.NewDatabaseClient()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer dbClient.Close()

	if err := dbClient.SaveTransaction(tx); err != nil {
		log.Printf("manual transaction save failed vendor=%q amount=%.2f err=%v", tx.Vendor, tx.Amount, err)
		http.Error(w, "Failed to save transaction", http.StatusInternalServerError)
		return
	}

	log.Printf("manual transaction saved vendor=%q amount=%.2f category=%q", tx.Vendor, tx.Amount, tx.Category)
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "transaction": tx})
}

func syncHDFCHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("bank email sync requested method=%s path=%s remote_addr=%s user_agent=%q", r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())

	srv, err := InitGmailService()
	if err != nil {
		log.Printf("bank email sync failed during gmail init err=%v", err)
		http.Error(w, fmt.Sprintf("Failed to initialize Gmail service: %v", err), http.StatusInternalServerError)
		return
	}

	dbClient, err := models.NewDatabaseClient()
	if err != nil {
		log.Printf("bank email sync failed during db init err=%v", err)
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer dbClient.Close()

	stats, err := services.ProcessEmails(srv, "me", dbClient)
	if err != nil {
		log.Printf("bank email sync failed err=%v stats=%+v", err, stats)
		http.Error(w, fmt.Sprintf("Gmail sync failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("bank email sync completed stats=%+v", stats)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"job":    "bank_email_sync",
		"stats":  stats,
	})
}

func importGooglePayHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		startGooglePayImportHandler(w, r)
	case http.MethodGet:
		googlePayImportStatusHandler(w, r)
	default:
		http.Error(w, "Only GET and POST methods allowed", http.StatusMethodNotAllowed)
	}
}

func startGooglePayImportHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "invalid upload, expected multipart form with HTML file", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read uploaded file", http.StatusBadRequest)
		return
	}

	job := googlePayImports.Start(content)

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"status": "accepted",
		"job":    "google_pay_import",
		"job_id": job.ID,
		"import": job,
	})
}

func googlePayImportStatusHandler(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("id")
	if jobID == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	job, ok := googlePayImports.Get(jobID)
	if !ok {
		http.Error(w, "import job not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"job":    "google_pay_import",
		"import": job,
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
	InitWebAuth()

	// Auth routes (no middleware)
	http.HandleFunc("/auth/signin", loginPageHandler)
	http.HandleFunc("/auth/login", loginHandler)
	http.HandleFunc("/auth/callback", callbackHandler)
	http.HandleFunc("/auth/logout", logoutHandler)

	// Public health check
	http.HandleFunc("/api/health", healthHandler)

	// Protected API routes
	http.HandleFunc("/api/jobs/sync-hdfc", syncHDFCHandler)
	http.HandleFunc("/api/transactions/manual", apiAuthMiddleware(addManualTransactionHandler))
	http.HandleFunc("/api/transactions/update", apiAuthMiddleware(updateTransactionHandler))
	http.HandleFunc("/api/import/google-pay", apiAuthMiddleware(importGooglePayHandler))
	http.HandleFunc("/api/transactions", apiAuthMiddleware(transactionsHandler))
	http.HandleFunc("/api/transactions/range", apiAuthMiddleware(transactionsByRangeHandler))
	http.HandleFunc("/api/transactions/last-10-days", apiAuthMiddleware(lastTenDaysTransactionsHandler))
	http.HandleFunc("/api/summary/total", apiAuthMiddleware(totalSummaryHandler))
	http.HandleFunc("/api/summary/category", apiAuthMiddleware(categorySummaryHandler))
	http.HandleFunc("/api/summary/source", apiAuthMiddleware(sourceSummaryHandler))
	http.HandleFunc("/api/summary/trend", apiAuthMiddleware(trendSummaryHandler))
	http.HandleFunc("/api/summary/trend/last-10-days", apiAuthMiddleware(lastTenDaysTrendHandler))
	http.HandleFunc("/api/summary/monthly-comparison", apiAuthMiddleware(monthlyComparisonHandler))

	// Protected frontend
	http.HandleFunc("/", webAuthMiddleware(serveStaticFiles))

	log.Println("API Server starting on :8080")
	log.Println("Frontend available at: http://localhost:8080")

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
