package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOAuthConfigEndpoint_DefaultModeReady(t *testing.T) {
	t.Setenv("KUBEDECK_OAUTH_MODE", "")
	t.Setenv("KUBEDECK_OAUTH_PROVIDER", "")
	t.Setenv("KUBEDECK_OIDC_ISSUER", "")
	t.Setenv("KUBEDECK_OIDC_CLIENT_ID", "")
	t.Setenv("KUBEDECK_OIDC_CLIENT_SECRET", "")
	t.Setenv("KUBEDECK_OIDC_REDIRECT_URL", "")

	router := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/config", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var body struct {
		Mode     string   `json:"mode"`
		Provider string   `json:"provider"`
		Ready    bool     `json:"ready"`
		Missing  []string `json:"missing"`
		OIDC     struct {
			IssuerExists       bool `json:"issuer_exists"`
			ClientIDExists     bool `json:"client_id_exists"`
			ClientSecretExists bool `json:"client_secret_exists"`
			RedirectURLExists  bool `json:"redirect_url_exists"`
		} `json:"oidc"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON body, got error: %v", err)
	}

	if body.Mode != "stub" {
		t.Fatalf("expected mode stub, got %q", body.Mode)
	}
	if body.Provider != "oauth" {
		t.Fatalf("expected provider oauth, got %q", body.Provider)
	}
	if !body.Ready {
		t.Fatalf("expected ready true, got false")
	}
	if len(body.Missing) != 0 {
		t.Fatalf("expected no missing fields, got %v", body.Missing)
	}
	if body.OIDC.IssuerExists || body.OIDC.ClientIDExists || body.OIDC.ClientSecretExists || body.OIDC.RedirectURLExists {
		t.Fatalf("expected all oidc flags false, got %+v", body.OIDC)
	}
}

func TestOAuthConfigEndpoint_OIDCModeReportsMissingFields(t *testing.T) {
	t.Setenv("KUBEDECK_OAUTH_MODE", "oidc")
	t.Setenv("KUBEDECK_OAUTH_PROVIDER", "corp-sso")
	t.Setenv("KUBEDECK_OIDC_ISSUER", "https://issuer.example")
	t.Setenv("KUBEDECK_OIDC_CLIENT_ID", "client-id")
	t.Setenv("KUBEDECK_OIDC_CLIENT_SECRET", "")
	t.Setenv("KUBEDECK_OIDC_REDIRECT_URL", "")

	router := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/config", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var body struct {
		Mode     string   `json:"mode"`
		Provider string   `json:"provider"`
		Ready    bool     `json:"ready"`
		Missing  []string `json:"missing"`
		OIDC     struct {
			IssuerExists       bool `json:"issuer_exists"`
			ClientIDExists     bool `json:"client_id_exists"`
			ClientSecretExists bool `json:"client_secret_exists"`
			RedirectURLExists  bool `json:"redirect_url_exists"`
		} `json:"oidc"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON body, got error: %v", err)
	}

	if body.Mode != "oidc" {
		t.Fatalf("expected mode oidc, got %q", body.Mode)
	}
	if body.Provider != "oauth" {
		t.Fatalf("expected provider oauth fallback, got %q", body.Provider)
	}
	if body.Ready {
		t.Fatalf("expected ready false when oidc required fields missing")
	}
	if len(body.Missing) != 2 {
		t.Fatalf("expected 2 missing fields, got %v", body.Missing)
	}

	missing := map[string]bool{}
	for _, field := range body.Missing {
		missing[field] = true
	}
	if !missing["KUBEDECK_OIDC_CLIENT_SECRET"] || !missing["KUBEDECK_OIDC_REDIRECT_URL"] {
		t.Fatalf("expected missing client secret + redirect url, got %v", body.Missing)
	}
	if !body.OIDC.IssuerExists || !body.OIDC.ClientIDExists || body.OIDC.ClientSecretExists || body.OIDC.RedirectURLExists {
		t.Fatalf("unexpected oidc existence flags: %+v", body.OIDC)
	}
}
