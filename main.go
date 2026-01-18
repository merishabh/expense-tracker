package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yourusername/expense-tracker/ai"
	"github.com/yourusername/expense-tracker/handlers"
	"github.com/yourusername/expense-tracker/models"
	"github.com/yourusername/expense-tracker/services"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func main() {
	// Check if user wants to run API server
	if len(os.Args) > 1 && os.Args[1] == "api" {
		handlers.StartAPIServer()
		return
	}

	// Original email processing functionality
	b, err := os.ReadFile("credentials/client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client_secret.json: %v", err)
	}

	handlers.Config, err = google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret: %v", err)
	}

	client := handlers.GetClient()

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to create Gmail client: %v", err)
	}

	fmt.Println("Fetching and processing emails...")

	// Create database client (will choose MongoDB or Firestore based on environment)
	dbClient, err := models.NewDatabaseClient()
	if err != nil {
		log.Fatalf("Failed to create database client: %v", err)
	}
	defer dbClient.Close()

	// Create Groq client for AI-powered vendor categorization (optional)
	var groqClient *ai.GroqClient
	if apiKey := os.Getenv("GROQ_API_KEY"); apiKey != "" {
		groqClient = ai.NewGroqClient(apiKey)
		fmt.Println("Groq AI client initialized for smart vendor categorization")
	} else {
		fmt.Println("GROQ_API_KEY not found - AI vendor categorization disabled")
	}

	services.ProcessEmails(srv, "me", dbClient, groqClient)
}
