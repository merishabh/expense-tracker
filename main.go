package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yourusername/expense-tracker/handlers"
	"github.com/yourusername/expense-tracker/models"
	"github.com/yourusername/expense-tracker/services"
)

func main() {
	// Check if user wants to run API server
	if len(os.Args) > 1 && os.Args[1] == "api" {
		handlers.StartAPIServer()
		return
	}

	srv, err := handlers.InitGmailService()
	if err != nil {
		log.Fatalf("Unable to initialize Gmail service: %v", err)
	}

	fmt.Println("Fetching and processing emails...")

	dbClient, err := models.NewDatabaseClient()
	if err != nil {
		log.Fatalf("Failed to create database client: %v", err)
	}
	defer dbClient.Close()

	stats, err := services.ProcessEmails(srv, "me", dbClient)
	if err != nil {
		log.Fatalf("Email sync failed: %v", err)
	}

	log.Printf("Email sync completed: %+v", stats)
}
