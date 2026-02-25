package api

import "net/http"

type AuditHandler struct{}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{}
}

func (h *AuditHandler) Events(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	session, ok := mustSession(r, w)
	if !ok {
		return
	}
	events := defaultAuditWriter.List(session.ActiveTenantID)
	_ = writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
	})
}

