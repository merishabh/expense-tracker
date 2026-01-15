package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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

	// Create client that will automatically refresh the token
	client := Config.Client(context.Background(), tok)

	// Check if we need to save the refreshed token
	if tok.RefreshToken != "" {
		// Create a token source that will automatically handle refresh
		tokenSource := Config.TokenSource(context.Background(), tok)
		newTok, err := tokenSource.Token()
		if err == nil && newTok.AccessToken != tok.AccessToken {
			// Token was refreshed, save it
			saveToken(tokFile, newTok)
		}
	}

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

	http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		fmt.Fprintln(w, "Authorization successful! You can close this window.")
		codeCh <- code
	})

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	authURL := Config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
	fmt.Printf("Open this URL in your browser:\n%v\n", authURL)

	code := <-codeCh
	srv.Shutdown(context.Background())

	token, err := Config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to exchange code for token: %v", err)
	}

	saveToken("credentials/token.json", token)
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
