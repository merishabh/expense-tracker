package handlers

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
	Config *oauth2.Config
	state  = "randomstatestring"
)

func GetClient() *http.Client {
	tokFile := "credentials/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		// If token is not found, start OAuth flow interactively
		log.Printf("No valid token found, starting OAuth flow...")
		tok = StartAuthServer()
	}

	// Try to refresh the token if it exists
	if tok.RefreshToken != "" {
		// Create a token source that will automatically handle refresh
		tokenSource := Config.TokenSource(context.Background(), tok)
		newTok, err := tokenSource.Token()
		if err != nil {
			// Token refresh failed - token is expired or revoked
			log.Printf("Token refresh failed: %v", err)
			log.Printf("Token expired or revoked. Starting new OAuth flow...")
			// Delete the old token file
			os.Remove(tokFile)
			// Start new OAuth flow
			tok = StartAuthServer()
		} else {
			// Token was successfully refreshed
			if newTok.AccessToken != tok.AccessToken {
				// Token was refreshed, save it
				saveToken(tokFile, newTok)
				tok = newTok
			}
		}
	} else {
		// No refresh token available, need to re-authenticate
		log.Printf("No refresh token available. Starting OAuth flow...")
		os.Remove(tokFile)
		tok = StartAuthServer()
	}

	// Create client that will automatically refresh the token
	client := Config.Client(context.Background(), tok)
	return client
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

func StartAuthServer() *oauth2.Token {
	codeCh := make(chan string)
	srv := &http.Server{Addr: ":8080"}

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Authorization code not found", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "Authorization successful! You can close this window.")
		codeCh <- code
	})
	srv.Handler = mux

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("OAuth callback server error: %v", err)
		}
	}()

	authURL := Config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
	fmt.Printf("\n=== OAuth Authentication Required ===\n")
	fmt.Printf("Open this URL in your browser:\n%v\n\n", authURL)
	fmt.Printf("Waiting for authorization...\n")

	code := <-codeCh

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down OAuth server: %v", err)
	}

	token, err := Config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to exchange code for token: %v", err)
	}

	saveToken("credentials/token.json", token)
	fmt.Printf("✓ Token saved successfully!\n\n")
	return token
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
