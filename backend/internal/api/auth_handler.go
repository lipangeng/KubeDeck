package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"kubedeck/backend/internal/auth"
	"kubedeck/backend/internal/core/audit"
)

type AuthHandler struct {
	provider auth.Provider
}

type tenantInfo struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type authSession struct {
	Token          string
	User           auth.User
	Available      []tenantInfo
	ActiveTenantID string
}

var (
	authSessionsMu sync.RWMutex
	authSessions   = map[string]authSession{}
)

func NewAuthHandler() *AuthHandler {
	ensureIAMPersistence()
	return &AuthHandler{provider: auth.NewLocalProvider()}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var req struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		TenantCode string `json:"tenant_code"`
		TenantID   string `json:"tenant_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.provider.Authenticate(req.Username, req.Password)
	if err != nil {
		_ = defaultAuditWriter.Write(audit.Event{Action: "auth.login", TargetType: "session", Result: "denied", Reason: "invalid_credentials"})
		writeJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tenants, memberships := defaultMembershipsForUser(user.ID)
	activeTenantID := resolveActiveTenant(req.TenantCode, req.TenantID, tenants)
	if activeTenantID == "" {
		_ = defaultAuditWriter.Write(audit.Event{ActorID: user.ID, Action: "auth.login", TargetType: "tenant", Result: "denied", Reason: "tenant_not_found"})
		writeJSONError(w, http.StatusForbidden, "tenant_not_found")
		return
	}
	if !isMembershipActiveForTenant(memberships, activeTenantID, time.Now().UTC()) {
		_ = defaultAuditWriter.Write(audit.Event{ActorID: user.ID, TenantID: activeTenantID, Action: "auth.login", TargetType: "tenant_membership", Result: "denied", Reason: "membership_expired"})
		writeJSONError(w, http.StatusForbidden, "membership_expired")
		return
	}

	user.Memberships = memberships
	user.ActiveTenantID = activeTenantID
	token := newToken()

	saveAuthSession(token, authSession{
		Token:          token,
		User:           user,
		Available:      tenants,
		ActiveTenantID: activeTenantID,
	})
	_ = defaultAuditWriter.Write(audit.Event{ActorID: user.ID, TenantID: activeTenantID, Action: "auth.login", TargetType: "session", TargetID: token, Result: "allowed"})

	_ = writeJSON(w, http.StatusOK, map[string]any{
		"token":            token,
		"user":             map[string]any{"id": user.ID, "username": user.Username, "roles": user.Roles},
		"tenants":          tenants,
		"active_tenant_id": activeTenantID,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	_, session, ok, reason := currentValidSessionFromRequest(r)
	if !ok {
		if reason == "membership_expired" {
			writeJSONError(w, http.StatusForbidden, "membership_expired")
			return
		}
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":             session.User.ID,
			"username":       session.User.Username,
			"roles":          session.User.Roles,
			"activeTenantID": session.ActiveTenantID,
		},
		"tenants":          session.Available,
		"active_tenant_id": session.ActiveTenantID,
	})
}

func (h *AuthHandler) SwitchTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	token, session, ok, reason := currentValidSessionFromRequest(r)
	if !ok {
		if reason == "membership_expired" {
			writeJSONError(w, http.StatusForbidden, "membership_expired")
			return
		}
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		TenantID   string `json:"tenant_id"`
		TenantCode string `json:"tenant_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	nextTenantID := resolveActiveTenant(req.TenantCode, req.TenantID, session.Available)
	if nextTenantID == "" {
		writeJSONError(w, http.StatusForbidden, "tenant_not_found")
		return
	}
	if !isMembershipActiveForTenant(session.User.Memberships, nextTenantID, time.Now().UTC()) {
		_ = defaultAuditWriter.Write(audit.Event{ActorID: session.User.ID, TenantID: nextTenantID, Action: "auth.switch_tenant", TargetType: "tenant_membership", Result: "denied", Reason: "membership_expired"})
		writeJSONError(w, http.StatusForbidden, "membership_expired")
		return
	}

	session.ActiveTenantID = nextTenantID
	session.User.ActiveTenantID = nextTenantID
	saveAuthSession(token, session)
	_ = defaultAuditWriter.Write(audit.Event{ActorID: session.User.ID, TenantID: nextTenantID, Action: "auth.switch_tenant", TargetType: "session", TargetID: token, Result: "allowed"})

	_ = writeJSON(w, http.StatusOK, map[string]any{
		"active_tenant_id": nextTenantID,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	token := extractBearerToken(r.Header.Get("Authorization"))
	if token != "" {
		var actorID string
		var tenantID string
		authSessionsMu.RLock()
		if existing, ok := authSessions[token]; ok {
			actorID = existing.User.ID
			tenantID = existing.ActiveTenantID
		}
		authSessionsMu.RUnlock()
		deleteAuthSession(token)
		_ = defaultAuditWriter.Write(audit.Event{ActorID: actorID, TenantID: tenantID, Action: "auth.logout", TargetType: "session", TargetID: token, Result: "allowed"})
	}

	_ = writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *AuthHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var req struct {
		Token    string `json:"token"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Token) == "" {
		writeJSONError(w, http.StatusBadRequest, "token_required")
		return
	}

	invite, ok := inviteByToken(req.Token)
	if !ok {
		_ = defaultAuditWriter.Write(audit.Event{Action: "auth.accept_invite", TargetType: "invite", Result: "denied", Reason: "invite_not_found"})
		writeJSONError(w, http.StatusNotFound, "invite_not_found")
		return
	}
	if invite.Status != "pending" {
		_ = defaultAuditWriter.Write(audit.Event{TenantID: invite.TenantID, Action: "auth.accept_invite", TargetType: "invite", TargetID: invite.ID, Result: "denied", Reason: "invite_not_pending"})
		writeJSONError(w, http.StatusConflict, "invite_not_pending")
		return
	}
	invitesMu.Lock()
	current, ok := invites[req.Token]
	if !ok {
		invitesMu.Unlock()
		writeJSONError(w, http.StatusNotFound, "invite_not_found")
		return
	}
	invite = current
	if invite.Status != "pending" {
		invitesMu.Unlock()
		writeJSONError(w, http.StatusConflict, "invite_not_pending")
		return
	}
	if !time.Now().UTC().Before(invite.ExpiresAt) {
		invite.Status = "expired"
		invites[req.Token] = invite
		invitesMu.Unlock()
		persistIAMInvites()
		_ = defaultAuditWriter.Write(audit.Event{TenantID: invite.TenantID, Action: "auth.accept_invite", TargetType: "invite", TargetID: invite.ID, Result: "denied", Reason: "invite_expired"})
		writeJSONError(w, http.StatusGone, "invite_expired")
		return
	}
	invite.Status = "accepted"
	invites[req.Token] = invite
	invitesMu.Unlock()
	persistIAMInvites()
	_ = defaultAuditWriter.Write(audit.Event{TenantID: invite.TenantID, Action: "auth.accept_invite", TargetType: "invite", TargetID: invite.ID, Result: "allowed"})

	_ = writeJSON(w, http.StatusOK, map[string]any{
		"status":    "accepted",
		"tenant_id": invite.TenantID,
		"username":  req.Username,
	})
}

func defaultMembershipsForUser(userID string) ([]tenantInfo, []auth.TenantMembership) {
	now := time.Now().UTC()
	tenants := []tenantInfo{
		{ID: "tenant-dev", Code: "dev", Name: "Development"},
		{ID: "tenant-staging", Code: "staging", Name: "Staging"},
		{ID: "tenant-prod", Code: "prod", Name: "Production"},
	}
	memberships := []auth.TenantMembership{
		{TenantID: "tenant-dev", UserID: userID, EffectiveFrom: now.Add(-24 * time.Hour)},
		{TenantID: "tenant-staging", UserID: userID, EffectiveFrom: now.Add(-24 * time.Hour)},
		{TenantID: "tenant-prod", UserID: userID, EffectiveFrom: now.Add(-24 * time.Hour)},
	}
	return tenants, memberships
}

func resolveActiveTenant(tenantCode, tenantID string, tenants []tenantInfo) string {
	if tenantID != "" {
		for _, t := range tenants {
			if t.ID == tenantID {
				return t.ID
			}
		}
	}
	if tenantCode != "" {
		for _, t := range tenants {
			if t.Code == tenantCode {
				return t.ID
			}
		}
		return ""
	}
	if len(tenants) == 0 {
		return ""
	}
	return tenants[0].ID
}

func isMembershipActiveForTenant(
	memberships []auth.TenantMembership,
	tenantID string,
	now time.Time,
) bool {
	for _, membership := range memberships {
		if membership.TenantID == tenantID {
			return membership.IsActiveAt(now)
		}
	}
	return false
}

func extractBearerToken(value string) string {
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func currentSessionFromRequest(r *http.Request) (authSession, bool) {
	_, session, ok := currentSessionWithToken(r)
	return session, ok
}

func currentSessionWithToken(r *http.Request) (string, authSession, bool) {
	token := extractBearerToken(r.Header.Get("Authorization"))
	if token == "" {
		return "", authSession{}, false
	}
	authSessionsMu.RLock()
	session, ok := authSessions[token]
	authSessionsMu.RUnlock()
	if ok {
		return token, session, true
	}
	if err := reloadAuthSessionsFromPersistence(); err != nil {
		return token, authSession{}, false
	}
	authSessionsMu.RLock()
	session, ok = authSessions[token]
	authSessionsMu.RUnlock()
	return token, session, ok
}

func currentValidSessionFromRequest(r *http.Request) (string, authSession, bool, string) {
	token, session, ok := currentSessionWithToken(r)
	if !ok {
		return "", authSession{}, false, "unauthorized"
	}
	if !isMembershipActiveForTenant(session.User.Memberships, session.ActiveTenantID, time.Now().UTC()) {
		deleteAuthSession(token)
		return "", authSession{}, false, "membership_expired"
	}
	return token, session, true, ""
}

func saveAuthSession(token string, session authSession) {
	authSessionsMu.Lock()
	authSessions[token] = session
	authSessionsMu.Unlock()
	persistAuthSessions()
}

func deleteAuthSession(token string) {
	authSessionsMu.Lock()
	delete(authSessions, token)
	authSessionsMu.Unlock()
	persistAuthSessions()
}

func newToken() string {
	var raw [24]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "token-fallback"
	}
	return hex.EncodeToString(raw[:])
}
