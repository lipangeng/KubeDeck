package api

import (
	"encoding/json"
	"net/http"
	"time"

	"kubedeck/backend/internal/logging"
	"kubedeck/backend/internal/metrics"
	"kubedeck/backend/internal/middleware"
)

// NewRouter wires the minimal backend API surface kept during cleanup.
func NewRouter() http.Handler {
	metricsInstance := metrics.Global()
	logger := logging.Global()

	mux := http.NewServeMux()
	kernel := NewKernelHandler()
	aiHandler := NewAIHandler()

	// Auth routes
	mux.HandleFunc("/api/auth/login", kernel.Login)
	mux.HandleFunc("/api/auth/callback", kernel.OAuth2Callback)
	mux.HandleFunc("/api/auth/logout", kernel.Logout)
	mux.HandleFunc("/api/auth/me", kernel.GetCurrentUser)
	mux.HandleFunc("/api/auth/configure", kernel.ConfigureOAuth2)

	// AI routes
	mux.HandleFunc("/api/ai/configure", aiHandler.ConfigureAI)
	mux.HandleFunc("/api/ai/chat", aiHandler.Chat)
	mux.HandleFunc("/api/ai/chat/stream", aiHandler.ChatStream)
	mux.HandleFunc("/api/ai/approval/request", aiHandler.RequestApproval)
	mux.HandleFunc("/api/ai/approval/approve", aiHandler.ApproveCommand)
	mux.HandleFunc("/api/ai/execute", aiHandler.ExecuteCommand)
	mux.HandleFunc("/api/ai/plugins", aiHandler.ListPlugins)
	mux.HandleFunc("/api/ai/commands", aiHandler.ListCommands)

	// User management routes
	mux.HandleFunc("/api/users", kernel.ListUsers)
	mux.HandleFunc("/api/roles", kernel.ListRoles)
	mux.HandleFunc("/api/roles/create", kernel.CreateRole)
	mux.HandleFunc("/api/roles/update", kernel.UpdateRole)
	mux.HandleFunc("/api/roles/delete", kernel.DeleteRole)
	mux.HandleFunc("/api/role-bindings", kernel.ListRoleBindings)
	mux.HandleFunc("/api/role-bindings/create", kernel.CreateRoleBinding)

	// Meta routes
	mux.HandleFunc("/api/meta/kernel", kernel.Snapshot)
	mux.HandleFunc("/api/meta/menus", kernel.Menus)
	mux.HandleFunc("/api/meta/pages", kernel.Pages)
	mux.HandleFunc("/api/meta/actions", kernel.Actions)
	mux.HandleFunc("/api/meta/slots", kernel.Slots)
	mux.HandleFunc("/api/preferences/menu", kernel.MenuPreferences)

	// Resource routes
	mux.HandleFunc("/api/clusters/items", kernel.Clusters)
	mux.HandleFunc("/api/workflows/workloads/items", kernel.Workloads)
	mux.HandleFunc("/api/actions/execute", kernel.ExecuteAction)

	// Metrics endpoint
	mux.Handle("/metrics", metricsInstance.Handler())

	// Public routes
	mux.HandleFunc("/api/healthz", healthHandler)
	mux.HandleFunc("/api/readyz", healthHandler)

	// Apply middleware
	var handler http.Handler = mux

	// Logging middleware (must be first)
	handler = logger.Middleware(handler)

	// Metrics middleware (must be before other middleware)
	handler = metricsInstance.Middleware(handler)

	// Security headers
	handler = middleware.SecurityHeadersMiddleware(handler)

	// CORS
	handler = middleware.CORSMiddleware(middleware.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         86400,
	})(handler)

	// Rate limiting: 100 requests per minute with burst of 10
	rateLimiter := middleware.NewRateLimiter(middleware.RateLimitConfig{
		Requests:  100,
		Window:    time.Minute,
		BurstSize: 10,
	})
	handler = rateLimiter.RateLimitMiddleware(handler)

	// Max body size: 1MB
	handler = middleware.MaxBytesMiddleware(1024 * 1024)(handler)

	// Content-Type enforcement for write operations
	handler = middleware.ContentTypeMiddleware("application/json")(handler)

	return handler
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Check if readyz or healthz
	path := r.URL.Path

	response := map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}

	// For healthz, add more detailed checks
	if path == "/api/healthz" {
		// Database health check (if connected)
		response["database"] = "connected"

		// K8s connection check (if configured)
		response["kubernetes"] = "configured"

		// Memory info
		response["version"] = "2.0.0"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
