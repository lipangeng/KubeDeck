package storage

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"path/filepath"
	"testing"
	"time"
)

func TestInsertQueryDialects(t *testing.T) {
	columns := []string{"id", "tenant_id", "name"}

	sqliteQuery := insertQuery("sqlite", "iam_groups", columns)
	if sqliteQuery != "INSERT INTO iam_groups(id, tenant_id, name) VALUES(?, ?, ?)" {
		t.Fatalf("unexpected sqlite query: %s", sqliteQuery)
	}

	mysqlQuery := insertQuery("mysql", "iam_groups", columns)
	if mysqlQuery != "INSERT INTO iam_groups(id, tenant_id, name) VALUES(?, ?, ?)" {
		t.Fatalf("unexpected mysql query: %s", mysqlQuery)
	}

	postgresQuery := insertQuery("pgx", "iam_groups", columns)
	if postgresQuery != "INSERT INTO iam_groups(id, tenant_id, name) VALUES($1, $2, $3)" {
		t.Fatalf("unexpected postgres query: %s", postgresQuery)
	}
}

func TestNewIAMPersistenceUnsupportedDriver(t *testing.T) {
	_, err := NewIAMPersistence("oracle", "")
	if err == nil {
		t.Fatal("expected unsupported driver error")
	}
}

func TestSQLiteIAMPersistenceRoundTrip(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "iam-roundtrip.sqlite")
	repo, err := NewIAMPersistence("sqlite", dbPath)
	if err != nil {
		t.Fatalf("new persistence: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	now := time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC)
	until := now.Add(24 * time.Hour)

	groups := []IAMGroupRecord{
		{
			ID:          "grp-admin",
			TenantID:    "tenant-dev",
			Name:        "admins",
			Description: "platform admins",
			Permissions: []string{"iam:read", "iam:write"},
		},
	}
	memberships := []IAMMembershipRecord{
		{
			ID:             "mbr-u1-tenant-dev",
			TenantID:       "tenant-dev",
			UserID:         "u-1",
			UserLabel:      "admin",
			GroupIDs:       []string{"grp-admin"},
			EffectiveFrom:  now,
			EffectiveUntil: &until,
		},
	}
	invites := []IAMInviteRecord{
		{
			Token:        "tok-1",
			ID:           "inv-1",
			TenantID:     "tenant-dev",
			TenantCode:   "dev",
			InviteeEmail: "user@example.com",
			InviteePhone: "",
			RoleHint:     "member",
			InviteLink:   "#/accept-invite?token=tok-1",
			CreatedAt:    now,
			ExpiresAt:    until,
			Status:       "pending",
		},
	}
	sessions := []AuthSessionRecord{
		{
			Token:          "sess-1",
			UserJSON:       `{"id":"u-1"}`,
			AvailableJSON:  `[{"id":"tenant-dev","code":"dev","name":"Development"}]`,
			ActiveTenantID: "tenant-dev",
			ExpiresAt:      until,
		},
	}

	if err := repo.ReplaceGroups(groups); err != nil {
		t.Fatalf("replace groups: %v", err)
	}
	if err := repo.ReplaceMemberships(memberships); err != nil {
		t.Fatalf("replace memberships: %v", err)
	}
	if err := repo.ReplaceInvites(invites); err != nil {
		t.Fatalf("replace invites: %v", err)
	}
	if err := repo.ReplaceSessions(sessions); err != nil {
		t.Fatalf("replace sessions: %v", err)
	}

	snapshot, err := repo.Load()
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	if len(snapshot.Groups) != 1 || snapshot.Groups[0].Name != "admins" {
		t.Fatalf("unexpected groups snapshot: %+v", snapshot.Groups)
	}
	if len(snapshot.Memberships) != 1 || snapshot.Memberships[0].UserID != "u-1" {
		t.Fatalf("unexpected memberships snapshot: %+v", snapshot.Memberships)
	}
	expectedInviteTokenHash := sha256.Sum256([]byte("tok-1"))
	if len(snapshot.Invites) != 1 || snapshot.Invites[0].Token != hex.EncodeToString(expectedInviteTokenHash[:]) {
		t.Fatalf("unexpected invites snapshot: %+v", snapshot.Invites)
	}
	expectedSessionTokenHash := sha256.Sum256([]byte("sess-1"))
	if len(snapshot.Sessions) != 1 || snapshot.Sessions[0].Token != hex.EncodeToString(expectedSessionTokenHash[:]) {
		t.Fatalf("unexpected sessions snapshot: %+v", snapshot.Sessions)
	}
	if !snapshot.Sessions[0].ExpiresAt.Equal(until.UTC()) {
		t.Fatalf("unexpected session expires_at: %+v", snapshot.Sessions[0].ExpiresAt)
	}
}

func TestSQLiteIAMPersistenceCreatesMigrationHistory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "iam-migrations.sqlite")
	repo, err := NewIAMPersistence("sqlite", dbPath)
	if err != nil {
		t.Fatalf("new persistence: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("query migration history: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected at least one migration record, got %d", count)
	}
}
