package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sort"
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
	iamGroupsMu      sync.RWMutex
	iamGroups        = map[string]iamGroup{}
	iamMembershipsMu sync.RWMutex
	iamMemberships   = map[string]iamMembership{}
	invitesMu        sync.RWMutex
	invites          = map[string]iamInvite{}
)

type iamMembership struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	UserID         string     `json:"user_id"`
	UserLabel      string     `json:"user_label"`
	GroupIDs       []string   `json:"group_ids"`
	EffectiveFrom  time.Time  `json:"effective_from"`
	EffectiveUntil *time.Time `json:"effective_until,omitempty"`
}

type iamUser struct {
	ID             string     `json:"id"`
	Username       string     `json:"username"`
	Roles          []string   `json:"roles"`
	TenantID       string     `json:"tenant_id"`
	MembershipID   string     `json:"membership_id"`
	EffectiveFrom  time.Time  `json:"effective_from"`
	EffectiveUntil *time.Time `json:"effective_until,omitempty"`
}

type iamTenant struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

func NewIAMHandler() *IAMHandler {
	ensureIAMPersistence()
	return &IAMHandler{notifier: notification.NewProviderFromEnv()}
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
	CreatedAt    time.Time `json:"created_at"`
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
			{"code": "audit:read", "scope": "platform"},
			{"code": "menu:read", "scope": "platform"},
			{"code": "menu:write", "scope": "platform"},
			{"code": "resource:read", "scope": "cluster"},
			{"code": "resource:write", "scope": "cluster"},
			{"code": "resource:apply", "scope": "cluster"},
			{"code": "resource:delete", "scope": "cluster"},
			{"code": "cluster:switch", "scope": "platform"},
		},
	})
}

func (h *IAMHandler) Groups(w http.ResponseWriter, r *http.Request) {
	session, ok := mustSession(r, w)
	if !ok {
		return
	}
	_ = reloadIAMStateFromPersistence()

	switch r.Method {
	case http.MethodGet:
		h.listGroups(w, session)
	case http.MethodPost:
		if !sessionCanIAMWrite(session) {
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
	_ = reloadIAMStateFromPersistence()

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
		if !sessionCanIAMWrite(session) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.replaceGroupPermissions(w, r, session, groupID)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		if !sessionCanIAMWrite(session) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.patchGroup(w, r, session, trimmed)
	case http.MethodDelete:
		if !sessionCanIAMWrite(session) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.deleteGroup(w, session, trimmed)
	default:
		methodNotAllowed(w, http.MethodPatch+", "+http.MethodDelete)
	}
}

func (h *IAMHandler) Memberships(w http.ResponseWriter, r *http.Request) {
	session, ok := mustSession(r, w)
	if !ok {
		return
	}
	_ = reloadIAMStateFromPersistence()
	switch r.Method {
	case http.MethodGet:
		h.listMemberships(w, session)
	default:
		methodNotAllowed(w, http.MethodGet)
	}
}

func (h *IAMHandler) MembershipByID(w http.ResponseWriter, r *http.Request) {
	session, ok := mustSession(r, w)
	if !ok {
		return
	}
	_ = reloadIAMStateFromPersistence()
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/iam/memberships/")
	if trimmed == "" {
		writeJSONError(w, http.StatusBadRequest, "membership_id_required")
		return
	}
	switch {
	case strings.HasSuffix(trimmed, "/groups"):
		membershipID := strings.TrimSuffix(trimmed, "/groups")
		if r.Method != http.MethodPut {
			methodNotAllowed(w, http.MethodPut)
			return
		}
		if !sessionCanIAMWrite(session) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.replaceMembershipGroups(w, r, session, membershipID)
	case strings.HasSuffix(trimmed, "/validity"):
		membershipID := strings.TrimSuffix(trimmed, "/validity")
		if r.Method != http.MethodPut {
			methodNotAllowed(w, http.MethodPut)
			return
		}
		if !sessionCanIAMWrite(session) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.replaceMembershipValidity(w, r, session, membershipID)
	default:
		methodNotAllowed(w, http.MethodPut)
	}
}

func (h *IAMHandler) Invites(w http.ResponseWriter, r *http.Request) {
	session, ok := mustSession(r, w)
	if !ok {
		return
	}
	_ = reloadIAMStateFromPersistence()
	switch r.Method {
	case http.MethodGet:
		h.listInvites(w, session)
	case http.MethodPost:
		if !sessionCanIAMWrite(session) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.createInvite(w, r, session)
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPost)
	}
}

