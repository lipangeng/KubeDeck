package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"kubedeck/backend/internal/api"
	"kubedeck/backend/internal/webui"
)

func main() {
	defaultPort := getenv("PORT", "8080")
	defaultStaticDir := os.Getenv("STATIC_DIR")
	defaultDBDriver := getenv("KUBEDECK_DB_DRIVER", "sqlite")
	defaultDBDSN := getenv("KUBEDECK_SQLITE_DSN", "")
	defaultDisablePersist := strings.EqualFold(getenv("KUBEDECK_IAM_PERSIST", "1"), "0")

	port := flag.String("port", defaultPort, "HTTP listen port")
	staticDir := flag.String("static-dir", defaultStaticDir, "Optional local static directory override")
	dbDriver := flag.String("db-driver", defaultDBDriver, "Persistence driver: sqlite|mysql|postgres")
	dbDSN := flag.String("db-dsn", defaultDBDSN, "Persistence DSN/path (sqlite default: kubedeck.sqlite)")
	disablePersist := flag.Bool("disable-persist", defaultDisablePersist, "Disable IAM persistence")
	flag.Parse()

	mux := http.NewServeMux()
	api.ConfigurePersistence(api.PersistenceConfig{
		Driver:   *dbDriver,
		DSN:      *dbDSN,
		Disabled: *disablePersist,
	})
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
	log.Printf("kubedeck backend listening on %s (static=%s, dbDriver=%s, persist=%t)", addr, source, *dbDriver, !*disablePersist)
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
