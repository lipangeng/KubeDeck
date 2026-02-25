package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"kubedeck/backend/internal/auth"
	"kubedeck/backend/internal/core/audit"
)

type AuthHandler struct {
	provider      auth.Provider
	oauthProvider auth.OAuthProvider
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
	ExpiresAt      time.Time
}

var (
	authSessionsMu sync.RWMutex
	authSessions   = map[string]authSession{}
	oauthStatesMu  sync.Mutex
	oauthStates    = map[string]time.Time{}
)

func NewAuthHandler() *AuthHandler {
	ensureIAMPersistence()
	return &AuthHandler{
		provider:      auth.NewLocalProvider(),
		oauthProvider: auth.NewOAuthProviderFromEnv(),
	}
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
	h.loginWithResolvedUser(w, user, req.TenantCode, req.TenantID, "auth.login")
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
	_ = reloadIAMStateFromPersistence()

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
	storageKey := req.Token
	current, ok := invites[storageKey]
	if !ok {
		storageKey = hashToken(req.Token)
		current, ok = invites[storageKey]
	}
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
		invites[storageKey] = invite
		invitesMu.Unlock()
		persistIAMInvites()
		_ = defaultAuditWriter.Write(audit.Event{TenantID: invite.TenantID, Action: "auth.accept_invite", TargetType: "invite", TargetID: invite.ID, Result: "denied", Reason: "invite_expired"})
		writeJSONError(w, http.StatusGone, "invite_expired")
		return
	}
	invite.Status = "accepted"
	invites[storageKey] = invite
	invitesMu.Unlock()
	persistIAMInvites()
	_ = defaultAuditWriter.Write(audit.Event{TenantID: invite.TenantID, Action: "auth.accept_invite", TargetType: "invite", TargetID: invite.ID, Result: "allowed"})

	_ = writeJSON(w, http.StatusOK, map[string]any{
		"status":    "accepted",
		"tenant_id": invite.TenantID,
		"username":  req.Username,
	})
}

func (h *AuthHandler) OAuthURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	if auth.OAuthProviderInitError(h.oauthProvider) != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "oauth_provider_unavailable")
		return
	}
	state := issueOAuthState()
	_ = writeJSON(w, http.StatusOK, map[string]any{
		"provider": h.oauthProvider.Name(),
		"state":    state,
		"auth_url": h.oauthProvider.BeginAuthURL(state),
	})
}

func (h *AuthHandler) OAuthConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	if auth.IsProductionRuntime() {
		writeJSONError(w, http.StatusNotFound, "not_found")
		return
	}

	diagnostics := auth.OAuthConfigDiagnosticsFromEnv()
	_ = writeJSON(w, http.StatusOK, map[string]any{
		"mode":     diagnostics.Mode,
		"provider": h.oauthProvider.Name(),
		"ready":    diagnostics.Ready,
		"missing":  diagnostics.Missing,
		"oidc": map[string]bool{
			"issuer_exists":        diagnostics.OIDC.IssuerExists,
			"client_id_exists":     diagnostics.OIDC.ClientIDExists,
			"client_secret_exists": diagnostics.OIDC.ClientSecretExists,
			"redirect_url_exists":  diagnostics.OIDC.RedirectURLExists,
		},
	})
}

