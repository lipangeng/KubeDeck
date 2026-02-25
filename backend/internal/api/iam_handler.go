package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

type IAMHandler struct{}

type iamGroup struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenant_id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

var (
	iamGroupsMu sync.RWMutex
	iamGroups   = map[string]iamGroup{}
)

func NewIAMHandler() *IAMHandler {
	return &IAMHandler{}
}

func (h *IAMHandler) Permissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	if _, ok := mustSession(r, w); !ok {
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]any{
		"permissions": []map[string]string{
			{"code": "iam:read", "scope": "platform"},
			{"code": "iam:write", "scope": "platform"},
			{"code": "tenant:read", "scope": "platform"},
			{"code": "tenant:write", "scope": "platform"},
			{"code": "resource:read", "scope": "cluster"},
			{"code": "resource:write", "scope": "cluster"},
		},
	})
}

func (h *IAMHandler) Groups(w http.ResponseWriter, r *http.Request) {
	session, ok := mustSession(r, w)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listGroups(w, session)
	case http.MethodPost:
		if !hasIAMWrite(session.User.Roles) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.createGroup(w, r, session)
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPost)
	}
}

func (h *IAMHandler) GroupByID(w http.ResponseWriter, r *http.Request) {
	session, ok := mustSession(r, w)
	if !ok {
		return
	}

	trimmed := strings.TrimPrefix(r.URL.Path, "/api/iam/groups/")
	if trimmed == "" {
		writeJSONError(w, http.StatusBadRequest, "group_id_required")
		return
	}
	if strings.HasSuffix(trimmed, "/permissions") {
		groupID := strings.TrimSuffix(trimmed, "/permissions")
		if r.Method != http.MethodPut {
			methodNotAllowed(w, http.MethodPut)
			return
		}
		if !hasIAMWrite(session.User.Roles) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.replaceGroupPermissions(w, r, session, groupID)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		if !hasIAMWrite(session.User.Roles) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.patchGroup(w, r, session, trimmed)
	case http.MethodDelete:
		if !hasIAMWrite(session.User.Roles) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.deleteGroup(w, session, trimmed)
	default:
		methodNotAllowed(w, http.MethodPatch+", "+http.MethodDelete)
	}
}

func (h *IAMHandler) ReplaceMembershipGroups(w http.ResponseWriter, r *http.Request) {
	session, ok := mustSession(r, w)
	if !ok {
		return
	}
	if r.Method != http.MethodPut {
		methodNotAllowed(w, http.MethodPut)
		return
	}
	if !hasIAMWrite(session.User.Roles) {
		writeJSONError(w, http.StatusForbidden, "permission_denied")
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *IAMHandler) listGroups(w http.ResponseWriter, session authSession) {
	iamGroupsMu.RLock()
	defer iamGroupsMu.RUnlock()

	out := make([]iamGroup, 0)
	for _, group := range iamGroups {
		if group.TenantID == session.ActiveTenantID {
			out = append(out, group)
		}
	}
	_ = writeJSON(w, http.StatusOK, map[string]any{"groups": out})
}

func (h *IAMHandler) createGroup(w http.ResponseWriter, r *http.Request, session authSession) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeJSONError(w, http.StatusBadRequest, "name_required")
		return
	}
	id := "grp-" + strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
	group := iamGroup{
		ID:          id,
		TenantID:    session.ActiveTenantID,
		Name:        req.Name,
		Description: req.Description,
	}
	iamGroupsMu.Lock()
	iamGroups[id] = group
	iamGroupsMu.Unlock()

	_ = writeJSON(w, http.StatusCreated, group)
}

func (h *IAMHandler) patchGroup(
	w http.ResponseWriter,
	r *http.Request,
	session authSession,
	groupID string,
) {
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	iamGroupsMu.Lock()
	defer iamGroupsMu.Unlock()
	group, ok := iamGroups[groupID]
	if !ok || group.TenantID != session.ActiveTenantID {
		writeJSONError(w, http.StatusNotFound, "group_not_found")
		return
	}
	if req.Name != nil {
		group.Name = *req.Name
	}
	if req.Description != nil {
		group.Description = *req.Description
	}
	iamGroups[groupID] = group
	_ = writeJSON(w, http.StatusOK, group)
}

func (h *IAMHandler) deleteGroup(w http.ResponseWriter, session authSession, groupID string) {
	iamGroupsMu.Lock()
	defer iamGroupsMu.Unlock()
	group, ok := iamGroups[groupID]
	if !ok || group.TenantID != session.ActiveTenantID {
		writeJSONError(w, http.StatusNotFound, "group_not_found")
		return
	}
	delete(iamGroups, groupID)
	_ = writeJSON(w, http.StatusOK, map[string]any{"deleted": group.ID})
}

func (h *IAMHandler) replaceGroupPermissions(
	w http.ResponseWriter,
	r *http.Request,
	session authSession,
	groupID string,
) {
	var req struct {
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	iamGroupsMu.Lock()
	defer iamGroupsMu.Unlock()
	group, ok := iamGroups[groupID]
	if !ok || group.TenantID != session.ActiveTenantID {
		writeJSONError(w, http.StatusNotFound, "group_not_found")
		return
	}
	group.Permissions = append([]string{}, req.Permissions...)
	iamGroups[groupID] = group
	_ = writeJSON(w, http.StatusOK, group)
}

func mustSession(r *http.Request, w http.ResponseWriter) (authSession, bool) {
	session, ok := currentSessionFromRequest(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return authSession{}, false
	}
	return session, true
}

func hasIAMWrite(roles []string) bool {
	for _, role := range roles {
		if role == "admin" || role == "owner" {
			return true
		}
	}
	return false
}

