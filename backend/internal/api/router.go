package api

import (
	"net/http"
)

// NewRouter wires the minimal backend API surface kept during cleanup.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	kernel := NewKernelHandler()

	// Auth routes
	mux.HandleFunc("/api/auth/login", kernel.Login)
	mux.HandleFunc("/api/auth/callback", kernel.OAuth2Callback)
	mux.HandleFunc("/api/auth/logout", kernel.Logout)
	mux.HandleFunc("/api/auth/me", kernel.GetCurrentUser)

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

	// Public routes
	mux.HandleFunc("/api/healthz", healthHandler)
	mux.HandleFunc("/api/readyz", healthHandler)

	return mux
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
