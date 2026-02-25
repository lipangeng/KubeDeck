package webui

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func ResolveFileSystem(staticDir string) (http.FileSystem, string, error) {
	if staticDir != "" {
		info, err := os.Stat(staticDir)
		if err != nil {
			return nil, "", fmt.Errorf("invalid static dir: %w", err)
		}
		if !info.IsDir() {
			return nil, "", errors.New("invalid static dir: not a directory")
		}

		return http.Dir(staticDir), "override:" + filepath.Clean(staticDir), nil
	}

	embeddedFS, err := embeddedFileSystem()
	if err != nil {
		return nil, "", fmt.Errorf("resolve embedded static files: %w", err)
	}

	f, err := embeddedFS.Open("index.html")
	if err != nil {
		return nil, "", errors.New("embedded static files not found (build frontend assets first)")
	}
	_ = f.Close()

	return embeddedFS, "embedded", nil
}
