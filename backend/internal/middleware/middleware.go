package middleware

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// RateLimitConfig configures rate limiting
type RateLimitConfig struct {
	Requests  int           // Max requests per window
	Window    time.Duration // Time window
	BurstSize int           // Burst allowance
}

// rateLimiter implements token bucket rate limiting
type rateLimiter struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	lastRefill time.Time
	config     RateLimitConfig
}

// RateLimiter manages rate limiting for multiple clients
type RateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*rateLimiter
	config  RateLimitConfig
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	if config.Requests <= 0 {
		config.Requests = 100
	}
	if config.Window <= 0 {
		config.Window = time.Minute
	}
	if config.BurstSize <= 0 {
		config.BurstSize = config.Requests / 10
	}

	return &RateLimiter{
		clients: make(map[string]*rateLimiter),
		config:  config,
	}
}

// getLimiter returns or creates a rate limiter for a client
func (rl *RateLimiter) getLimiter(clientID string) *rateLimiter {
	rl.mu.RLock()
	limiter, exists := rl.clients[clientID]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists = rl.clients[clientID]; exists {
		return limiter
	}

	limiter = &rateLimiter{
		tokens:     rl.config.BurstSize,
		maxTokens:  rl.config.BurstSize,
		lastRefill: time.Now(),
		config:     rl.config,
	}
	rl.clients[clientID] = limiter
	return limiter
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	limiter := rl.getLimiter(clientID)

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(limiter.lastRefill)
	tokensToAdd := int(elapsed.Seconds() * float64(limiter.config.Requests) / limiter.config.Window.Seconds())

	if tokensToAdd > 0 {
		limiter.tokens = min(limiter.maxTokens, limiter.tokens+tokensToAdd)
		limiter.lastRefill = now
	}

	if limiter.tokens > 0 {
		limiter.tokens--
		return true
	}

	return false
}

// RateLimitMiddleware creates HTTP middleware for rate limiting
func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client identifier (IP address)
		clientID := getClientIP(r)

		if !rl.Allow(clientID) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Validation patterns
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)
	nameRegex     = regexp.MustCompile(`^[a-zA-Z\s'-]{2,100}$`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult holds validation results
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// Validator provides input validation
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateEmail validates an email address
func (v *Validator) ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ValidateUsername validates a username
func (v *Validator) ValidateUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

// ValidateName validates a name
func (v *Validator) ValidateName(name string) bool {
	return nameRegex.MatchString(name)
}

// ValidateRequired checks if a string is not empty
func (v *Validator) ValidateRequired(value string, fieldName string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " is required",
		}
	}
	return nil
}

// ValidateLength checks string length
func (v *Validator) ValidateLength(value string, min, max int, fieldName string) *ValidationError {
	length := len(value)
	if length < min {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " must be at least " + string(rune(min)) + " characters",
		}
	}
	if length > max {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " must be at most " + string(rune(max)) + " characters",
		}
	}
	return nil
}

// SanitizeString removes potentially dangerous characters
func (v *Validator) SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Limit length
	if len(input) > 10000 {
		input = input[:10000]
	}

	return input
}

// SanitizeHTML removes HTML tags
func (v *Validator) SanitizeHTML(input string) string {
	// Simple HTML tag removal
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	return tagRegex.ReplaceAllString(input, "")
}

// ValidateJSONSize checks if JSON payload is within limits
func (v *Validator) ValidateJSONSize(data []byte, maxSize int) *ValidationError {
	if len(data) > maxSize {
		return &ValidationError{
			Field:   "body",
			Message: "Request body too large",
		}
	}
	return nil
}

// MaxBytesMiddleware limits request body size
func MaxBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// ContentTypeMiddleware enforces Content-Type header
func ContentTypeMiddleware(allowedTypes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for GET/HEAD/DELETE requests
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodDelete {
				next.ServeHTTP(w, r)
				return
			}

			contentType := r.Header.Get("Content-Type")
			for _, allowed := range allowedTypes {
				if strings.HasPrefix(contentType, allowed) {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnsupportedMediaType)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "unsupported_media_type",
				"message": "Content-Type must be one of: " + strings.Join(allowedTypes, ", "),
			})
		})
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS filter
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// CORSConfig configures CORS
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

// CORSMiddleware handles CORS
func CORSMiddleware(config CORSConfig) func(http.Handler) http.Handler {
	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = []string{"*"}
	}
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}
	}
	if config.MaxAge == 0 {
		config.MaxAge = 86400
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			origin := r.Header.Get("Origin")
			allowed := false
			for _, o := range config.AllowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
				w.Header().Set("Access-Control-Max-Age", string(rune(config.MaxAge)))

				// Handle preflight
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