func (h *IAMHandler) InviteByID(w http.ResponseWriter, r *http.Request) {
	session, ok := mustSession(r, w)
	if !ok {
		return
	}
	inviteID := strings.TrimPrefix(r.URL.Path, "/api/iam/invites/")
	if strings.TrimSpace(inviteID) == "" {
		writeJSONError(w, http.StatusBadRequest, "invite_id_required")
		return
	}
	if r.Method != http.MethodDelete {
		methodNotAllowed(w, http.MethodDelete)
		return
	}
	if !sessionCanIAMWrite(session) {
		writeJSONError(w, http.StatusForbidden, "permission_denied")
		return
	}
	h.revokeInvite(w, session, inviteID)
}

func (h *IAMHandler) Users(w http.ResponseWriter, r *http.Request) {
	session, ok := mustSession(r, w)
	if !ok {
		return
	}
	_ = reloadIAMStateFromPersistence()
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	h.listUsers(w, session)
}

func (h *IAMHandler) Tenants(w http.ResponseWriter, r *http.Request) {
	session, ok := mustSession(r, w)
	if !ok {
		return
	}
	_ = reloadIAMStateFromPersistence()
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	tenants := make([]iamTenant, 0, len(session.Available))
	for _, tenant := range session.Available {
		tenants = append(tenants, iamTenant{
			ID:   tenant.ID,
			Code: tenant.Code,
			Name: tenant.Name,
		})
	}
	_ = writeJSON(w, http.StatusOK, map[string]any{"tenants": tenants})
}

func (h *IAMHandler) TenantMembers(w http.ResponseWriter, r *http.Request) {
	session, ok := mustSession(r, w)
	if !ok {
		return
	}
	_ = reloadIAMStateFromPersistence()
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/iam/tenants/")
	if !strings.HasSuffix(trimmed, "/members") {
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodDelete)
		return
	}
	tenantID := strings.TrimSuffix(trimmed, "/members")
	tenantID = strings.TrimSuffix(tenantID, "/")
	if tenantID == "" {
		writeJSONError(w, http.StatusBadRequest, "tenant_id_required")
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.listTenantMembers(w, session, tenantID)
	case http.MethodPost:
		if !sessionCanIAMWrite(session) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.createTenantMember(w, r, session, tenantID)
	case http.MethodDelete:
		if !sessionCanIAMWrite(session) {
			writeJSONError(w, http.StatusForbidden, "permission_denied")
			return
		}
		h.deleteTenantMember(w, r, session, tenantID)
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodDelete)
	}
}

func (h *IAMHandler) listGroups(w http.ResponseWriter, session authSession) {
	ensureDefaultTenantGroups(session.ActiveTenantID)
	out := groupsByTenant(session.ActiveTenantID)
	_ = writeJSON(w, http.StatusOK, map[string]any{"groups": out})
}

func (h *IAMHandler) listMemberships(w http.ResponseWriter, session authSession) {
	out := membershipsByTenant(session.ActiveTenantID)
	if len(out) == 0 {
		seed := iamMembership{
			ID:            "mbr-" + session.User.ID + "-" + session.ActiveTenantID,
			TenantID:      session.ActiveTenantID,
			UserID:        session.User.ID,
			UserLabel:     session.User.Username,
			GroupIDs:      []string{},
			EffectiveFrom: time.Now().UTC().Add(-24 * time.Hour),
		}
		iamMemberships[seed.ID] = seed
		iamMembershipsMu.Unlock()
		persistIAMMemberships()
		out = append(out, seed)
	} else {
		iamMembershipsMu.Unlock()
	}
	_ = writeJSON(w, http.StatusOK, map[string]any{"memberships": out})
}

