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

	type routeEntry struct {
		pattern string
		handler http.HandlerFunc
		policy  routePolicy
	}
	routes := []routeEntry{
		{pattern: "/api/meta/registry", handler: meta.Registry},
		{pattern: "/api/meta/clusters", handler: meta.Clusters},
		{pattern: "/api/meta/menus", handler: meta.Menus},
		{pattern: "/api/auth/login", handler: authHandler.Login},
		{pattern: "/api/auth/me", handler: authHandler.Me, policy: routePolicy{requireSession: true}},
		{
			pattern: "/api/auth/switch-tenant",
			handler: authHandler.SwitchTenant,
			policy: routePolicy{
				requireSession:          true,
				requiredAnyPermissions: []string{"cluster:switch"},
			},
		},
		{pattern: "/api/auth/logout", handler: authHandler.Logout, policy: routePolicy{requireSession: true}},
		{pattern: "/api/auth/accept-invite", handler: authHandler.AcceptInvite},
		{
			pattern: "/api/iam/permissions",
			handler: iam.Permissions,
			policy: routePolicy{
				requireSession:          true,
				requiredAnyPermissions: []string{"iam:read"},
			},
		},
		{
			pattern: "/api/iam/groups",
			handler: iam.Groups,
			policy: routePolicy{
				requireSession: true,
				methodPermissions: map[string][]string{
					http.MethodGet:  {"iam:read"},
					http.MethodPost: {"iam:write"},
				},
			},
		},
		{
			pattern: "/api/iam/groups/",
			handler: iam.GroupByID,
			policy: routePolicy{
				requireSession: true,
				methodPermissions: map[string][]string{
					http.MethodPatch:  {"iam:write"},
					http.MethodDelete: {"iam:write"},
					http.MethodPut:    {"iam:write"},
				},
			},
		},
		{
			pattern: "/api/iam/users",
			handler: iam.Users,
			policy: routePolicy{
				requireSession:          true,
				requiredAnyPermissions: []string{"iam:read"},
			},
		},
		{
			pattern: "/api/iam/tenants",
			handler: iam.Tenants,
			policy: routePolicy{
				requireSession:          true,
				requiredAnyPermissions: []string{"tenant:read"},
			},
		},
		{
			pattern: "/api/iam/tenants/",
			handler: iam.TenantMembers,
			policy: routePolicy{
				requireSession: true,
				methodPermissions: map[string][]string{
					http.MethodGet:    {"tenant:read"},
					http.MethodPost:   {"tenant:write"},
					http.MethodDelete: {"tenant:write"},
				},
			},
		},
		{
			pattern: "/api/iam/memberships",
			handler: iam.Memberships,
			policy: routePolicy{
				requireSession:          true,
				requiredAnyPermissions: []string{"iam:read"},
			},
		},
		{
			pattern: "/api/iam/memberships/",
			handler: iam.MembershipByID,
			policy: routePolicy{
				requireSession: true,
				methodPermissions: map[string][]string{
					http.MethodPut: {"iam:write"},
				},
			},
		},
		{
			pattern: "/api/iam/invites",
			handler: iam.Invites,
			policy: routePolicy{
				requireSession: true,
				methodPermissions: map[string][]string{
					http.MethodGet:  {"iam:read"},
					http.MethodPost: {"iam:write"},
				},
			},
		},
		{
			pattern: "/api/iam/invites/",
			handler: iam.InviteByID,
			policy: routePolicy{
				requireSession: true,
				methodPermissions: map[string][]string{
					http.MethodDelete: {"iam:write"},
				},
			},
		},
		{
			pattern: "/api/audit/events",
			handler: audit.Events,
			policy: routePolicy{
				requireSession:          true,
				requiredAnyPermissions: []string{"audit:read"},
			},
		},
		{
			pattern: "/api/resources/apply",
			handler: resources.Apply,
			policy: routePolicy{
				requireSession:          true,
				requiredAnyPermissions: []string{"resource:apply"},
			},
		},
	}
	for _, entry := range routes {
		mux.HandleFunc(entry.pattern, withRoutePolicy(entry.handler, entry.policy))
	}

	mux.HandleFunc("/api/healthz", healthHandler)
	mux.HandleFunc("/api/readyz", healthHandler)

	return mux
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
