package api

import "net/http"

// NewRouter wires the minimal backend API surface kept during cleanup.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/healthz", healthHandler)
	mux.HandleFunc("/api/readyz", healthHandler)
	return mux
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
