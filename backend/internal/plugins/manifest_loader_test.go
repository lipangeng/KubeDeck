package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifestProvidersFromDirBuildsCapabilityDescriptors(t *testing.T) {
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
    "menus": [
      {
        "id": "menu.ops-console",
        "workflowDomainId": "ops-console",
        "entryKey": "ops-console",
        "route": "/ops-console",
        "placement": "primary",
        "order": 90,
        "visible": true,
        "title": { "key": "opsConsole.title", "fallback": "Operations Console" }
      }
    ],
    "actions": [
      {
        "id": "refresh-ops-console",
        "workflowDomainId": "ops-console",
        "surface": "inline",
        "visible": true,
        "title": { "key": "opsConsole.actions.refresh", "fallback": "Refresh Operations Console" }
      }
    ],
    "slots": [
      {
        "id": "slot.ops-console.summary",
        "workflowDomainId": "ops-console",
        "slotId": "ops-console.summary",
        "placement": "summary",
        "visible": true,
        "title": { "key": "opsConsole.slots.summary", "fallback": "Operations Summary" }
      }
    ],
    "resourcePageExtensions": [
      {
        "kind": "Service",
        "capabilityType": "tab",
        "tabId": "endpoints",
        "title": { "key": "opsConsole.resource.endpoints", "fallback": "Endpoints" },
        "contentFallback": "Service endpoints from manifest"
      },
      {
        "kind": "StatefulSet",
        "capabilityType": "page-takeover",
        "tabId": "statefulset.takeover",
        "priority": 60,
        "title": { "key": "opsConsole.resource.statefulset", "fallback": "StatefulSet takeover" },
        "contentFallback": "Manifest StatefulSet takeover"
      },
      {
        "kind": "Deployment",
        "capabilityType": "action",
        "actionId": "restart-rollout",
        "priority": 40,
        "title": { "key": "opsConsole.resource.restart", "fallback": "Restart Rollout" }
      }
    ]
  }
}`
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	providers, err := LoadManifestProvidersFromDir(root)
	if err != nil {
		t.Fatalf("load manifest providers: %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}

	descriptor := providers[0].CapabilityDescriptor()
	if descriptor.ID != "plugin.ops-console" {
		t.Fatalf("expected plugin id plugin.ops-console, got %q", descriptor.ID)
	}
	if len(descriptor.Pages) != 1 || descriptor.Pages[0].Route != "/ops-console" {
		t.Fatalf("expected one ops-console page contribution, got %+v", descriptor.Pages)
	}
	if len(descriptor.Menus) != 1 || descriptor.Menus[0].Placement != "primary" {
		t.Fatalf("expected one primary menu contribution, got %+v", descriptor.Menus)
	}
	if len(descriptor.Actions) != 1 || descriptor.Actions[0].Surface != "inline" {
		t.Fatalf("expected one inline action contribution, got %+v", descriptor.Actions)
	}
	if len(descriptor.Slots) != 1 || descriptor.Slots[0].Placement != "summary" {
		t.Fatalf("expected one summary slot contribution, got %+v", descriptor.Slots)
	}
    if len(descriptor.ResourcePageExtensions) != 3 {
        t.Fatalf("expected three resource page extensions, got %+v", descriptor.ResourcePageExtensions)
    }
    if descriptor.ResourcePageExtensions[0].TabID != "endpoints" {
        t.Fatalf("expected endpoints tab extension first, got %+v", descriptor.ResourcePageExtensions)
    }
    takeover := descriptor.ResourcePageExtensions[1]
    if takeover.CapabilityType != "page-takeover" || takeover.Priority != 60 {
        t.Fatalf("expected takeover extension with priority 60, got %+v", takeover)
    }
    action := descriptor.ResourcePageExtensions[2]
    if action.CapabilityType != "action" || action.ActionID != "restart-rollout" {
        t.Fatalf("expected action extension restart-rollout, got %+v", action)
    }
}