func (h *AuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	if auth.OAuthProviderInitError(h.oauthProvider) != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "oauth_provider_unavailable")
		return
	}

	var req struct {
		Code       string `json:"code"`
		State      string `json:"state"`
		TenantCode string `json:"tenant_code"`
		TenantID   string `json:"tenant_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !consumeOAuthState(req.State) {
		_ = defaultAuditWriter.Write(audit.Event{Action: "auth.oauth.callback", TargetType: "session", Result: "denied", Reason: "invalid_state"})
		writeJSONError(w, http.StatusBadRequest, "invalid_state")
		return
	}
	user, err := h.oauthProvider.ExchangeCode(req.Code)
	if err != nil {
		_ = defaultAuditWriter.Write(audit.Event{Action: "auth.oauth.callback", TargetType: "session", Result: "denied", Reason: "invalid_oauth_code"})
		writeJSONError(w, http.StatusUnauthorized, "invalid_oauth_code")
		return
	}
	h.loginWithResolvedUser(w, user, req.TenantCode, req.TenantID, "auth.oauth.callback")
}

func (h *AuthHandler) loginWithResolvedUser(
	w http.ResponseWriter,
	user auth.User,
	tenantCode string,
	tenantID string,
	auditAction string,
) {
	tenants, memberships := defaultMembershipsForUser(user.ID)
	activeTenantID := resolveActiveTenant(tenantCode, tenantID, tenants)
	if activeTenantID == "" {
		_ = defaultAuditWriter.Write(audit.Event{ActorID: user.ID, Action: auditAction, TargetType: "tenant", Result: "denied", Reason: "tenant_not_found"})
		writeJSONError(w, http.StatusForbidden, "tenant_not_found")
		return
	}
	if !isMembershipActiveForTenant(memberships, activeTenantID, time.Now().UTC()) {
		_ = defaultAuditWriter.Write(audit.Event{ActorID: user.ID, TenantID: activeTenantID, Action: auditAction, TargetType: "tenant_membership", Result: "denied", Reason: "membership_expired"})
		writeJSONError(w, http.StatusForbidden, "membership_expired")
		return
	}

	user.Memberships = memberships
	user.ActiveTenantID = activeTenantID
	token := newToken()
	expiresAt := time.Now().UTC().Add(authSessionTTL())

	saveAuthSession(token, authSession{
		Token:          token,
		User:           user,
		Available:      tenants,
		ActiveTenantID: activeTenantID,
		ExpiresAt:      expiresAt,
	})
	_ = defaultAuditWriter.Write(audit.Event{ActorID: user.ID, TenantID: activeTenantID, Action: auditAction, TargetType: "session", TargetID: token, Result: "allowed"})

	_ = writeJSON(w, http.StatusOK, map[string]any{
		"token":            token,
		"user":             map[string]any{"id": user.ID, "username": user.Username, "roles": user.Roles},
		"tenants":          tenants,
		"active_tenant_id": activeTenantID,
		"expires_at":       expiresAt,
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
	hashedToken := hashToken(token)
	authSessionsMu.RLock()
	session, ok := authSessions[token]
	if !ok {
		session, ok = authSessions[hashedToken]
	}
	authSessionsMu.RUnlock()
	if ok {
		return token, session, true
	}
	if err := reloadAuthSessionsFromPersistence(); err != nil {
		return token, authSession{}, false
	}
	authSessionsMu.RLock()
	session, ok = authSessions[token]
	if !ok {
		session, ok = authSessions[hashedToken]
	}
	authSessionsMu.RUnlock()
	return token, session, ok
}

func currentValidSessionFromRequest(r *http.Request) (string, authSession, bool, string) {
	token, session, ok := currentSessionWithToken(r)
	if !ok {
		return "", authSession{}, false, "unauthorized"
	}
	if !session.ExpiresAt.IsZero() && !time.Now().UTC().Before(session.ExpiresAt) {
		deleteAuthSession(token)
		return "", authSession{}, false, "session_expired"
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
	hashedToken := hashToken(token)
	authSessionsMu.Lock()
	delete(authSessions, token)
	delete(authSessions, hashedToken)
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

func issueOAuthState() string {
	now := time.Now().UTC()
	state := newToken()
	expiry := now.Add(10 * time.Minute)

	oauthStatesMu.Lock()
	for token, expiresAt := range oauthStates {
		if !now.Before(expiresAt) {
			delete(oauthStates, token)
		}
	}
	oauthStates[state] = expiry
	oauthStatesMu.Unlock()

	return state
}

func consumeOAuthState(state string) bool {
	state = strings.TrimSpace(state)
	if state == "" {
		return false
	}
	now := time.Now().UTC()
	oauthStatesMu.Lock()
	expiry, ok := oauthStates[state]
	if ok {
		delete(oauthStates, state)
	}
	oauthStatesMu.Unlock()
	if !ok {
		return false
	}
	return now.Before(expiry)
}

func authSessionTTL() time.Duration {
	const defaultTTL = 24 * time.Hour
	raw := strings.TrimSpace(os.Getenv("KUBEDECK_AUTH_SESSION_TTL_HOURS"))
	if raw == "" {
		return defaultTTL
	}
	hours, err := strconv.Atoi(raw)
	if err != nil || hours <= 0 {
		return defaultTTL
	}
	return time.Duration(hours) * time.Hour
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}
