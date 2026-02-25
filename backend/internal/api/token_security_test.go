package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kubedeck/backend/internal/auth"
)

func TestCurrentValidSessionFromRequestRejectsExpiredSession(t *testing.T) {
	resetIAMPersistenceForTest()

	token := "expired-session-token"
	authSessionsMu.Lock()
	authSessions = map[string]authSession{
		token: {
			Token:          token,
			User:           auth.User{ID: "u-1", Username: "tester", Roles: []string{"admin"}, Memberships: []auth.TenantMembership{{TenantID: "tenant-dev", UserID: "u-1", EffectiveFrom: time.Now().UTC().Add(-time.Hour)}}},
			Available:      []tenantInfo{{ID: "tenant-dev", Code: "dev", Name: "Development"}},
			ActiveTenantID: "tenant-dev",
			ExpiresAt:      time.Now().UTC().Add(-time.Minute),
		},
	}
	authSessionsMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	_, _, ok, reason := currentValidSessionFromRequest(req)
	if ok {
		t.Fatalf("expected expired session to be rejected")
	}
	if reason != "session_expired" {
		t.Fatalf("expected session_expired reason, got %q", reason)
	}

	authSessionsMu.RLock()
	_, exists := authSessions[token]
	authSessionsMu.RUnlock()
	if exists {
		t.Fatalf("expected expired session to be removed from memory")
	}
}

func TestInviteByTokenMatchesHashedStorageKey(t *testing.T) {
	invite := iamInvite{
		ID:         "inv-hash",
		Token:      "",
		TenantID:   "tenant-dev",
		TenantCode: "dev",
		CreatedAt:  time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(time.Hour),
		Status:     "pending",
	}
	rawToken := "plain-invite-token"
	hashedToken := hashToken(rawToken)

	invitesMu.Lock()
	invites = map[string]iamInvite{hashedToken: invite}
	invitesMu.Unlock()

	found, ok := inviteByToken(rawToken)
	if !ok {
		t.Fatalf("expected invite lookup by raw token to match hashed storage key")
	}
	if found.ID != invite.ID {
		t.Fatalf("unexpected invite returned: %+v", found)
	}
}
