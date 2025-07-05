package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
)

var (
	config *oauth2.Config
	state  = "randomstatestring"
)

func StartAuthServer() *oauth2.Token {
	codeCh := make(chan string)
	srv := &http.Server{Addr: ":8080"}

	http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "State doesn't match", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		fmt.Fprintln(w, "Authorization successful! You can close this window.")
		codeCh <- code
	})

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Open this URL in your browser:\n%v\n", authURL)

	code := <-codeCh
	srv.Shutdown(context.Background())

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to exchange code for token: %v", err)
	}

	saveToken("credentials/token.json", token)
	return token
}

func getClient() *http.Client {
	tokFile := "credentials/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		// Check if we're in a non-interactive environment (cron job)
		if isNonInteractiveEnvironment() {
			log.Fatalf("❌ No valid token found and running in non-interactive environment (cron job). Please run './refresh-token.sh' first to get a valid token.")
		}
		// Interactive environment - start OAuth flow
		tok = StartAuthServer()
	}

	// Check if token needs refresh
	if tok.Expiry.Before(time.Now()) {
		fmt.Println("🔄 Token expired, attempting to refresh...")
		tok, err = refreshToken(tok)
		if err != nil {
			fmt.Printf("❌ Token refresh failed: %v\n", err)
			if isNonInteractiveEnvironment() {
				log.Fatalf("❌ Token refresh failed in non-interactive environment. Please run './refresh-token.sh' to get a fresh token.")
			}
			fmt.Println("🔄 Starting new OAuth flow...")
			tok = StartAuthServer()
		} else {
			fmt.Println("✅ Token refreshed successfully")
			saveToken(tokFile, tok)
		}
	}

	return config.Client(context.Background(), tok)
}

func refreshToken(token *oauth2.Token) (*oauth2.Token, error) {
	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	tokenSource := config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}

	return newToken, nil
}

func isNonInteractiveEnvironment() bool {
	// Check for common cron job indicators
	if os.Getenv("CRON_JOB") == "true" {
		return true
	}

	// Check if running in Docker without TTY
	if os.Getenv("DOCKER_CONTAINER") == "true" {
		return true
	}

	// Check if no display available (common in cron jobs)
	if os.Getenv("DISPLAY") == "" && os.Getenv("SSH_CLIENT") == "" {
		return true
	}

	// Check if stdin is not a terminal
	if !isTerminal() {
		return true
	}

	return false
}

func isTerminal() bool {
	// Simple check for terminal availability
	fileInfo, _ := os.Stdin.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving token to %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to save token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
