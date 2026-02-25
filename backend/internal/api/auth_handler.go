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
		writeJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tenants, memberships := defaultMembershipsForUser(user.ID)
	activeTenantID := resolveActiveTenant(req.TenantCode, req.TenantID, tenants)
	if activeTenantID == "" {
		writeJSONError(w, http.StatusForbidden, "tenant_not_found")
		return
	}
	if !isMembershipActiveForTenant(memberships, activeTenantID, time.Now().UTC()) {
		writeJSONError(w, http.StatusForbidden, "membership_expired")
		return
	}

	user.Memberships = memberships
	user.ActiveTenantID = activeTenantID
	token := newToken()

	authSessionsMu.Lock()
	authSessions[token] = authSession{
		Token:          token,
		User:           user,
		Available:      tenants,
		ActiveTenantID: activeTenantID,
	}
	authSessionsMu.Unlock()

	_ = writeJSON(w, http.StatusOK, map[string]any{
		"token":            token,
		"user":             map[string]any{"id": user.ID, "username": user.Username},
		"tenants":          tenants,
		"active_tenant_id": activeTenantID,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	session, ok := currentSessionFromRequest(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":         session.User.ID,
			"username":   session.User.Username,
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

	token, session, ok := currentSessionWithToken(r)
	if !ok {
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
		writeJSONError(w, http.StatusForbidden, "membership_expired")
		return
	}

	session.ActiveTenantID = nextTenantID
	session.User.ActiveTenantID = nextTenantID
	authSessionsMu.Lock()
	authSessions[token] = session
	authSessionsMu.Unlock()

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
		authSessionsMu.Lock()
		delete(authSessions, token)
		authSessionsMu.Unlock()
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

	invitesMu.Lock()
	invite, ok := invites[req.Token]
	if !ok {
		invitesMu.Unlock()
		writeJSONError(w, http.StatusNotFound, "invite_not_found")
		return
	}
	if invite.Status != "pending" {
		invitesMu.Unlock()
		writeJSONError(w, http.StatusConflict, "invite_not_pending")
		return
	}
	if !time.Now().UTC().Before(invite.ExpiresAt) {
		invite.Status = "expired"
		invites[req.Token] = invite
		invitesMu.Unlock()
		writeJSONError(w, http.StatusGone, "invite_expired")
		return
	}
	invite.Status = "accepted"
	invites[req.Token] = invite
	invitesMu.Unlock()

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
	return token, session, ok
}

func newToken() string {
	var raw [24]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "token-fallback"
	}
	return hex.EncodeToString(raw[:])
}
