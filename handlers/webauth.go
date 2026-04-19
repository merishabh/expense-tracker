package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	webOAuthConfig *oauth2.Config
	sessions       = make(map[string]sessionData)
	sessionsMu     sync.RWMutex
)

type sessionData struct {
	Email     string
	ExpiresAt time.Time
}

func InitWebAuth() {
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	webOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_WEB_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_WEB_CLIENT_SECRET"),
		RedirectURL:  baseURL + "/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	// If already authenticated, redirect to home
	if isAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <title>Expense Tracker - Sign In</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #f5f5f5; display: flex; align-items: center; justify-content: center; min-height: 100vh; }
    .card { background: white; border-radius: 12px; padding: 48px 40px; box-shadow: 0 2px 16px rgba(0,0,0,0.1); text-align: center; max-width: 380px; width: 100%; }
    h1 { font-size: 24px; color: #1a1a1a; margin-bottom: 8px; }
    p { color: #666; margin-bottom: 32px; font-size: 14px; }
    a.btn { display: inline-flex; align-items: center; gap: 12px; background: white; border: 1px solid #ddd; border-radius: 8px; padding: 12px 24px; text-decoration: none; color: #333; font-size: 15px; font-weight: 500; transition: box-shadow 0.2s; }
    a.btn:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.15); }
    svg { width: 20px; height: 20px; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Expense Tracker</h1>
    <p>Sign in to view your personal dashboard</p>
    <a class="btn" href="/auth/login">
      <svg viewBox="0 0 48 48"><path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"/><path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"/><path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"/><path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.18 1.48-4.97 2.31-8.16 2.31-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"/><path fill="none" d="M0 0h48v48H0z"/></svg>
      Sign in with Google
    </a>
  </div>
</body>
</html>`))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	state := generateSessionToken()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   300,
		HttpOnly: true,
		Secure:   os.Getenv("APP_BASE_URL") != "",
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	url := webOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	token, err := webOAuthConfig.Exchange(context.Background(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := webOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	allowedEmail := os.Getenv("ALLOWED_EMAIL")
	if allowedEmail != "" && userInfo.Email != allowedEmail {
		http.Error(w, "Access denied: unauthorized email", http.StatusForbidden)
		return
	}

	sessionToken := generateSessionToken()
	sessionsMu.Lock()
	sessions[sessionToken] = sessionData{
		Email:     userInfo.Email,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	sessionsMu.Unlock()

	secure := os.Getenv("APP_BASE_URL") != ""
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil {
		sessionsMu.Lock()
		delete(sessions, cookie.Value)
		sessionsMu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
	http.Redirect(w, r, "/auth/signin", http.StatusTemporaryRedirect)
}

// webAuthMiddleware redirects unauthenticated requests to the login page.
func webAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/auth/signin", http.StatusTemporaryRedirect)
			return
		}
		next(w, r)
	}
}


// apiAuthMiddleware returns 401 JSON for unauthenticated API requests.
func apiAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next(w, r)
	}
}

func isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if err != nil {
		return false
	}
	sessionsMu.RLock()
	data, ok := sessions[cookie.Value]
	sessionsMu.RUnlock()
	if !ok || time.Now().After(data.ExpiresAt) {
		if ok {
			sessionsMu.Lock()
			delete(sessions, cookie.Value)
			sessionsMu.Unlock()
		}
		return false
	}
	return true
}

func generateSessionToken() string {
	b := make([]byte, 24)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
