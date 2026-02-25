package api

import "net/http"

type routePolicy struct {
	requireSession         bool
	requiredAnyPermissions []string
	methodPermissions      map[string][]string
}

func withRoutePolicy(handler http.HandlerFunc, policy routePolicy) http.HandlerFunc {
	if !policy.requireSession && len(policy.requiredAnyPermissions) == 0 && len(policy.methodPermissions) == 0 {
		return handler
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var session authSession
		if policy.requireSession || len(policy.requiredAnyPermissions) > 0 || len(policy.methodPermissions) > 0 {
			_, current, ok, reason := currentValidSessionFromRequest(r)
			if !ok {
				if reason == "membership_expired" {
					writeJSONError(w, http.StatusForbidden, "membership_expired")
					return
				}
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			session = current
		}

		required := policy.requiredAnyPermissions
		if perms, ok := policy.methodPermissions[r.Method]; ok {
			required = perms
		}
		if len(required) > 0 && !sessionHasAnyPermission(session, required) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}

		handler(w, r)
	}
}
