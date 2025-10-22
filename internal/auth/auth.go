package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// UserInfo represents authenticated user information
type UserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Sub   string `json:"sub"`
}

type contextKey string

const (
	userContextKey contextKey = "user"
	sessionName    string     = "outalator-session"
)

// Config holds OIDC configuration
type Config struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	SessionKey   string
}

// Authenticator handles OIDC authentication
type Authenticator struct {
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
	oauth2Config oauth2.Config
	store        *sessions.CookieStore
}

// NewAuthenticator creates a new OIDC authenticator
func NewAuthenticator(cfg Config) (*Authenticator, error) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})

	// Use provided session key or generate one
	sessionKey := cfg.SessionKey
	if sessionKey == "" {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate session key: %w", err)
		}
		sessionKey = base64.StdEncoding.EncodeToString(key)
	}

	store := sessions.NewCookieStore([]byte(sessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	return &Authenticator{
		provider:     provider,
		verifier:     verifier,
		oauth2Config: oauth2Config,
		store:        store,
	}, nil
}

// LoginHandler initiates the OAuth2 login flow
func (a *Authenticator) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := generateRandomState()

		session, _ := a.store.Get(r, sessionName)
		session.Values["state"] = state
		session.Save(r, w)

		http.Redirect(w, r, a.oauth2Config.AuthCodeURL(state), http.StatusFound)
	}
}

// CallbackHandler handles the OAuth2 callback
func (a *Authenticator) CallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := a.store.Get(r, sessionName)
		if err != nil {
			http.Error(w, "Failed to get session", http.StatusInternalServerError)
			return
		}

		// Verify state
		savedState, ok := session.Values["state"].(string)
		if !ok || savedState != r.URL.Query().Get("state") {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}

		// Exchange code for token
		oauth2Token, err := a.oauth2Config.Exchange(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			return
		}

		// Extract ID token
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "No id_token field in oauth2 token", http.StatusInternalServerError)
			return
		}

		// Verify ID token
		idToken, err := a.verifier.Verify(r.Context(), rawIDToken)
		if err != nil {
			http.Error(w, "Failed to verify ID token", http.StatusInternalServerError)
			return
		}

		// Extract user info
		var userInfo UserInfo
		if err := idToken.Claims(&userInfo); err != nil {
			http.Error(w, "Failed to parse claims", http.StatusInternalServerError)
			return
		}

		// Store user info in session
		session.Values["user"] = userInfo
		session.Values["id_token"] = rawIDToken
		if err := session.Save(r, w); err != nil {
			http.Error(w, "Failed to save session", http.StatusInternalServerError)
			return
		}

		// Redirect to home
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// LogoutHandler handles user logout
func (a *Authenticator) LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := a.store.Get(r, sessionName)
		session.Options.MaxAge = -1
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// Middleware enforces authentication on routes
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login/callback/health endpoints
		if r.URL.Path == "/auth/login" || r.URL.Path == "/auth/callback" || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		session, err := a.store.Get(r, sessionName)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userInfoRaw, ok := session.Values["user"]
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Convert stored user info
		var userInfo UserInfo
		userInfoBytes, err := json.Marshal(userInfoRaw)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(userInfoBytes, &userInfo); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), userContextKey, &userInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext extracts user info from request context
func GetUserFromContext(ctx context.Context) (*UserInfo, error) {
	user, ok := ctx.Value(userContextKey).(*UserInfo)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}
	return user, nil
}

// generateRandomState generates a random state parameter
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
