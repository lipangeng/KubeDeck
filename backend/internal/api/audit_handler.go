package api

import (
	"net/http"
	"strconv"
	"strings"

	"kubedeck/backend/internal/core/audit"
)

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
	events = filterAuditEvents(events, r.URL.Query().Get("action"), r.URL.Query().Get("result"))
	events = limitAuditEvents(events, r.URL.Query().Get("limit"))
	_ = writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
	})
}

func filterAuditEvents(events []audit.Event, actionFilter string, resultFilter string) []audit.Event {
	actionFilter = strings.TrimSpace(actionFilter)
	resultFilter = strings.TrimSpace(resultFilter)
	if actionFilter == "" && resultFilter == "" {
		return events
	}
	out := make([]audit.Event, 0, len(events))
	for _, event := range events {
		if actionFilter != "" && !strings.Contains(event.Action, actionFilter) {
			continue
		}
		if resultFilter != "" && event.Result != resultFilter {
			continue
		}
		out = append(out, event)
	}
	return out
}

func limitAuditEvents(events []audit.Event, rawLimit string) []audit.Event {
	rawLimit = strings.TrimSpace(rawLimit)
	if rawLimit == "" {
		return events
	}
	limit, err := strconv.Atoi(rawLimit)
	if err != nil || limit <= 0 {
		return events
	}
	if len(events) <= limit {
		return events
	}
	return events[len(events)-limit:]
}
