package webui

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveFileSystem_EmbeddedDefault(t *testing.T) {
	f, source, err := ResolveFileSystem("")
	if err != nil {
		t.Fatalf("resolve embedded fs: %v", err)
	}
	if source != "embedded" {
		t.Fatalf("expected embedded source, got %q", source)
	}

	indexFile, err := f.Open("index.html")
	if err != nil {
		t.Fatalf("read embedded index: %v", err)
	}
	defer indexFile.Close()
	index, err := io.ReadAll(indexFile)
	if err != nil {
		t.Fatalf("read embedded index body: %v", err)
	}
	if !strings.Contains(string(index), "KubeDeck") {
		t.Fatalf("unexpected embedded index content: %q", string(index))
	}
}

func TestResolveFileSystem_StaticOverride(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "index.html")
	if err := os.WriteFile(indexPath, []byte("override"), 0o644); err != nil {
		t.Fatalf("write temp index: %v", err)
	}

	f, source, err := ResolveFileSystem(dir)
	if err != nil {
		t.Fatalf("resolve override fs: %v", err)
	}
	if !strings.HasPrefix(source, "override:") {
		t.Fatalf("expected override source, got %q", source)
	}

	contentFile, err := f.Open("index.html")
	if err != nil {
		t.Fatalf("read override index: %v", err)
	}
	defer contentFile.Close()
	content, err := io.ReadAll(contentFile)
	if err != nil {
		t.Fatalf("read override index body: %v", err)
	}
	if string(content) != "override" {
		t.Fatalf("unexpected override index: %q", string(content))
	}
}

func TestResolveFileSystem_InvalidOverride(t *testing.T) {
	_, _, err := ResolveFileSystem("/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for invalid override path")
	}
}

func TestSPAHandler_FallbackToIndex(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("INDEX"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	h := NewSPAHandler(http.Dir(dir))
	req := httptest.NewRequest("GET", "/ui/does/not/exist", nil)
	resp := httptest.NewRecorder()

	h.ServeHTTP(resp, req)

	if resp.Code != 200 {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), "INDEX") {
		t.Fatalf("expected fallback index body, got %q", resp.Body.String())
	}
}

func TestSPAHandler_InvalidPath(t *testing.T) {
	h := NewSPAHandler(http.Dir(t.TempDir()))
	req := httptest.NewRequest("GET", "/../../etc/passwd", nil)
	resp := httptest.NewRecorder()

	h.ServeHTTP(resp, req)

	if resp.Code != 400 {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}
