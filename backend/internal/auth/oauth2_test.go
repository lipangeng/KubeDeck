package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestNewOAuth2Provider(t *testing.T) {
	cfg := &OAuth2Config{
		Provider:     "github",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/api/auth/callback",
		Scopes:       []string{"user:email"},
	}

	provider, err := NewOAuth2Provider(cfg)
	if err != nil {
		t.Fatalf("Failed to create OAuth2 provider: %v", err)
	}

	if provider.config.Provider != "github" {
		t.Errorf("Expected provider 'github', got '%s'", provider.config.Provider)
	}
}

func TestOAuth2Provider_AuthCodeURL(t *testing.T) {
	cfg := &OAuth2Config{
		Provider:     "github",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"user:email"},
	}

	provider, _ := NewOAuth2Provider(cfg)
	url := provider.AuthCodeURL("test-state")

	if url == "" {
		t.Error("Expected non-empty auth code URL")
	}
}

func TestJWTManager(t *testing.T) {
	manager := NewJWTManager("test-secret-key", "kubedeck")

	userID := uuid.New()
	token, err := manager.GenerateToken(userID, "testuser", "test@example.com", []string{"admin"})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Expected non-empty token")
	}

	claims, err := manager.VerifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if claims.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", claims.Username)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", claims.Email)
	}

	if len(claims.Groups) != 1 || claims.Groups[0] != "admin" {
		t.Errorf("Expected groups ['admin'], got %v", claims.Groups)
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key", "kubedeck")

	userID := uuid.New()
	// Create a token that expires immediately
	claims := Claims{
		UserID:   userID,
		Username: "testuser",
		Email:    "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			Issuer:    "kubedeck",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret-key"))

	_, err := manager.VerifyToken(tokenString)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key", "kubedeck")

	_, err := manager.VerifyToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}
