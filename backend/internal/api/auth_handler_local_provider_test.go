package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthLoginDeniedWhenLocalProviderDisabledByDefaultInProduction(t *testing.T) {
	t.Setenv("KUBEDECK_ENV", "production")
	t.Setenv("KUBEDECK_LOCAL_AUTH_ENABLED", "")

	h := NewAuthHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)))
	resp := httptest.NewRecorder()

	h.Login(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, resp.Code, resp.Body.String())
	}
}
