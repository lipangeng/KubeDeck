package api

import (
	"strings"
	"time"
)

var rolePermissions = map[string][]string{
	"owner": {
		"*",
	},
	"admin": {
		"*",
	},
	"viewer": {
		"iam:read",
		"tenant:read",
		"menu:read",
		"resource:read",
		"cluster:switch",
	},
}

func sessionHasAnyPermission(session authSession, required []string) bool {
	granted := sessionPermissionSet(session)
	if _, ok := granted["*"]; ok {
		return true
	}
	for _, permission := range required {
		if _, ok := granted[strings.ToLower(strings.TrimSpace(permission))]; ok {
			return true
		}
	}
	return false
}

func sessionPermissionSet(session authSession) map[string]struct{} {
	granted := map[string]struct{}{}
	for _, role := range session.User.Roles {
		addRolePermissions(granted, role)
	}
	now := time.Now().UTC()
	for _, permission := range membershipPermissions(session.ActiveTenantID, session.User.ID, now) {
		code := strings.ToLower(strings.TrimSpace(permission))
		if code == "" {
			continue
		}
		granted[code] = struct{}{}
	}
	return granted
}

func rolesHavePermission(roles []string, required string) bool {
	target := strings.ToLower(strings.TrimSpace(required))
	if target == "" {
		return true
	}
	granted := map[string]struct{}{}
	for _, role := range roles {
		addRolePermissions(granted, role)
	}
	if _, ok := granted[target]; ok {
		return true
	}
	if _, ok := granted["*"]; ok {
		return true
	}
	return false
}

func addRolePermissions(granted map[string]struct{}, role string) {
	roleKey := strings.ToLower(strings.TrimSpace(role))
	if roleKey == "" {
		return
	}
	granted[roleKey] = struct{}{}
	for _, permission := range rolePermissions[roleKey] {
		code := strings.ToLower(strings.TrimSpace(permission))
		if code == "" {
			continue
		}
		granted[code] = struct{}{}
	}
}

func membershipPermissions(tenantID string, userID string, now time.Time) []string {
	load := func() []string {
		membershipGroupIDs := make([]string, 0, 4)
		iamMembershipsMu.RLock()
		for _, membership := range iamMemberships {
			if membership.TenantID != tenantID || !strings.EqualFold(membership.UserID, userID) {
				continue
			}
			if !membershipActiveAt(membership, now) {
				continue
			}
			membershipGroupIDs = append(membershipGroupIDs, membership.GroupIDs...)
		}
		iamMembershipsMu.RUnlock()
		if len(membershipGroupIDs) == 0 {
			return nil
		}

		permissions := make([]string, 0, len(membershipGroupIDs))
		iamGroupsMu.RLock()
		for _, groupID := range membershipGroupIDs {
			group, ok := iamGroups[groupID]
			if !ok || group.TenantID != tenantID {
				continue
			}
			permissions = append(permissions, group.Permissions...)
		}
		iamGroupsMu.RUnlock()
		return permissions
	}

	permissions := load()
	if len(permissions) > 0 {
		return permissions
	}
	_ = reloadIAMStateFromPersistence()
	return load()
}

func membershipActiveAt(membership iamMembership, now time.Time) bool {
	if !membership.EffectiveFrom.IsZero() && now.Before(membership.EffectiveFrom) {
		return false
	}
	if membership.EffectiveUntil != nil && !now.Before(*membership.EffectiveUntil) {
		return false
	}
	return true
}
