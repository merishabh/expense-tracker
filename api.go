package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type AskGeminiRequest struct {
	Question string `json:"question"`
}

type AskGeminiResponse struct {
	Answer string `json:"answer"`
	Error  string `json:"error,omitempty"`
}

type ClassifyIntentRequest struct {
	Question string `json:"question"`
}

type ClassifyIntentResponse struct {
	Intent *ExpenseIntent `json:"intent"`
	Error  string         `json:"error,omitempty"`
}

func askGeminiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AskGeminiRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Question == "" {
		http.Error(w, "Question is required", http.StatusBadRequest)
		return
	}

	// Initialize database client
	dbClient, err := NewDatabaseClient()
	if err != nil {
		log.Printf("Failed to create database client: %v", err)
		resp := AskGeminiResponse{Error: "Database connection failed"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	defer dbClient.Close()

	// Fetch transactions
	transactions, err := dbClient.FetchAllTransactions()
	if err != nil {
		log.Printf("Failed to fetch transactions: %v", err)
		resp := AskGeminiResponse{Error: "Failed to fetch transactions"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Initialize Gemini client
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("GEMINI_API_KEY not set")
		resp := AskGeminiResponse{Error: "Gemini API key not configured"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	geminiClient := NewGeminiClient(apiKey)
	answer, err := geminiClient.AskGemini(transactions, req.Question)
	if err != nil {
		log.Printf("Gemini API error: %v", err)
		resp := AskGeminiResponse{Error: fmt.Sprintf("Failed to get answer: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := AskGeminiResponse{Answer: answer}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func classifyIntentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClassifyIntentRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Question == "" {
		http.Error(w, "Question is required", http.StatusBadRequest)
		return
	}

	// Initialize Gemini client (no database queries needed)
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("GEMINI_API_KEY not set")
		resp := ClassifyIntentResponse{Error: "Gemini API key not configured"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	geminiClient := NewGeminiClient(apiKey)
	intent, err := geminiClient.ClassifyIntent(req.Question)
	if err != nil {
		log.Printf("Intent classification error: %v", err)
		resp := ClassifyIntentResponse{Error: fmt.Sprintf("Failed to classify intent: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := ClassifyIntentResponse{Intent: intent}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type AnalyticsResponse struct {
	Analytics *SpendingAnalytics `json:"analytics"`
	Error     string             `json:"error,omitempty"`
}

type InsightsResponse struct {
	Insights []SpendingInsight `json:"insights"`
	Error    string            `json:"error,omitempty"`
}

type RecommendationsResponse struct {
	Recommendations []BudgetRecommendation `json:"recommendations"`
	Error           string                 `json:"error,omitempty"`
}

type ScoreResponse struct {
	Score       int    `json:"score"`
	Explanation string `json:"explanation"`
	Error       string `json:"error,omitempty"`
}

type PredictionsResponse struct {
	Predictions map[string]float64 `json:"predictions"`
	Error       string             `json:"error,omitempty"`
}

func analyticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Initialize database client
	dbClient, err := NewDatabaseClient()
	if err != nil {
		log.Printf("Failed to create database client: %v", err)
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer dbClient.Close()

	// Fetch all transactions
	transactions, err := dbClient.FetchAllTransactions()
	if err != nil {
		log.Printf("Failed to fetch transactions: %v", err)
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	// Initialize Gemini client for analytics
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("GEMINI_API_KEY not set")
		http.Error(w, "Gemini API key not configured", http.StatusInternalServerError)
		return
	}
	geminiClient := NewGeminiClient(apiKey)

	// Generate analytics
	analytics := geminiClient.AnalyzeTransactions(transactions)

	resp := AnalyticsResponse{Analytics: analytics}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func insightsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get transactions and generate analytics
	dbClient, err := NewDatabaseClient()
	if err != nil {
		resp := InsightsResponse{Error: "Database connection failed"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	defer dbClient.Close()

	transactions, err := dbClient.FetchAllTransactions()
	if err != nil {
		resp := InsightsResponse{Error: "Failed to fetch transactions"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		resp := InsightsResponse{Error: "Gemini API key not configured"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	geminiClient := NewGeminiClient(apiKey)
	analytics := geminiClient.AnalyzeTransactions(transactions)

	// Generate insights
	analyticsService := NewAnalyticsService()
	insights := analyticsService.GenerateInsights(analytics)

	resp := InsightsResponse{Insights: insights}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func recommendationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get transactions and generate analytics
	dbClient, err := NewDatabaseClient()
	if err != nil {
		resp := RecommendationsResponse{Error: "Database connection failed"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	defer dbClient.Close()

	transactions, err := dbClient.FetchAllTransactions()
	if err != nil {
		resp := RecommendationsResponse{Error: "Failed to fetch transactions"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		resp := RecommendationsResponse{Error: "Gemini API key not configured"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	geminiClient := NewGeminiClient(apiKey)
	analytics := geminiClient.AnalyzeTransactions(transactions)

	// Generate recommendations
	analyticsService := NewAnalyticsService()
	recommendations := analyticsService.GenerateBudgetRecommendations(analytics)

	resp := RecommendationsResponse{Recommendations: recommendations}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func scoreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get transactions and generate analytics
	dbClient, err := NewDatabaseClient()
	if err != nil {
		resp := ScoreResponse{Error: "Database connection failed"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	defer dbClient.Close()

	transactions, err := dbClient.FetchAllTransactions()
	if err != nil {
		resp := ScoreResponse{Error: "Failed to fetch transactions"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		resp := ScoreResponse{Error: "Gemini API key not configured"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	geminiClient := NewGeminiClient(apiKey)
	analytics := geminiClient.AnalyzeTransactions(transactions)

	// Calculate score
	analyticsService := NewAnalyticsService()
	score, explanation := analyticsService.CalculateSpendingScore(analytics)

	resp := ScoreResponse{Score: score, Explanation: explanation}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func predictionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get transactions and generate analytics
	dbClient, err := NewDatabaseClient()
	if err != nil {
		resp := PredictionsResponse{Error: "Database connection failed"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	defer dbClient.Close()

	transactions, err := dbClient.FetchAllTransactions()
	if err != nil {
		resp := PredictionsResponse{Error: "Failed to fetch transactions"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		resp := PredictionsResponse{Error: "Gemini API key not configured"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	geminiClient := NewGeminiClient(apiKey)
	analytics := geminiClient.AnalyzeTransactions(transactions)

	// Generate predictions
	analyticsService := NewAnalyticsService()
	predictions := analyticsService.PredictNextMonthSpending(analytics)

	resp := PredictionsResponse{Predictions: predictions}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// serveStaticFiles handles serving the frontend files
func serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	// Serve the index.html file for root path
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "static/index.html")
		return
	}

	// Remove the leading slash and serve from static directory
	path := r.URL.Path[1:]
	fullPath := filepath.Join("static", path)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// If file doesn't exist, serve index.html (for SPA routing)
		http.ServeFile(w, r, "static/index.html")
		return
	}

	// Serve the requested file
	http.ServeFile(w, r, fullPath)
}

func startAPIServer() {
	// API routes
	http.HandleFunc("/ask-gemini", askGeminiHandler)
	http.HandleFunc("/classify-intent", classifyIntentHandler)
	http.HandleFunc("/analytics", analyticsHandler)
	http.HandleFunc("/insights", insightsHandler)
	http.HandleFunc("/recommendations", recommendationsHandler)
	http.HandleFunc("/score", scoreHandler)
	http.HandleFunc("/predictions", predictionsHandler)

	// Static file serving
	http.HandleFunc("/", serveStaticFiles)

	log.Println("🚀 API Server starting on :8080")
	log.Println("🌐 Frontend available at: http://localhost:8080")
	log.Println("📊 Available API endpoints:")
	log.Println("  POST /ask-gemini          - Ask AI questions about your spending")
	log.Println("  POST /classify-intent     - Classify user question into validated intent JSON")
	log.Println("  GET  /analytics           - Get comprehensive spending analytics")
	log.Println("  GET  /insights            - Get spending insights and warnings")
	log.Println("  GET  /recommendations     - Get budget recommendations")
	log.Println("  GET  /score               - Get financial health score")
	log.Println("  GET  /predictions         - Get next month spending predictions")
	log.Println()
	log.Println("💡 Example usage:")
	log.Println("  curl -X POST http://localhost:8080/ask-gemini -H 'Content-Type: application/json' -d '{\"question\": \"How much did I spend on food?\"}'")
	log.Println("  curl -X POST http://localhost:8080/classify-intent -H 'Content-Type: application/json' -d '{\"question\": \"How much did I spend on food this month?\"}'")
	log.Println("  curl -X GET http://localhost:8080/insights")
	log.Println("  curl -X GET http://localhost:8080/score")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
