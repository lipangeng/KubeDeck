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
}
