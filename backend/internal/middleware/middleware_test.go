package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		Requests:  10,
		Window:    time.Second,
		BurstSize: 2,
	})

	clientID := "test-client"

	// Should allow burst
	for i := 0; i < 2; i++ {
		if !limiter.Allow(clientID) {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	// Should deny after burst
	if limiter.Allow(clientID) {
		t.Error("Request after burst should be denied")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		Requests:  10,
		Window:    time.Second,
		BurstSize: 2,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := limiter.RateLimitMiddleware(handler)

	// First request should succeed
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("First request should succeed, got %d", rec.Code)
	}

	// Second request from same IP should succeed (burst)
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec = httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Second request should succeed (burst), got %d", rec.Code)
	}

	// Third request from same IP should be rate limited
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec = httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Third request should be rate limited, got %d", rec.Code)
	}

	// Request from different IP should succeed
	// Wait a bit for tokens to refill
	time.Sleep(100 * time.Millisecond)
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec = httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Request from different IP should succeed, got %d", rec.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remote   string
		expected string
	}{
		{
			name: "X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			remote:   "127.0.0.1:12345",
			expected: "192.168.1.1",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.2",
			},
			remote:   "127.0.0.1:12345",
			expected: "192.168.1.2",
		},
		{
			name:     "RemoteAddr",
			headers:  map[string]string{},
			remote:   "192.168.1.3:12345",
			expected: "192.168.1.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			req.RemoteAddr = tt.remote

			ip := getClientIP(req)
			if ip != tt.expected {
				t.Errorf("Expected IP %s, got %s", tt.expected, ip)
			}
		})
	}
}

func TestValidator(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"valid email", "test@example.com", true},
		{"invalid email", "test@", false},
		{"invalid email no domain", "test@example", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidateEmail(tt.email)
			if result != tt.expected {
				t.Errorf("ValidateEmail(%s) = %v, expected %v", tt.email, result, tt.expected)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{"valid username", "test_user", true},
		{"valid username with dash", "test-user", true},
		{"invalid username too short", "ab", false},
		{"invalid username special chars", "test@user", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidateUsername(tt.username)
			if result != tt.expected {
				t.Errorf("ValidateUsername(%s) = %v, expected %v", tt.username, result, tt.expected)
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	v := NewValidator()

	// Empty should fail
	err := v.ValidateRequired("", "test_field")
	if err == nil {
		t.Error("Empty field should fail validation")
	}

	// Whitespace should fail
	err = v.ValidateRequired("   ", "test_field")
	if err == nil {
		t.Error("Whitespace field should fail validation")
	}

	// Non-empty should pass
	err = v.ValidateRequired("value", "test_field")
	if err != nil {
		t.Error("Non-empty field should pass validation")
	}
}

func TestSanitizeString(t *testing.T) {
	v := NewValidator()

	// Should remove null bytes
	input := "test\x00string"
	result := v.SanitizeString(input)
	if strings.Contains(result, "\x00") {
		t.Error("SanitizeString should remove null bytes")
	}

	// Should trim whitespace
	input = "  test  "
	result = v.SanitizeString(input)
	if result != "test" {
		t.Errorf("SanitizeString should trim whitespace, got '%s'", result)
	}

	// Should limit length
	input = strings.Repeat("a", 20000)
	result = v.SanitizeString(input)
	if len(result) > 10000 {
		t.Error("SanitizeString should limit length to 10000")
	}
}

func TestSanitizeHTML(t *testing.T) {
	v := NewValidator()

	input := "<script>alert('xss')</script>Hello"
	result := v.SanitizeHTML(input)
	if strings.Contains(result, "<script>") {
		t.Error("SanitizeHTML should remove HTML tags")
	}
	if !strings.Contains(result, "Hello") {
		t.Error("SanitizeHTML should preserve text content")
	}
}

func TestMaxBytesMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := MaxBytesMiddleware(10)(handler)

	// Request within limit should succeed
	req := httptest.NewRequest("POST", "/test", strings.NewReader("small"))
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Small request should succeed, got %d", rec.Code)
	}
}

func TestContentTypeMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := ContentTypeMiddleware("application/json")(handler)

	// GET request should pass without Content-Type
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET request should pass, got %d", rec.Code)
	}

	// POST with correct Content-Type should pass
	req = httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST with correct Content-Type should pass, got %d", rec.Code)
	}

	// POST with wrong Content-Type should fail
	req = httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "text/plain")
	rec = httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("POST with wrong Content-Type should fail, got %d", rec.Code)
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := SecurityHeadersMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	headers := rec.Header()

	if headers.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Missing X-Content-Type-Options header")
	}

	if headers.Get("X-XSS-Protection") != "1; mode=block" {
		t.Error("Missing X-XSS-Protection header")
	}

	if headers.Get("X-Frame-Options") != "DENY" {
		t.Error("Missing X-Frame-Options header")
	}
}

func TestCORSMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORSMiddleware(CORSConfig{
		AllowedOrigins: []string{"http://example.com"},
		AllowedMethods: []string{"GET", "POST"},
	})(handler)

	// Request with allowed origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Error("Missing or incorrect Access-Control-Allow-Origin header")
	}

	// OPTIONS preflight request
	req = httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rec = httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("OPTIONS request should return 204, got %d", rec.Code)
	}
}
