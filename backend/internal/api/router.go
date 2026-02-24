package api

import "net/http"

// NewRouter wires API endpoints for metadata and resource actions.
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	meta := NewMetaHandler()
	resources := NewResourceHandler()

	mux.HandleFunc("/api/meta/registry", meta.Registry)
	mux.HandleFunc("/api/meta/clusters", meta.Clusters)
	mux.HandleFunc("/api/resources/apply", resources.Apply)

	return mux
}
