package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegistryEndpoint(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/meta/registry?cluster=dev", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}

	if _, ok := body["resourceTypes"]; !ok {
		t.Fatalf("expected response to contain resourceTypes key, body=%s", resp.Body.String())
	}
}

func TestClustersEndpoint(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/meta/clusters", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}

	if _, ok := body["clusters"]; !ok {
		t.Fatalf("expected response to contain clusters key, body=%s", resp.Body.String())
	}
}

func TestMenusEndpoint(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/meta/menus?cluster=dev", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}

	if _, ok := body["menus"]; !ok {
		t.Fatalf("expected response to contain menus key, body=%s", resp.Body.String())
	}
}

func TestResourceApplyEndpoint(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/resources/apply", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}

	if _, ok := body["status"]; !ok {
		t.Fatalf("expected response to contain status key, body=%s", resp.Body.String())
	}

	if _, ok := body["message"]; !ok {
		t.Fatalf("expected response to contain message key, body=%s", resp.Body.String())
	}
}

func TestRegistryMethodNotAllowed(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/meta/registry", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, resp.Code)
	}

	if allow := resp.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, allow)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON error response, got error: %v", err)
	}

	if _, ok := body["error"]; !ok {
		t.Fatalf("expected response to contain error key, body=%s", resp.Body.String())
	}
}
