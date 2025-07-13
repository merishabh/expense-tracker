package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func main() {
	// Check if user wants to run API server
	if len(os.Args) > 1 && os.Args[1] == "api" {
		startAPIServer()
		return
	}

	// Original email processing functionality
	b, err := os.ReadFile("credentials/client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client_secret.json: %v", err)
	}

	config, err = google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret: %v", err)
	}

	client := getClient()

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to create Gmail client: %v", err)
	}

	fmt.Println("Fetching and processing emails...")

	// Create database client (will choose MongoDB or Firestore based on environment)
	dbClient, err := NewDatabaseClient()
	if err != nil {
		log.Fatalf("Failed to create database client: %v", err)
	}
	defer dbClient.Close()

	// Create Gemini client for AI-powered vendor categorization (optional)
	var geminiClient *GeminiClient
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		geminiClient = NewGeminiClient(apiKey)
		fmt.Println("Gemini AI client initialized for smart vendor categorization")
	} else {
		fmt.Println("GEMINI_API_KEY not found - AI vendor categorization disabled")
	}

	processEmails(srv, "me", dbClient, geminiClient)
}
