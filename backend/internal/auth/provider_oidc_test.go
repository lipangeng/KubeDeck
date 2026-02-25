package auth

import "testing"

func TestParseOIDCRoleClaims(t *testing.T) {
	got := parseOIDCRoleClaims("realm_roles, groups , custom_roles")
	if len(got) != 3 {
		t.Fatalf("expected 3 role claims, got %v", got)
	}
	if got[0] != "realm_roles" || got[1] != "groups" || got[2] != "custom_roles" {
		t.Fatalf("unexpected role claims: %v", got)
	}
}

func TestParseOIDCRoleMap(t *testing.T) {
	got := parseOIDCRoleMap("platform-admin=admin, platform-owner=owner,readonly=viewer")
	if got["platform-admin"] != "admin" {
		t.Fatalf("expected platform-admin map to admin, got %q", got["platform-admin"])
	}
	if got["platform-owner"] != "owner" {
		t.Fatalf("expected platform-owner map to owner, got %q", got["platform-owner"])
	}
	if got["readonly"] != "viewer" {
		t.Fatalf("expected readonly map to viewer, got %q", got["readonly"])
	}
}

func TestExtractOIDCIdentityWithCustomClaimsAndRoleMap(t *testing.T) {
	claims := map[string]any{
		"uid":         "user-001",
		"displayName": "alice",
		"groups":      []any{"platform-admin", "team-ops", "platform-admin"},
	}
	config := oidcClaimConfig{
		subjectClaim:  "uid",
		usernameClaim: "displayName",
		roleClaims:    []string{"groups"},
		roleMap: map[string]string{
			"platform-admin": "admin",
		},
		defaultRole: "viewer",
	}
	userID, username, roles, err := extractOIDCIdentity(claims, config)
	if err != nil {
		t.Fatalf("extract identity: %v", err)
	}
	if userID != "user-001" {
		t.Fatalf("expected user id user-001, got %q", userID)
	}
	if username != "alice" {
		t.Fatalf("expected username alice, got %q", username)
	}
	if len(roles) != 2 {
		t.Fatalf("expected two mapped roles, got %v", roles)
	}
	if roles[0] != "admin" || roles[1] != "team-ops" {
		t.Fatalf("unexpected roles %v", roles)
	}
}

func TestExtractOIDCIdentityFallbacksToDefaultRole(t *testing.T) {
	claims := map[string]any{
		"sub": "user-002",
	}
	config := oidcClaimConfig{
		subjectClaim:  "sub",
		usernameClaim: "preferred_username",
		roleClaims:    []string{"roles", "groups"},
		roleMap:       map[string]string{},
		defaultRole:   "viewer",
	}
	userID, username, roles, err := extractOIDCIdentity(claims, config)
	if err != nil {
		t.Fatalf("extract identity: %v", err)
	}
	if userID != "user-002" {
		t.Fatalf("expected user id user-002, got %q", userID)
	}
	if username != "user-002" {
		t.Fatalf("expected username fallback to user id, got %q", username)
	}
	if len(roles) != 1 || roles[0] != "viewer" {
		t.Fatalf("expected default viewer role, got %v", roles)
	}
}

func TestExtractOIDCIdentityWhitelistFilter(t *testing.T) {
	claims := map[string]any{
		"sub":    "user-003",
		"groups": []any{"team-ops", "platform-admin"},
	}
	config := oidcClaimConfig{
		subjectClaim:       "sub",
		usernameClaim:      "preferred_username",
		roleClaims:         []string{"groups"},
		roleMap:            map[string]string{"platform-admin": "admin", "team-ops": "ops"},
		defaultRole:        "viewer",
		allowedRoles:       map[string]struct{}{"admin": {}},
		requireAllowedRole: false,
	}
	_, _, roles, err := extractOIDCIdentity(claims, config)
	if err != nil {
		t.Fatalf("extract identity: %v", err)
	}
	if len(roles) != 1 || roles[0] != "admin" {
		t.Fatalf("expected whitelisted admin role only, got %v", roles)
	}
}

func TestExtractOIDCIdentityWhitelistStrictModeDenied(t *testing.T) {
	claims := map[string]any{
		"sub":    "user-004",
		"groups": []any{"readonly"},
	}
	config := oidcClaimConfig{
		subjectClaim:       "sub",
		usernameClaim:      "preferred_username",
		roleClaims:         []string{"groups"},
		roleMap:            map[string]string{"readonly": "viewer"},
		defaultRole:        "viewer",
		allowedRoles:       map[string]struct{}{"admin": {}},
		requireAllowedRole: true,
	}
	_, _, _, err := extractOIDCIdentity(claims, config)
	if err == nil {
		t.Fatal("expected strict whitelist denial error")
	}
}

func TestParseOIDCAllowedRoles(t *testing.T) {
	allowed := parseOIDCAllowedRoles("admin, viewer, owner")
	if len(allowed) != 3 {
		t.Fatalf("expected three allowed roles, got %v", allowed)
	}
	if _, ok := allowed["admin"]; !ok {
		t.Fatalf("expected admin allowed role")
	}
	if _, ok := allowed["viewer"]; !ok {
		t.Fatalf("expected viewer allowed role")
	}
	if _, ok := allowed["owner"]; !ok {
		t.Fatalf("expected owner allowed role")
	}
}
