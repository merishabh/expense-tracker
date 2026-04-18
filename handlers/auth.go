package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var (
	Config *oauth2.Config
	state  = "randomstatestring"
)

func GetClient() (*http.Client, error) {
	tokFile := gmailTokenPath()
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		if shouldUseInteractiveAuth() {
			log.Printf("No valid token found, starting OAuth flow...")
			tok = StartAuthServer()
		} else {
			return nil, fmt.Errorf("gmail token not available at %s: %w", tokFile, err)
		}
	}

	if tok.RefreshToken != "" {
		tokenSource := Config.TokenSource(context.Background(), tok)
		newTok, err := tokenSource.Token()
		if err != nil {
			log.Printf("Token refresh failed: %v", err)
			if shouldUseInteractiveAuth() {
				log.Printf("Token expired or revoked. Starting new OAuth flow...")
				os.Remove(tokFile)
				tok = StartAuthServer()
			} else {
				return nil, fmt.Errorf("failed to refresh gmail token: %w", err)
			}
		} else {
			if newTok.AccessToken != tok.AccessToken {
				saveToken(tokFile, newTok)
				tok = newTok
			}
		}
	} else {
		if shouldUseInteractiveAuth() {
			log.Printf("No refresh token available. Starting OAuth flow...")
			os.Remove(tokFile)
			tok = StartAuthServer()
		} else {
			return nil, errors.New("gmail token has no refresh token; interactive auth required")
		}
	}

	client := Config.Client(context.Background(), tok)
	return client, nil
}

func InitGmailService() (*gmail.Service, error) {
	clientSecretPath := gmailClientSecretPath()

	b, err := os.ReadFile(clientSecretPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read gmail client secret %s: %w", clientSecretPath, err)
	}

	Config, err = google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse gmail client secret: %w", err)
	}

	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	srv, err := gmail.New(client)
	if err != nil {
		return nil, fmt.Errorf("unable to create gmail client: %w", err)
	}

	return srv, nil
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

func gmailClientSecretPath() string {
	if path := os.Getenv("GMAIL_CLIENT_SECRET_PATH"); path != "" {
		return path
	}
	return "credentials/client_secret.json"
}

func gmailTokenPath() string {
	if path := os.Getenv("GMAIL_TOKEN_PATH"); path != "" {
		return path
	}
	return "credentials/token.json"
}

func shouldUseInteractiveAuth() bool {
	if strings.EqualFold(os.Getenv("DISABLE_INTERACTIVE_AUTH"), "true") {
		return false
	}
	if env := strings.ToLower(os.Getenv("ENVIRONMENT")); env == "production" || env == "prod" {
		return false
	}
	return true
}
