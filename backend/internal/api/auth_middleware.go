package api

import "net/http"

type routePolicy struct {
	requireSession bool
}

func withRoutePolicy(handler http.HandlerFunc, policy routePolicy) http.HandlerFunc {
	if !policy.requireSession {
		return handler
	}
	return requireSession(handler)
}

func requireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _, ok, reason := currentValidSessionFromRequest(r)
		if !ok {
			if reason == "membership_expired" {
				writeJSONError(w, http.StatusForbidden, "membership_expired")
				return
			}
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}
