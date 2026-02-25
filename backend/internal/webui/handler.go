package webui

import (
	"io"
	"net/http"
	"path"
	"strings"
)

type SPAHandler struct {
	fs         http.FileSystem
	fileServer http.Handler
}

func NewSPAHandler(fs http.FileSystem) *SPAHandler {
	return &SPAHandler{fs: fs, fileServer: http.FileServer(fs)}
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	cleaned := path.Clean(r.URL.Path)
	if strings.HasPrefix(cleaned, "../") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	name := strings.TrimPrefix(cleaned, "/")
	if name == "" || name == "." {
		name = "index.html"
	}

	f, err := h.fs.Open(name)
	if err == nil {
		_ = f.Close()
		h.fileServer.ServeHTTP(w, r)
		return
	}

	indexFile, err := h.fs.Open("index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer indexFile.Close()

	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, indexFile)
}
