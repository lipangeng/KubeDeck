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
		})
	}
}
