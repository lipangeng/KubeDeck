package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type registryResponse struct {
	Cluster       string `json:"cluster"`
	ResourceTypes []struct {
		ID               string `json:"id"`
		Group            string `json:"group"`
		Version          string `json:"version"`
		Kind             string `json:"kind"`
		Plural           string `json:"plural"`
		Namespaced       bool   `json:"namespaced"`
		PreferredVersion string `json:"preferredVersion"`
		Source           string `json:"source"`
	} `json:"resourceTypes"`
}

type menusResponse struct {
	Cluster string `json:"cluster"`
	Menus   []struct {
		ID         string `json:"id"`
		Group      string `json:"group"`
		Title      string `json:"title"`
		TargetType string `json:"targetType"`
		TargetRef  string `json:"targetRef"`
		Source     string `json:"source"`
		Order      int    `json:"order"`
		Visible    bool   `json:"visible"`
	} `json:"menus"`
}

func TestRegistryEndpoint(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/meta/registry?cluster=dev", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var body registryResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}

	if body.Cluster != "dev" {
		t.Fatalf("expected cluster dev, got %q", body.Cluster)
	}
	if len(body.ResourceTypes) == 0 {
		t.Fatalf("expected non-empty resourceTypes, body=%s", resp.Body.String())
	}
	if body.ResourceTypes[0].ID == "" || body.ResourceTypes[0].Kind == "" {
		t.Fatalf("expected typed resource fields, body=%s", resp.Body.String())
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

	var body menusResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}

	if body.Cluster != "dev" {
		t.Fatalf("expected cluster dev, got %q", body.Cluster)
	}
	if len(body.Menus) == 0 {
		t.Fatalf("expected non-empty menus, body=%s", resp.Body.String())
	}
	first := body.Menus[0]
	if first.ID == "" || first.TargetType == "" || first.TargetRef == "" {
		t.Fatalf("expected typed menu fields, body=%s", resp.Body.String())
	}
}

func TestResourceApplyEndpoint(t *testing.T) {
	router := NewRouter()

	body := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-ok
---
apiVersion: v1
kind: Service
metadata:
  name: svc-fail
---
kind: Broken
`
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/resources/apply?cluster=dev&defaultNs=kube-system",
		strings.NewReader(body),
	)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var payload struct {
		Status    string `json:"status"`
		Cluster   string `json:"cluster"`
		DefaultNS string `json:"defaultNamespace"`
		Total     int    `json:"total"`
		Succeeded int    `json:"succeeded"`
		Failed    int    `json:"failed"`
		Results   []struct {
			Index     int    `json:"index"`
			Kind      string `json:"kind"`
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Status    string `json:"status"`
			Reason    string `json:"reason"`
		} `json:"results"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}

	if payload.Status != "partial" {
		t.Fatalf("expected status partial, got %q body=%s", payload.Status, resp.Body.String())
	}
	if payload.Cluster != "dev" {
		t.Fatalf("expected cluster dev, got %q", payload.Cluster)
	}
	if payload.DefaultNS != "kube-system" {
		t.Fatalf("expected defaultNamespace kube-system, got %q", payload.DefaultNS)
	}
	if payload.Total != 3 || payload.Succeeded != 1 || payload.Failed != 2 {
		t.Fatalf("unexpected counters total=%d succeeded=%d failed=%d", payload.Total, payload.Succeeded, payload.Failed)
	}
	if len(payload.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(payload.Results))
	}
	if payload.Results[0].Namespace != "kube-system" {
		t.Fatalf("expected default namespace injected, got %q", payload.Results[0].Namespace)
	}
	if payload.Results[1].Status != "failed" {
		t.Fatalf("expected second result failed, got %q", payload.Results[1].Status)
	}
	if payload.Results[2].Status != "failed" {
		t.Fatalf("expected third result failed, got %q", payload.Results[2].Status)
	}
}

func TestHealthzEndpoint(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if resp.Body.String() != "ok" {
		t.Fatalf("expected body ok, got %q", resp.Body.String())
	}
}

func TestReadyzEndpoint(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/readyz", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if resp.Body.String() != "ok" {
		t.Fatalf("expected body ok, got %q", resp.Body.String())
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
