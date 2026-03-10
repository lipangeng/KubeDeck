package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestKernelHandlerSlots(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/meta/slots", nil)
	rec := httptest.NewRecorder()

	NewKernelHandler().Slots(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if want := "slot.workloads.summary.insights"; !contains(rec.Body.String(), want) {
		t.Fatalf("expected body to contain %q, got %s", want, rec.Body.String())
	}
}

func TestKernelHandlerSnapshot(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/meta/kernel", nil)
	rec := httptest.NewRecorder()

	NewKernelHandler().Snapshot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if want := "pages"; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := "actions"; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := "slots"; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := "operations"; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
}

func TestKernelHandlerWorkloads(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/workflows/workloads/items?workflowDomainId=workloads&cluster=dev",
		nil,
	)
	rec := httptest.NewRecorder()

	NewKernelHandler().Workloads(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if want := "workload-api-dev"; !contains(rec.Body.String(), want) {
		t.Fatalf("expected body to contain %q, got %s", want, rec.Body.String())
	}
}

func TestKernelHandlerExecuteAction(t *testing.T) {
	body := `{"actionId":"apply","workflowDomainId":"workloads","target":{"cluster":"default","namespace":"default","scope":"namespace"},"input":{"name":"api"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/actions/execute", strings.NewReader(body))
	rec := httptest.NewRecorder()

	NewKernelHandler().ExecuteAction(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if want := "apply accepted"; !contains(rec.Body.String(), want) {
		t.Fatalf("expected body to contain %q, got %s", want, rec.Body.String())
	}
}

func TestKernelHandlerSnapshotIncludesDiscoveredPluginContributions(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, "ops-console")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir plugin dir: %v", err)
	}

	manifest := `{
  "pluginId": "plugin.ops-console",
  "version": "1.0.0",
  "displayName": "Operations Console",
  "contributions": {
    "pages": [
      {
        "id": "page.ops-console",
        "workflowDomainId": "ops-console",
        "route": "/ops-console",
        "entryKey": "ops-console",
        "title": { "key": "opsConsole.title", "fallback": "Operations Console" }
      }
    ],
    "menus": [],
    "actions": [],
    "slots": []
  }
}`
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/meta/kernel", nil)

	NewKernelHandlerWithPluginRoot(root).Snapshot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if want := "page.ops-console"; !contains(rec.Body.String(), want) {
		t.Fatalf("expected body to contain %q, got %s", want, rec.Body.String())
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
