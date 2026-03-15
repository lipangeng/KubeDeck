package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type manifestShape struct {
	PluginID      string          `json:"pluginId"`
	Version       string          `json:"version"`
	DisplayName   string          `json:"displayName"`
	Contributions json.RawMessage `json:"contributions"`
}

type contributionShape struct {
	Pages   []json.RawMessage `json:"pages"`
	Menus   []json.RawMessage `json:"menus"`
	Actions []json.RawMessage `json:"actions"`
	Slots   []json.RawMessage `json:"slots"`
	ResourcePageExtensions []json.RawMessage `json:"resourcePageExtensions"`

	Extensions json.RawMessage `json:"extensions"`
	Resources  json.RawMessage `json:"resources"`
}

type resourcePageExtensionShape struct {
	CapabilityType string `json:"capabilityType"`
	ActionID       string `json:"actionId"`
}

func TestTemplateManifestShape(t *testing.T) {
	t.Helper()

	templates := []string{
		filepath.Join("..", "..", "..", "plugins", "templates", "frontend-plugin-template", "plugin.manifest.json"),
		filepath.Join("..", "..", "..", "plugins", "templates", "backend-plugin-template", "plugin.manifest.json"),
	}

	for _, templatePath := range templates {
		t.Run(templatePath, func(t *testing.T) {
			content, err := os.ReadFile(templatePath)
			if err != nil {
				t.Fatalf("read manifest: %v", err)
			}

			var m manifestShape
			if err := json.Unmarshal(content, &m); err != nil {
				t.Fatalf("unmarshal manifest: %v", err)
			}

			if m.PluginID == "" {
				t.Fatal("pluginId is required")
			}
			if m.Version == "" {
				t.Fatal("version is required")
			}
			if m.DisplayName == "" {
				t.Fatal("displayName is required")
			}
			if len(m.Contributions) == 0 || string(m.Contributions) == "null" {
				t.Fatal("contributions is required")
			}

			var contributions contributionShape
			if err := json.Unmarshal(m.Contributions, &contributions); err != nil {
				t.Fatalf("unmarshal contributions: %v", err)
			}

			if contributions.Pages == nil {
				t.Fatal("contributions.pages is required")
			}
			if contributions.Menus == nil {
				t.Fatal("contributions.menus is required")
			}
			if contributions.Actions == nil {
				t.Fatal("contributions.actions is required")
			}
			if contributions.Slots == nil {
				t.Fatal("contributions.slots is required")
			}
			if contributions.ResourcePageExtensions == nil {
				t.Fatal("contributions.resourcePageExtensions is required")
			}
			hasActionExtension := false
			for _, rawExtension := range contributions.ResourcePageExtensions {
				var extension resourcePageExtensionShape
				if err := json.Unmarshal(rawExtension, &extension); err != nil {
					t.Fatalf("unmarshal resourcePageExtension: %v", err)
				}
				if extension.CapabilityType == "action" && extension.ActionID != "" {
					hasActionExtension = true
				}
			}
			if !hasActionExtension {
				t.Fatal("contributions.resourcePageExtensions must include an action example")
			}
			if len(contributions.Extensions) > 0 {
				t.Fatal("contributions.extensions must not be used in the kernel template")
			}
			if len(contributions.Resources) > 0 {
				t.Fatal("contributions.resources must not be used in the kernel template")
			}
		})
	}
}