func (h *IAMHandler) listUsers(w http.ResponseWriter, session authSession) {
	memberships := membershipsByTenant(session.ActiveTenantID)
	if len(memberships) == 0 {
		seed := iamMembership{
			ID:            "mbr-" + session.User.ID + "-" + session.ActiveTenantID,
			TenantID:      session.ActiveTenantID,
			UserID:        session.User.ID,
			UserLabel:     session.User.Username,
			GroupIDs:      []string{},
			EffectiveFrom: time.Now().UTC().Add(-24 * time.Hour),
		}
		iamMemberships[seed.ID] = seed
		iamMembershipsMu.Unlock()
		persistIAMMemberships()
		memberships = append(memberships, seed)
	} else {
		iamMembershipsMu.Unlock()
	}

	users := make([]iamUser, 0, len(memberships))
	for _, membership := range memberships {
		roles := []string{"viewer"}
		if membership.UserID == session.User.ID {
			roles = append([]string{}, session.User.Roles...)
		}
		users = append(users, iamUser{
			ID:             membership.UserID,
			Username:       membership.UserLabel,
			Roles:          roles,
			TenantID:       membership.TenantID,
			MembershipID:   membership.ID,
			EffectiveFrom:  membership.EffectiveFrom,
			EffectiveUntil: membership.EffectiveUntil,
		})
	}

	_ = writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (h *IAMHandler) listTenantMembers(
	w http.ResponseWriter,
	session authSession,
	tenantID string,
) {
	if !tenantVisibleToSession(session, tenantID) {
		writeJSONError(w, http.StatusNotFound, "tenant_not_found")
		return
	}
	out := membershipsByTenant(tenantID)
	iamMembershipsMu.Unlock()
	_ = writeJSON(w, http.StatusOK, map[string]any{"members": out})
}

func (h *IAMHandler) createTenantMember(
	w http.ResponseWriter,
	r *http.Request,
	session authSession,
	tenantID string,
) {
	if !tenantVisibleToSession(session, tenantID) {
		writeJSONError(w, http.StatusNotFound, "tenant_not_found")
		return
	}
	var req struct {
		UserID         string `json:"user_id"`
		UserLabel      string `json:"user_label"`
		EffectiveFrom  string `json:"effective_from"`
		EffectiveUntil string `json:"effective_until"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.UserID) == "" {
		writeJSONError(w, http.StatusBadRequest, "user_id_required")
		return
	}
	var effectiveFrom time.Time
	if strings.TrimSpace(req.EffectiveFrom) == "" {
		effectiveFrom = time.Now().UTC()
	} else {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EffectiveFrom))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_effective_from")
			return
		}
		effectiveFrom = parsed.UTC()
	}
	var effectiveUntilPtr *time.Time
	if strings.TrimSpace(req.EffectiveUntil) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EffectiveUntil))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_effective_until")
			return
		}
		if parsed.Before(effectiveFrom) {
			writeJSONError(w, http.StatusBadRequest, "effective_until_before_from")
			return
		}
		parsed = parsed.UTC()
		effectiveUntilPtr = &parsed
	}
	userLabel := strings.TrimSpace(req.UserLabel)
	if userLabel == "" {
		userLabel = req.UserID
	}
	if membershipExistsByTenantUser(tenantID, req.UserID) {
		writeJSONError(w, http.StatusConflict, "membership_exists")
		return
	}
	membership := iamMembership{
		ID:             "mbr-" + strings.ToLower(strings.ReplaceAll(req.UserID, " ", "-")) + "-" + tenantID,
		TenantID:       tenantID,
		UserID:         req.UserID,
		UserLabel:      userLabel,
		GroupIDs:       []string{},
		EffectiveFrom:  effectiveFrom,
		EffectiveUntil: effectiveUntilPtr,
	}
	iamMembershipsMu.Lock()
	iamMemberships[membership.ID] = membership
	iamMembershipsMu.Unlock()
	persistIAMMemberships()
	_ = defaultAuditWriter.Write(audit.Event{
		TenantID:   tenantID,
		ActorID:    session.User.ID,
		Action:     "iam.tenant.member.create",
		TargetType: "membership",
		TargetID:   membership.ID,
		Result:     "allowed",
	})
	_ = writeJSON(w, http.StatusCreated, membership)
}

func (h *IAMHandler) deleteTenantMember(
	w http.ResponseWriter,
	r *http.Request,
	session authSession,
	tenantID string,
) {
	if !tenantVisibleToSession(session, tenantID) {
		writeJSONError(w, http.StatusNotFound, "tenant_not_found")
		return
	}
	membershipID := strings.TrimSpace(r.URL.Query().Get("membership_id"))
	if membershipID == "" {
		writeJSONError(w, http.StatusBadRequest, "membership_id_required")
		return
	}
	iamMembershipsMu.Lock()
	membership, ok := iamMemberships[membershipID]
	if !ok {
		iamMembershipsMu.Unlock()
		_ = reloadIAMStateFromPersistence()
		iamMembershipsMu.Lock()
		membership, ok = iamMemberships[membershipID]
	}
	if !ok || membership.TenantID != tenantID {
		iamMembershipsMu.Unlock()
		writeJSONError(w, http.StatusNotFound, "membership_not_found")
		return
	}
	delete(iamMemberships, membershipID)
	iamMembershipsMu.Unlock()
	persistIAMMemberships()
	_ = defaultAuditWriter.Write(audit.Event{
		TenantID:   tenantID,
		ActorID:    session.User.ID,
		Action:     "iam.tenant.member.delete",
		TargetType: "membership",
		TargetID:   membershipID,
		Result:     "allowed",
	})
	_ = writeJSON(w, http.StatusOK, map[string]any{"deleted": membershipID})
}

func tenantVisibleToSession(session authSession, tenantID string) bool {
	for _, tenant := range session.Available {
		if tenant.ID == tenantID {
			return true
		}
	}
	return false
}

func groupsByTenant(tenantID string) []iamGroup {
	ensureDefaultTenantGroups(tenantID)
	load := func() []iamGroup {
		out := make([]iamGroup, 0)
		for _, group := range iamGroups {
			if group.TenantID == tenantID {
				out = append(out, group)
			}
		}
		return out
	}
	iamGroupsMu.RLock()
	out := load()
	iamGroupsMu.RUnlock()
	if len(out) > 0 {
		return out
	}
	_ = reloadIAMStateFromPersistence()
	iamGroupsMu.RLock()
	out = load()
	iamGroupsMu.RUnlock()
	return out
}

func ensureDefaultTenantGroups(tenantID string) {
	defaults := []iamGroup{
		{
			ID:          "grp-" + tenantID + "-owner",
			TenantID:    tenantID,
			Name:        "tenant-owner",
			Description: "Tenant owner group",
			Permissions: []string{"*"},
		},
		{
			ID:          "grp-" + tenantID + "-admin",
			TenantID:    tenantID,
			Name:        "tenant-admin",
			Description: "Tenant admin group",
			Permissions: []string{"iam:write", "iam:read", "tenant:write", "tenant:read", "audit:read", "resource:apply", "resource:write", "resource:read"},
		},
		{
			ID:          "grp-" + tenantID + "-viewer",
			TenantID:    tenantID,
			Name:        "tenant-viewer",
			Description: "Tenant viewer group",
			Permissions: []string{"iam:read", "tenant:read", "menu:read", "resource:read", "cluster:switch"},
		},
	}

	changed := false
	iamGroupsMu.Lock()
	for _, group := range defaults {
		existing, ok := iamGroups[group.ID]
		if ok {
			if existing.TenantID == tenantID && len(existing.Permissions) == 0 {
				existing.Permissions = append([]string{}, group.Permissions...)
				if strings.TrimSpace(existing.Description) == "" {
					existing.Description = group.Description
				}
				iamGroups[group.ID] = existing
				changed = true
			}
			continue
		}
		iamGroups[group.ID] = group
		changed = true
	}
	iamGroupsMu.Unlock()

	if changed {
		persistIAMGroups()
	}
}

func membershipsByTenant(tenantID string) []iamMembership {
	load := func() []iamMembership {
		out := make([]iamMembership, 0)
		for _, membership := range iamMemberships {
			if membership.TenantID == tenantID {
				out = append(out, membership)
			}
		}
		return out
	}
	iamMembershipsMu.Lock()
	out := load()
	if len(out) > 0 {
		return out
	}
	iamMembershipsMu.Unlock()
	_ = reloadIAMStateFromPersistence()
	iamMembershipsMu.Lock()
	return load()
}

func membershipExistsByTenantUser(tenantID string, userID string) bool {
	iamMembershipsMu.RLock()
	for _, existing := range iamMemberships {
		if existing.TenantID == tenantID && strings.EqualFold(existing.UserID, userID) {
			iamMembershipsMu.RUnlock()
			return true
		}
	}
	iamMembershipsMu.RUnlock()
	_ = reloadIAMStateFromPersistence()
	iamMembershipsMu.RLock()
	defer iamMembershipsMu.RUnlock()
	for _, existing := range iamMemberships {
		if existing.TenantID == tenantID && strings.EqualFold(existing.UserID, userID) {
			return true
		}
	}
	return false
}

func groupNameExistsByTenant(tenantID string, groupName string) bool {
	iamGroupsMu.RLock()
	for _, existing := range iamGroups {
		if existing.TenantID == tenantID && strings.EqualFold(existing.Name, groupName) {
			iamGroupsMu.RUnlock()
			return true
		}
	}
	iamGroupsMu.RUnlock()
	_ = reloadIAMStateFromPersistence()
	iamGroupsMu.RLock()
	defer iamGroupsMu.RUnlock()
	for _, existing := range iamGroups {
		if existing.TenantID == tenantID && strings.EqualFold(existing.Name, groupName) {
			return true
		}
	}
	return false
}

func groupByID(groupID string) (iamGroup, bool) {
	iamGroupsMu.RLock()
	group, ok := iamGroups[groupID]
	iamGroupsMu.RUnlock()
	if ok {
		return group, true
	}
	_ = reloadIAMStateFromPersistence()
	iamGroupsMu.RLock()
	defer iamGroupsMu.RUnlock()
	group, ok = iamGroups[groupID]
	return group, ok
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
	normalizedName := strings.TrimSpace(req.Name)
	if groupNameExistsByTenant(session.ActiveTenantID, normalizedName) {
		writeJSONError(w, http.StatusConflict, "group_exists")
		return
	}
	id := "grp-" + strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
	group := iamGroup{
		ID:          id,
		TenantID:    session.ActiveTenantID,
		Name:        normalizedName,
		Description: req.Description,
	}
	iamGroupsMu.Lock()
	iamGroups[id] = group
	iamGroupsMu.Unlock()
	persistIAMGroups()
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
	group, ok := iamGroups[groupID]
	if !ok {
		iamGroupsMu.Unlock()
		_ = reloadIAMStateFromPersistence()
		iamGroupsMu.Lock()
		group, ok = iamGroups[groupID]
	}
	if !ok || group.TenantID != session.ActiveTenantID {
		iamGroupsMu.Unlock()
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
	iamGroupsMu.Unlock()
	persistIAMGroups()
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
	group, ok := iamGroups[groupID]
	if !ok {
		iamGroupsMu.Unlock()
		_ = reloadIAMStateFromPersistence()
		iamGroupsMu.Lock()
		group, ok = iamGroups[groupID]
	}
	if !ok || group.TenantID != session.ActiveTenantID {
		iamGroupsMu.Unlock()
		writeJSONError(w, http.StatusNotFound, "group_not_found")
		return
	}
	delete(iamGroups, groupID)
	iamGroupsMu.Unlock()
	persistIAMGroups()
	// Keep memberships consistent after group deletion by dropping stale group refs.
	iamMembershipsMu.Lock()
	for id, membership := range iamMemberships {
		if membership.TenantID != session.ActiveTenantID || len(membership.GroupIDs) == 0 {
			continue
		}
		nextGroupIDs := make([]string, 0, len(membership.GroupIDs))
		changed := false
		for _, existingGroupID := range membership.GroupIDs {
			if existingGroupID == groupID {
				changed = true
				continue
			}
			nextGroupIDs = append(nextGroupIDs, existingGroupID)
		}
		if changed {
			membership.GroupIDs = nextGroupIDs
			iamMemberships[id] = membership
		}
	}
	iamMembershipsMu.Unlock()
	persistIAMMemberships()
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
	group, ok := iamGroups[groupID]
	if !ok {
		iamGroupsMu.Unlock()
		_ = reloadIAMStateFromPersistence()
		iamGroupsMu.Lock()
		group, ok = iamGroups[groupID]
	}
	if !ok || group.TenantID != session.ActiveTenantID {
		iamGroupsMu.Unlock()
		writeJSONError(w, http.StatusNotFound, "group_not_found")
		return
	}
	group.Permissions = append([]string{}, req.Permissions...)
	iamGroups[groupID] = group
	iamGroupsMu.Unlock()
	persistIAMGroups()
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

func (h *IAMHandler) replaceMembershipGroups(
	w http.ResponseWriter,
	r *http.Request,
	session authSession,
	membershipID string,
) {
	var req struct {
		GroupIDs []string `json:"group_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	for _, groupID := range req.GroupIDs {
		group, ok := groupByID(groupID)
		if !ok || group.TenantID != session.ActiveTenantID {
			writeJSONError(w, http.StatusBadRequest, "unknown_group_ids")
			return
		}
	}

	iamMembershipsMu.Lock()
	membership, ok := iamMemberships[membershipID]
	if !ok {
		iamMembershipsMu.Unlock()
		_ = reloadIAMStateFromPersistence()
		iamMembershipsMu.Lock()
		membership, ok = iamMemberships[membershipID]
	}
	if !ok {
		iamMembershipsMu.Unlock()
		writeJSONError(w, http.StatusNotFound, "membership_not_found")
		return
	}
	if membership.TenantID != session.ActiveTenantID {
		iamMembershipsMu.Unlock()
		writeJSONError(w, http.StatusNotFound, "membership_not_found")
		return
	}
	membership.GroupIDs = append([]string{}, req.GroupIDs...)
	iamMemberships[membershipID] = membership
	iamMembershipsMu.Unlock()
	persistIAMMemberships()
	_ = defaultAuditWriter.Write(audit.Event{
		TenantID:   session.ActiveTenantID,
		ActorID:    session.User.ID,
		Action:     "iam.membership.groups.replace",
		TargetType: "membership",
		TargetID:   membership.ID,
		Result:     "allowed",
	})
	_ = writeJSON(w, http.StatusOK, membership)
}

func (h *IAMHandler) replaceMembershipValidity(
	w http.ResponseWriter,
	r *http.Request,
	session authSession,
	membershipID string,
) {
	var req struct {
		EffectiveFrom  string `json:"effective_from"`
		EffectiveUntil string `json:"effective_until"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	effectiveFrom, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EffectiveFrom))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_effective_from")
		return
	}
	var effectiveUntilPtr *time.Time
	if trimmed := strings.TrimSpace(req.EffectiveUntil); trimmed != "" {
		effectiveUntil, parseErr := time.Parse(time.RFC3339, trimmed)
		if parseErr != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_effective_until")
			return
		}
		if effectiveUntil.Before(effectiveFrom) {
			writeJSONError(w, http.StatusBadRequest, "effective_until_before_from")
			return
		}
		effectiveUntilPtr = &effectiveUntil
	}

	iamMembershipsMu.Lock()
	membership, ok := iamMemberships[membershipID]
	if !ok {
		iamMembershipsMu.Unlock()
		_ = reloadIAMStateFromPersistence()
		iamMembershipsMu.Lock()
		membership, ok = iamMemberships[membershipID]
	}
	if !ok || membership.TenantID != session.ActiveTenantID {
		iamMembershipsMu.Unlock()
		writeJSONError(w, http.StatusNotFound, "membership_not_found")
		return
	}
	membership.EffectiveFrom = effectiveFrom.UTC()
	membership.EffectiveUntil = effectiveUntilPtr
	iamMemberships[membershipID] = membership
	iamMembershipsMu.Unlock()
	persistIAMMemberships()
	_ = defaultAuditWriter.Write(audit.Event{
		TenantID:   session.ActiveTenantID,
		ActorID:    session.User.ID,
		Action:     "iam.membership.validity.replace",
		TargetType: "membership",
		TargetID:   membership.ID,
		Result:     "allowed",
	})
	_ = writeJSON(w, http.StatusOK, membership)
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

func sessionCanIAMWrite(session authSession) bool {
	return sessionHasAnyPermission(session, []string{"iam:write"})
}

func (h *IAMHandler) listInvites(w http.ResponseWriter, session authSession) {
	out := invitesByTenant(session.ActiveTenantID)
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	for i := range out {
		out[i] = sanitizeInviteForRead(out[i])
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
	now := time.Now().UTC()
	invite := iamInvite{
		ID:           inviteID,
		TenantID:     session.ActiveTenantID,
		TenantCode:   tenantCode,
		InviteeEmail: strings.TrimSpace(req.Email),
		InviteePhone: strings.TrimSpace(req.Phone),
		RoleHint:     req.RoleHint,
		Token:        token,
		InviteLink:   "#/accept-invite?token=" + token,
		CreatedAt:    now,
		ExpiresAt:    now.Add(time.Duration(expiresIn) * time.Hour),
		Status:       "pending",
	}
	storedInvite := sanitizeInviteForStorage(invite)
	storageKey := hashToken(token)

	invitesMu.Lock()
	invites[storageKey] = storedInvite
	invitesMu.Unlock()
	persistIAMInvites()
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

func (h *IAMHandler) revokeInvite(w http.ResponseWriter, session authSession, inviteID string) {
	invitesMu.Lock()
	for token, invite := range invites {
		if invite.ID != inviteID {
			continue
		}
		if invite.TenantID != session.ActiveTenantID {
			invitesMu.Unlock()
			writeJSONError(w, http.StatusNotFound, "invite_not_found")
			return
		}
		if invite.Status != "pending" {
			invitesMu.Unlock()
			writeJSONError(w, http.StatusConflict, "invite_not_pending")
			return
		}
		invite.Status = "revoked"
		invites[token] = invite
		invitesMu.Unlock()
		persistIAMInvites()
		_ = defaultAuditWriter.Write(audit.Event{
			TenantID:   session.ActiveTenantID,
			ActorID:    session.User.ID,
			Action:     "iam.invite.revoke",
			TargetType: "invite",
			TargetID:   invite.ID,
			Result:     "allowed",
		})
		_ = writeJSON(w, http.StatusOK, invite)
		return
	}
	invitesMu.Unlock()
	_ = reloadIAMStateFromPersistence()
	invitesMu.Lock()
	for token, invite := range invites {
		if invite.ID != inviteID {
			continue
		}
		if invite.TenantID != session.ActiveTenantID {
			invitesMu.Unlock()
			writeJSONError(w, http.StatusNotFound, "invite_not_found")
			return
		}
		if invite.Status != "pending" {
			invitesMu.Unlock()
			writeJSONError(w, http.StatusConflict, "invite_not_pending")
			return
		}
		invite.Status = "revoked"
		invites[token] = invite
		invitesMu.Unlock()
		persistIAMInvites()
		_ = defaultAuditWriter.Write(audit.Event{
			TenantID:   session.ActiveTenantID,
			ActorID:    session.User.ID,
			Action:     "iam.invite.revoke",
			TargetType: "invite",
			TargetID:   invite.ID,
			Result:     "allowed",
		})
		_ = writeJSON(w, http.StatusOK, invite)
		return
	}
	invitesMu.Unlock()
	writeJSONError(w, http.StatusNotFound, "invite_not_found")
}

func invitesByTenant(tenantID string) []iamInvite {
	load := func() []iamInvite {
		out := make([]iamInvite, 0)
		for _, invite := range invites {
			if invite.TenantID == tenantID {
				out = append(out, invite)
			}
		}
		return out
	}
	invitesMu.RLock()
	out := load()
	invitesMu.RUnlock()
	if len(out) > 0 {
		return out
	}
	_ = reloadIAMStateFromPersistence()
	invitesMu.RLock()
	out = load()
	invitesMu.RUnlock()
	return out
}

func inviteByToken(token string) (iamInvite, bool) {
	hashedToken := hashToken(token)
	invitesMu.RLock()
	invite, ok := invites[token]
	if !ok {
		invite, ok = invites[hashedToken]
	}
	invitesMu.RUnlock()
	if ok {
		return invite, true
	}
	_ = reloadIAMStateFromPersistence()
	invitesMu.RLock()
	defer invitesMu.RUnlock()
	invite, ok = invites[token]
	if !ok {
		invite, ok = invites[hashedToken]
	}
	return invite, ok
}

func sanitizeInviteForStorage(invite iamInvite) iamInvite {
	invite.Token = ""
	invite.InviteLink = "#/accept-invite"
	return invite
}

func sanitizeInviteForRead(invite iamInvite) iamInvite {
	invite.Token = ""
	invite.InviteLink = "#/accept-invite"
	return invite
}

func newInviteToken() string {
	var raw [24]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "invite-fallback-token"
	}
	return hex.EncodeToString(raw[:])
}
