package api

import (
	"bytes"
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
	if want := "resourcePageExtensions"; !contains(body, want) {
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

func TestKernelHandlerSnapshotLoadsRepositorySamplePluginAndSkipsTemplates(t *testing.T) {
	root := filepath.Join("..", "..", "..", "plugins")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/meta/kernel", nil)

	NewKernelHandlerWithPluginRoot(root).Snapshot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if want := "page.sample-ops-console"; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := `"TabID":"endpoints"`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := `"CapabilityType":"page-takeover"`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := `"Priority":60`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := `"ActionID":"restart-rollout"`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := `"Placement":"summary"`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if unwanted := "example-frontend-plugin"; contains(body, unwanted) {
		t.Fatalf("expected body not to contain template plugin %q, got %s", unwanted, body)
	}
	if unwanted := "example-backend-plugin"; contains(body, unwanted) {
		t.Fatalf("expected body not to contain template plugin %q, got %s", unwanted, body)
	}
}

func TestNewKernelHandlerUsesRepositoryPluginsByDefaultWhenEnvUnset(t *testing.T) {
	original := os.Getenv("KUBEDECK_PLUGIN_DIR")
	if err := os.Unsetenv("KUBEDECK_PLUGIN_DIR"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	t.Cleanup(func() {
		if original == "" {
			_ = os.Unsetenv("KUBEDECK_PLUGIN_DIR")
			return
		}
		_ = os.Setenv("KUBEDECK_PLUGIN_DIR", original)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/meta/kernel", nil)

	NewKernelHandler().Snapshot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if want := "page.sample-ops-console"; !contains(rec.Body.String(), want) {
		t.Fatalf("expected body to contain %q, got %s", want, rec.Body.String())
	}
}

func TestKernelHandlerMenuPreferencesRoundTrip(t *testing.T) {
	handler := NewKernelHandler()
	putBody := `{
  "globalOverrides": [
    {
      "scope": "global",
      "moveEntryKeys": {
        "operations": "core"
      }
    }
  ],
  "clusterOverrides": [
    {
      "scope": "cluster",
      "pinEntryKeys": ["operations"]
    }
  ]
}`

	putReq := httptest.NewRequest(
		http.MethodPut,
		"/api/preferences/menu?cluster=prod-eu1",
		bytes.NewBufferString(putBody),
	)
	putRec := httptest.NewRecorder()
	handler.MenuPreferences(putRec, putReq)

	if putRec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", putRec.Code)
	}

	getReq := httptest.NewRequest(
		http.MethodGet,
		"/api/preferences/menu?cluster=prod-eu1",
		nil,
	)
	getRec := httptest.NewRecorder()
	handler.MenuPreferences(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getRec.Code)
	}
	body := getRec.Body.String()
	if want := `"globalOverrides"`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := `"operations":"core"`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := `"clusterOverrides"`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := `"pinEntryKeys":["operations"]`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
}

func TestKernelHandlerSnapshotAppliesClusterMenuOverrides(t *testing.T) {
	handler := NewKernelHandler()
	putBody := `{
  "clusterOverrides": [
    {
      "scope": "cluster",
      "moveEntryKeys": {
        "operations": "core"
      },
      "pinEntryKeys": ["operations"]
    }
  ]
}`

	putReq := httptest.NewRequest(
		http.MethodPut,
		"/api/preferences/menu?cluster=prod-eu1",
		bytes.NewBufferString(putBody),
	)
	putRec := httptest.NewRecorder()
	handler.MenuPreferences(putRec, putReq)
	if putRec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", putRec.Code)
	}

	snapshotReq := httptest.NewRequest(
		http.MethodGet,
		"/api/meta/kernel?cluster=prod-eu1",
		nil,
	)
	snapshotRec := httptest.NewRecorder()
	handler.Snapshot(snapshotRec, snapshotReq)

	if snapshotRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", snapshotRec.Code)
	}
	body := snapshotRec.Body.String()
	if want := `"menuOverrides"`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := `"groupKey":"core"`; !contains(body, want) {
		t.Fatalf("expected body to contain moved core group entry, got %s", body)
	}
	if want := `"pinned":true`; !contains(body, want) {
		t.Fatalf("expected body to contain pinned entry, got %s", body)
	}
}

func TestKernelHandlerMenuPreferencesRoundTripPreservesScopedOrderingFields(t *testing.T) {
	handler := NewKernelHandler()
	putBody := `{
  "globalOverrides": [
    {
      "scope": "work-global",
      "pinEntryKeys": ["operations"],
      "groupOrderOverrides": ["extensions", "core", "platform", "resources"],
      "itemOrderOverrides": {
        "core": ["operations", "workloads", "homepage"]
      }
    }
  ],
  "clusterOverrides": [
    {
      "scope": "work-cluster",
      "hiddenEntryKeys": ["services"]
    },
    {
      "scope": "cluster",
      "pinEntryKeys": ["menu-settings"]
    }
  ]
}`

	putReq := httptest.NewRequest(
		http.MethodPut,
		"/api/preferences/menu?cluster=prod-eu1",
		bytes.NewBufferString(putBody),
	)
	putRec := httptest.NewRecorder()
	handler.MenuPreferences(putRec, putReq)

	if putRec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", putRec.Code)
	}

	getReq := httptest.NewRequest(
		http.MethodGet,
		"/api/preferences/menu?cluster=prod-eu1",
		nil,
	)
	getRec := httptest.NewRecorder()
	handler.MenuPreferences(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getRec.Code)
	}
	body := getRec.Body.String()
	if want := `"scope":"work-global"`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := `"groupOrderOverrides":["extensions","core","platform","resources"]`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := `"itemOrderOverrides":{"core":["operations","workloads","homepage"]}`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
	if want := `"scope":"cluster"`; !contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
}

func TestKernelHandlerSnapshotSupportsScopedConfigurationMenus(t *testing.T) {
	handler := NewKernelHandler()

	systemReq := httptest.NewRequest(
		http.MethodGet,
		"/api/meta/kernel?cluster=prod-eu1&scope=system",
		nil,
	)
	systemRec := httptest.NewRecorder()
	handler.Snapshot(systemRec, systemReq)

	if systemRec.Code != http.StatusOK {
		t.Fatalf("expected system snapshot status 200, got %d", systemRec.Code)
	}
	systemBody := systemRec.Body.String()
	if want := `"entryKey":"menu-settings"`; !contains(systemBody, want) {
		t.Fatalf("expected system snapshot to contain %q, got %s", want, systemBody)
	}
	if want := `"entryKey":"plugin-settings"`; !contains(systemBody, want) {
		t.Fatalf("expected system snapshot to contain %q, got %s", want, systemBody)
	}

	clusterReq := httptest.NewRequest(
		http.MethodGet,
		"/api/meta/kernel?cluster=prod-eu1&scope=cluster",
		nil,
	)
	clusterRec := httptest.NewRecorder()
	handler.Snapshot(clusterRec, clusterReq)

	if clusterRec.Code != http.StatusOK {
		t.Fatalf("expected cluster snapshot status 200, got %d", clusterRec.Code)
	}
	clusterBody := clusterRec.Body.String()
	if want := `"entryKey":"menu-settings"`; !contains(clusterBody, want) {
		t.Fatalf("expected cluster snapshot to contain %q, got %s", want, clusterBody)
	}
	if want := `"entryKey":"extensions"`; !contains(clusterBody, want) {
		t.Fatalf("expected cluster snapshot to contain %q, got %s", want, clusterBody)
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
