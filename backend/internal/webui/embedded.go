package webui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist dist/*
var embeddedFiles embed.FS

func embeddedFileSystem() (http.FileSystem, error) {
	sub, err := fs.Sub(embeddedFiles, "dist")
	if err != nil {
		return nil, err
	}
	return http.FS(sub), nil
}
