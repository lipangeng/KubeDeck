package api

import "net/http"

// NewRouter wires the minimal backend API surface kept during cleanup.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	kernel := NewKernelHandler()

	mux.HandleFunc("/api/meta/kernel", kernel.Snapshot)
	mux.HandleFunc("/api/meta/menus", kernel.Menus)
	mux.HandleFunc("/api/meta/pages", kernel.Pages)
	mux.HandleFunc("/api/meta/actions", kernel.Actions)
	mux.HandleFunc("/api/meta/slots", kernel.Slots)
	mux.HandleFunc("/api/workflows/workloads/items", kernel.Workloads)
	mux.HandleFunc("/api/actions/execute", kernel.ExecuteAction)
	mux.HandleFunc("/api/healthz", healthHandler)
	mux.HandleFunc("/api/readyz", healthHandler)
	return mux
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
