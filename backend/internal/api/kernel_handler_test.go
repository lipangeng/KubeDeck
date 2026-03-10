package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKernelHandlerMenus(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/meta/menus", nil)
	rec := httptest.NewRecorder()

	NewKernelHandler().Menus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if body == "" {
		t.Fatalf("expected menu response body")
	}
	if want := "menu.homepage"; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
}

func TestKernelHandlerActions(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/meta/actions", nil)
	rec := httptest.NewRecorder()

	NewKernelHandler().Actions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if want := "apply"; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
}

func contains(body string, want string) bool {
	return len(body) >= len(want) && (body == want || len(body) > len(want) && (index(body, want) >= 0))
}

func index(s string, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
