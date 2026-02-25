package api

import "net/http"

// NewRouter wires API endpoints for metadata and resource actions.
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	meta := NewMetaHandler()
	resources := NewResourceHandler()
	authHandler := NewAuthHandler()
	iam := NewIAMHandler()
	audit := NewAuditHandler()

	mux.HandleFunc("/api/meta/registry", meta.Registry)
	mux.HandleFunc("/api/meta/clusters", meta.Clusters)
	mux.HandleFunc("/api/meta/menus", meta.Menus)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/me", authHandler.Me)
	mux.HandleFunc("/api/auth/switch-tenant", authHandler.SwitchTenant)
	mux.HandleFunc("/api/auth/logout", authHandler.Logout)
	mux.HandleFunc("/api/auth/accept-invite", authHandler.AcceptInvite)
	mux.HandleFunc("/api/iam/permissions", iam.Permissions)
	mux.HandleFunc("/api/iam/groups", iam.Groups)
	mux.HandleFunc("/api/iam/groups/", iam.GroupByID)
	mux.HandleFunc("/api/iam/users", iam.Users)
	mux.HandleFunc("/api/iam/memberships", iam.Memberships)
	mux.HandleFunc("/api/iam/memberships/", iam.MembershipByID)
	mux.HandleFunc("/api/iam/invites", iam.Invites)
	mux.HandleFunc("/api/iam/invites/", iam.InviteByID)
	mux.HandleFunc("/api/audit/events", audit.Events)
	mux.HandleFunc("/api/resources/apply", resources.Apply)
	mux.HandleFunc("/api/healthz", healthHandler)
	mux.HandleFunc("/api/readyz", healthHandler)

	return mux
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
