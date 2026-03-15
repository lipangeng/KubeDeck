package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackendTemplateUsesCapabilityProvider(t *testing.T) {
	content, err := os.ReadFile(
		filepath.Join("..", "..", "..", "plugins", "templates", "backend-plugin-template", "src", "index.go"),
	)
	if err != nil {
		t.Fatalf("read backend template source: %v", err)
	}

	source := string(content)
	if strings.Contains(source, "sdk.Plugin") {
		t.Fatal("backend template must not reference legacy sdk.Plugin")
	}
	if !strings.Contains(source, "sdk.CapabilityProvider") {
		t.Fatal("backend template must return sdk.CapabilityProvider")
	}
}

func TestFrontendTemplateUsesKernelContributions(t *testing.T) {
	content, err := os.ReadFile(
		filepath.Join("..", "..", "..", "plugins", "templates", "frontend-plugin-template", "src", "index.ts"),
	)
	if err != nil {
		t.Fatalf("read frontend template source: %v", err)
	}

	source := string(content)
	if strings.Contains(source, "registerExtensions") {
		t.Fatal("frontend template must not reference legacy registerExtensions")
	}
	if !strings.Contains(source, "registerSlots") {
		t.Fatal("frontend template must expose registerSlots")
	}
	if !strings.Contains(source, "registerResourcePageExtensions") {
		t.Fatal("frontend template must expose registerResourcePageExtensions")
	}
	if !strings.Contains(source, "createAction") {
		t.Fatal("frontend template should demonstrate resource-page action extensions")
	}
}

func TestBackendTemplateShowsResourcePageExtensions(t *testing.T) {
	content, err := os.ReadFile(
		filepath.Join("..", "..", "..", "plugins", "templates", "backend-plugin-template", "src", "index.go"),
	)
	if err != nil {
		t.Fatalf("read backend template source: %v", err)
	}

	source := string(content)
	if !strings.Contains(source, "ResourcePageExtensions") {
		t.Fatal("backend template should demonstrate ResourcePageExtensions")
	}
	if !strings.Contains(source, "ResourcePageExtensionAction") {
		t.Fatal("backend template should demonstrate resource-page action descriptors")
	}
}
