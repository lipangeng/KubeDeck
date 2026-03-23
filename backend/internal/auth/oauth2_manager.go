package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrInvalidState     = errors.New("invalid state parameter")
	ErrProviderNotFound = errors.New("oauth2 provider not found")
)

// OAuth2Config holds OAuth2 provider configuration
type OAuth2Config struct {
	Provider     string   `json:"provider"` // github, google, gitlab, oidc
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`

	// OIDC specific
	IssuerURL string `json:"issuer_url"`

	// Internal
	oauthConfig  *oauth2.Config
	oidcVerifier *oidc.IDTokenVerifier
}

// UserInfo represents authenticated user information
type UserInfo struct {
	ID        string          `json:"id"`
	Email     string          `json:"email"`
	Name      string          `json:"name"`
	Username  string          `json:"username"`
	Provider  string          `json:"provider"`
	Verified  bool            `json:"verified"`
	AvatarURL string          `json:"avatar_url,omitempty"`
	RawData   json.RawMessage `json:"-"`
}

// Claims represents JWT claims
type Claims struct {
	UserID   string   `json:"sub"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Provider string   `json:"provider"`
	Groups   []string `json:"groups,omitempty"`
	jwt.RegisteredClaims
}

// Session represents a user session
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Expiry       time.Time `json:"expiry"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// State stores OAuth2 state parameters
type State struct {
	Value    string    `json:"value"`
	Expiry   time.Time `json:"expiry"`
	Redirect string    `json:"redirect"`
}

// Manager manages OAuth2 providers and sessions
type Manager struct {
	providers map[string]*OAuth2Config
	sessions  map[string]*Session
	states    map[string]*State
	mu        sync.RWMutex
	jwtSecret []byte
	jwtExpiry time.Duration
}

// NewManager creates a new OAuth2 manager
func NewManager(jwtSecret string) *Manager {
	return &Manager{
		providers: make(map[string]*OAuth2Config),
		sessions:  make(map[string]*Session),
		states:    make(map[string]*State),
		jwtSecret: []byte(jwtSecret),
		jwtExpiry: 24 * time.Hour,
	}
}

// RegisterProvider registers an OAuth2 provider
func (m *Manager) RegisterProvider(cfg *OAuth2Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Initialize OAuth2 config
	cfg.oauthConfig = &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
	}

	// Set provider-specific endpoints
	switch cfg.Provider {
	case "github":
		cfg.oauthConfig.Endpoint = oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		}
	case "google":
		cfg.oauthConfig.Endpoint = oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		}
	case "gitlab":
		cfg.oauthConfig.Endpoint = oauth2.Endpoint{
			AuthURL:  "https://gitlab.com/oauth/authorize",
			TokenURL: "https://gitlab.com/oauth/token",
		}
	case "oidc":
		if cfg.IssuerURL == "" {
			return fmt.Errorf("issuer_url required for OIDC provider")
		}
		provider, err := oidc.NewProvider(context.Background(), cfg.IssuerURL)
		if err != nil {
			return fmt.Errorf("failed to initialize OIDC provider: %w", err)
		}
		cfg.oauthConfig.Endpoint = provider.Endpoint()
		cfg.oidcVerifier = provider.Verifier(&oidc.Config{
			ClientID: cfg.ClientID,
		})
	default:
		return fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}

	m.providers[cfg.Provider] = cfg
	return nil
}

// GetProvider returns a provider by name
func (m *Manager) GetProvider(name string) (*OAuth2Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[name]
	if !ok {
		return nil, ErrProviderNotFound
	}
	return provider, nil
}

// GenerateAuthURL generates OAuth2 authorization URL
func (m *Manager) GenerateAuthURL(provider, redirect string) (string, string, error) {
	cfg, err := m.GetProvider(provider)
	if err != nil {
		return "", "", err
	}

	// Generate state parameter
	state := &State{
		Value:    generateRandomString(32),
		Expiry:   time.Now().Add(10 * time.Minute),
		Redirect: redirect,
	}

	m.mu.Lock()
	m.states[state.Value] = state
	m.mu.Unlock()

	// Generate auth URL
	authURL := cfg.oauthConfig.AuthCodeURL(state.Value, oauth2.AccessTypeOffline)
	return authURL, state.Value, nil
}

// ExchangeCode exchanges authorization code for tokens
func (m *Manager) ExchangeCode(ctx context.Context, provider, code string) (*oauth2.Token, *UserInfo, error) {
	cfg, err := m.GetProvider(provider)
	if err != nil {
		return nil, nil, err
	}

	// Exchange code for token
	token, err := cfg.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info
	userInfo, err := m.GetUserInfo(ctx, cfg, token)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return token, userInfo, nil
}

// GetUserInfo retrieves user information from OAuth2 provider
func (m *Manager) GetUserInfo(ctx context.Context, cfg *OAuth2Config, token *oauth2.Token) (*UserInfo, error) {
	client := cfg.oauthConfig.Client(ctx, token)

	switch cfg.Provider {
	case "github":
		return m.getGitHubUserInfo(ctx, client)
	case "google":
		return m.getGoogleUserInfo(ctx, client, token)
	case "gitlab":
		return m.getGitLabUserInfo(ctx, client)
	case "oidc":
		return m.getOIDCUserInfo(ctx, cfg, token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

// GenerateJWT generates a JWT token for a user
func (m *Manager) GenerateJWT(userInfo *UserInfo) (string, error) {
	claims := Claims{
		UserID:   userInfo.ID,
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Provider: userInfo.Provider,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "kubedeck",
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(m.jwtSecret)
}

// VerifyJWT verifies and parses a JWT token
func (m *Manager) VerifyJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// CreateSession creates a new user session
func (m *Manager) CreateSession(userID, refreshToken string) *Session {
	session := &Session{
		ID:           generateRandomString(32),
		UserID:       userID,
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	m.mu.Lock()
	m.sessions[session.ID] = session
	m.mu.Unlock()

	return session
}

// GetSession returns a session by ID
func (m *Manager) GetSession(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[id]
	if !exists {
		return nil, false
	}

	// Check expiry
	if time.Now().After(session.Expiry) {
		return nil, false
	}

	// Update last activity
	session.LastActivity = time.Now()
	return session, true
}

// DeleteSession removes a session
func (m *Manager) DeleteSession(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, id)
}

// ValidateState validates and removes OAuth2 state
func (m *Manager) ValidateState(state string) (*State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stored, ok := m.states[state]
	if !ok {
		return nil, ErrInvalidState
	}

	if time.Now().After(stored.Expiry) {
		delete(m.states, state)
		return nil, ErrInvalidState
	}

	delete(m.states, state)
	return stored, nil
}

// Provider-specific user info getters

func (m *Manager) getGitHubUserInfo(ctx context.Context, client *http.Client) (*UserInfo, error) {
	// Get user profile
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var githubUser struct {
		Login     string `json:"login"`
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, err
	}

	// If email is not public, fetch from emails endpoint
	if githubUser.Email == "" {
		githubUser.Email, _ = m.getGitHubPrimaryEmail(ctx, client)
	}

	return &UserInfo{
		ID:        fmt.Sprintf("github:%d", githubUser.ID),
		Username:  githubUser.Login,
		Email:     githubUser.Email,
		Name:      githubUser.Name,
		Provider:  "github",
		Verified:  githubUser.Email != "",
		AvatarURL: githubUser.AvatarURL,
	}, nil
}

func (m *Manager) getGitHubPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	return "", fmt.Errorf("no verified primary email found")
}

func (m *Manager) getGoogleUserInfo(ctx context.Context, client *http.Client, token *oauth2.Token) (*UserInfo, error) {
	// Parse ID token for Google
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in response")
	}

	verifier := oidc.NewVerifier("https://accounts.google.com", nil, &oidc.Config{
		SkipClientIDCheck: true,
	})

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Sub           string `json:"sub"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:       fmt.Sprintf("google:%s", claims.Sub),
		Email:    claims.Email,
		Name:     claims.Name,
		Provider: "google",
		Verified: claims.EmailVerified,
	}, nil
}

func (m *Manager) getGitLabUserInfo(ctx context.Context, client *http.Client) (*UserInfo, error) {
	resp, err := client.Get("https://gitlab.com/api/v4/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gitlabUser struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&gitlabUser); err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:        fmt.Sprintf("gitlab:%d", gitlabUser.ID),
		Username:  gitlabUser.Username,
		Email:     gitlabUser.Email,
		Name:      gitlabUser.Name,
		Provider:  "gitlab",
		Verified:  gitlabUser.State == "active",
		AvatarURL: gitlabUser.AvatarURL,
	}, nil
}

func (m *Manager) getOIDCUserInfo(ctx context.Context, cfg *OAuth2Config, token *oauth2.Token) (*UserInfo, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in response")
	}

	idToken, err := cfg.oidcVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}

	var claims struct {
		Email             string `json:"email"`
		EmailVerified     bool   `json:"email_verified"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
		Sub               string `json:"sub"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:       fmt.Sprintf("oidc:%s", claims.Sub),
		Email:    claims.Email,
		Name:     claims.Name,
		Username: claims.PreferredUsername,
		Provider: "oidc",
		Verified: claims.EmailVerified,
	}, nil
}

// Utility functions

func generateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:n]
}
