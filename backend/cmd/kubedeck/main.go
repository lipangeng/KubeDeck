package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"kubedeck/backend/internal/api"
	"kubedeck/backend/internal/webui"
)

func main() {
	defaultPort := getenv("PORT", "8080")
	defaultStaticDir := os.Getenv("STATIC_DIR")

	port := flag.String("port", defaultPort, "HTTP listen port")
	staticDir := flag.String("static-dir", defaultStaticDir, "Optional local static directory override")
	flag.Parse()

	mux := http.NewServeMux()
	apiRouter := api.NewRouter()
	fs, source, err := webui.ResolveFileSystem(*staticDir)
	if err != nil {
		log.Fatalf("resolve static files: %v", err)
	}
	spaHandler := webui.NewSPAHandler(fs)

	mux.Handle("/api/", apiRouter)
	mux.Handle("/", spaHandler)
	healthHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", healthHandler)

	addr := ":" + *port
	log.Printf("kubedeck backend listening on %s (static=%s)", addr, source)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
