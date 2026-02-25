package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"kubedeck/backend/internal/core/audit"
	"kubedeck/backend/internal/core/notification"
)

type IAMHandler struct {
	notifier notification.Provider
}

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
	invitesMu   sync.RWMutex
	invites     = map[string]iamInvite{}
)

func NewIAMHandler() *IAMHandler {
	return &IAMHandler{notifier: notification.NewEmailStubProvider()}
}

type iamInvite struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	TenantCode   string    `json:"tenant_code"`
	InviteeEmail string    `json:"invitee_email,omitempty"`
	InviteePhone string    `json:"invitee_phone,omitempty"`
	RoleHint     string    `json:"role_hint,omitempty"`
	Token        string    `json:"token"`
	InviteLink   string    `json:"invite_link"`
	ExpiresAt    time.Time `json:"expires_at"`
	Status       string    `json:"status"`
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

func (h *IAMHandler) Invites(w http.ResponseWriter, r *http.Request) {
	session, ok := mustSession(r, w)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.listInvites(w, session)
	case http.MethodPost:
		if !hasIAMWrite(session.User.Roles) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.createInvite(w, r, session)
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPost)
	}
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
	_ = defaultAuditWriter.Write(audit.Event{
		TenantID:   session.ActiveTenantID,
		ActorID:    session.User.ID,
		Action:     "iam.group.create",
		TargetType: "group",
		TargetID:   group.ID,
		Result:     "allowed",
	})

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
	_ = defaultAuditWriter.Write(audit.Event{
		TenantID:   session.ActiveTenantID,
		ActorID:    session.User.ID,
		Action:     "iam.group.patch",
		TargetType: "group",
		TargetID:   group.ID,
		Result:     "allowed",
	})
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
	_ = defaultAuditWriter.Write(audit.Event{
		TenantID:   session.ActiveTenantID,
		ActorID:    session.User.ID,
		Action:     "iam.group.delete",
		TargetType: "group",
		TargetID:   group.ID,
		Result:     "allowed",
	})
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
	_ = defaultAuditWriter.Write(audit.Event{
		TenantID:   session.ActiveTenantID,
		ActorID:    session.User.ID,
		Action:     "iam.group.permissions.replace",
		TargetType: "group",
		TargetID:   group.ID,
		Result:     "allowed",
	})
	_ = writeJSON(w, http.StatusOK, group)
}

func mustSession(r *http.Request, w http.ResponseWriter) (authSession, bool) {
	_, session, ok, reason := currentValidSessionFromRequest(r)
	if !ok {
		if reason == "membership_expired" {
			writeJSONError(w, http.StatusForbidden, "membership_expired")
			return authSession{}, false
		}
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

func (h *IAMHandler) listInvites(w http.ResponseWriter, session authSession) {
	invitesMu.RLock()
	defer invitesMu.RUnlock()
	out := make([]iamInvite, 0)
	for _, invite := range invites {
		if invite.TenantID == session.ActiveTenantID {
			out = append(out, invite)
		}
	}
	_ = writeJSON(w, http.StatusOK, map[string]any{"invites": out})
}

func (h *IAMHandler) createInvite(w http.ResponseWriter, r *http.Request, session authSession) {
	var req struct {
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		RoleHint  string `json:"role_hint"`
		ExpiresIn int    `json:"expires_in_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Email) == "" && strings.TrimSpace(req.Phone) == "" {
		writeJSONError(w, http.StatusBadRequest, "email_or_phone_required")
		return
	}
	expiresIn := req.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 72
	}
	token := newInviteToken()
	inviteID := "inv-" + token[:12]
	tenantCode := "unknown"
	for _, tenant := range session.Available {
		if tenant.ID == session.ActiveTenantID {
			tenantCode = tenant.Code
			break
		}
	}
	invite := iamInvite{
		ID:           inviteID,
		TenantID:     session.ActiveTenantID,
		TenantCode:   tenantCode,
		InviteeEmail: strings.TrimSpace(req.Email),
		InviteePhone: strings.TrimSpace(req.Phone),
		RoleHint:     req.RoleHint,
		Token:        token,
		InviteLink:   "/accept-invite?token=" + token,
		ExpiresAt:    time.Now().UTC().Add(time.Duration(expiresIn) * time.Hour),
		Status:       "pending",
	}

	invitesMu.Lock()
	invites[token] = invite
	invitesMu.Unlock()
	_ = defaultAuditWriter.Write(audit.Event{
		TenantID:   session.ActiveTenantID,
		ActorID:    session.User.ID,
		Action:     "iam.invite.create",
		TargetType: "invite",
		TargetID:   invite.ID,
		Result:     "allowed",
	})

	if invite.InviteeEmail != "" {
		_ = h.notifier.SendEmail(invite.InviteeEmail, "KubeDeck Invite", invite.InviteLink)
	}
	if invite.InviteePhone != "" {
		_ = h.notifier.SendSMS(invite.InviteePhone, invite.InviteLink)
	}

	_ = writeJSON(w, http.StatusCreated, invite)
}

func newInviteToken() string {
	var raw [24]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "invite-fallback-token"
	}
	return hex.EncodeToString(raw[:])
}
