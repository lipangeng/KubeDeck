package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	// UserContextKey is the context key for storing user claims
	UserContextKey contextKey = "user"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public endpoints
			if isPublicEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			token := extractToken(r)
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := jwtManager.VerifyToken(token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken extracts JWT token from request
func extractToken(r *http.Request) string {
	// Try cookie first
	cookie, err := r.Cookie("auth_token")
	if err == nil {
		return cookie.Value
	}

	// Try Authorization header
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// isPublicEndpoint checks if the endpoint is public (no auth required)
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/api/auth/login",
		"/api/auth/callback",
		"/api/auth/logout",
		"/healthz",
		"/api/readyz",
		"/readyz",
	}
	for _, p := range publicPaths {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}

// GetClaimsFromContext extracts user claims from request context
func GetClaimsFromContext(r *http.Request) (*Claims, bool) {
	claims, ok := r.Context().Value(UserContextKey).(*Claims)
	return claims, ok
}
