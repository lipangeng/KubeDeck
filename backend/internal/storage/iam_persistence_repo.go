package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type IAMGroupRecord struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	Permissions []string
}

type IAMMembershipRecord struct {
	ID             string
	TenantID       string
	UserID         string
	UserLabel      string
	GroupIDs       []string
	EffectiveFrom  time.Time
	EffectiveUntil *time.Time
}

type IAMInviteRecord struct {
	Token        string
	ID           string
	TenantID     string
	TenantCode   string
	InviteeEmail string
	InviteePhone string
	RoleHint     string
	InviteLink   string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	Status       string
}

type AuthSessionRecord struct {
	Token          string
	UserJSON       string
	AvailableJSON  string
	ActiveTenantID string
}

type IAMStateSnapshot struct {
	Groups      []IAMGroupRecord
	Memberships []IAMMembershipRecord
	Invites     []IAMInviteRecord
	Sessions    []AuthSessionRecord
}

type IAMPersistence interface {
	Load() (IAMStateSnapshot, error)
	ReplaceGroups(records []IAMGroupRecord) error
	ReplaceMemberships(records []IAMMembershipRecord) error
	ReplaceInvites(records []IAMInviteRecord) error
	ReplaceSessions(records []AuthSessionRecord) error
	Close() error
}

type noopIAMPersistence struct{}

func (noopIAMPersistence) Load() (IAMStateSnapshot, error)                { return IAMStateSnapshot{}, nil }
func (noopIAMPersistence) ReplaceGroups([]IAMGroupRecord) error           { return nil }
func (noopIAMPersistence) ReplaceMemberships([]IAMMembershipRecord) error { return nil }
func (noopIAMPersistence) ReplaceInvites([]IAMInviteRecord) error         { return nil }
func (noopIAMPersistence) ReplaceSessions([]AuthSessionRecord) error      { return nil }
func (noopIAMPersistence) Close() error                                   { return nil }

func NewIAMPersistence(driver string, dsn string) (IAMPersistence, error) {
	driver = strings.ToLower(strings.TrimSpace(driver))
	if driver == "" || driver == "sqlite" {
		return newSQLiteIAMPersistence(dsn)
	}
	if driver == "mysql" || driver == "postgres" {
		return noopIAMPersistence{}, nil
	}
	return nil, fmt.Errorf("unsupported iam persistence driver %q", driver)
}

type sqliteIAMPersistence struct {
	db *sql.DB
}

func newSQLiteIAMPersistence(dsn string) (IAMPersistence, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		dsn = "kubedeck.sqlite"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	p := &sqliteIAMPersistence{db: db}
	if err := p.ensureSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return p, nil
}

func (p *sqliteIAMPersistence) Close() error {
	return p.db.Close()
}

func (p *sqliteIAMPersistence) ensureSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS iam_groups (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			permissions_json TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS iam_memberships (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			user_label TEXT NOT NULL,
			group_ids_json TEXT NOT NULL,
			effective_from TEXT NOT NULL,
			effective_until TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS iam_invites (
			token TEXT PRIMARY KEY,
			id TEXT NOT NULL,
			tenant_id TEXT NOT NULL,
			tenant_code TEXT NOT NULL,
			invitee_email TEXT NOT NULL,
			invitee_phone TEXT NOT NULL,
			role_hint TEXT NOT NULL,
			invite_link TEXT NOT NULL,
			created_at TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			status TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS auth_sessions (
			token TEXT PRIMARY KEY,
			user_json TEXT NOT NULL,
			available_json TEXT NOT NULL,
			active_tenant_id TEXT NOT NULL
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_groups_tenant_name ON iam_groups(tenant_id, name);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_memberships_tenant_user ON iam_memberships(tenant_id, user_id);`,
	}
	for _, stmt := range statements {
		if _, err := p.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (p *sqliteIAMPersistence) Load() (IAMStateSnapshot, error) {
	out := IAMStateSnapshot{}

	groupRows, err := p.db.Query(`SELECT id, tenant_id, name, description, permissions_json FROM iam_groups`)
	if err != nil {
		return out, err
	}
	for groupRows.Next() {
		var record IAMGroupRecord
		var permissionsJSON string
		if err := groupRows.Scan(&record.ID, &record.TenantID, &record.Name, &record.Description, &permissionsJSON); err != nil {
			groupRows.Close()
			return out, err
		}
		if permissionsJSON != "" {
			if err := json.Unmarshal([]byte(permissionsJSON), &record.Permissions); err != nil {
				groupRows.Close()
				return out, err
			}
		}
		out.Groups = append(out.Groups, record)
	}
	if err := groupRows.Err(); err != nil {
		groupRows.Close()
		return out, err
	}
	groupRows.Close()

	membershipRows, err := p.db.Query(`SELECT id, tenant_id, user_id, user_label, group_ids_json, effective_from, effective_until FROM iam_memberships`)
	if err != nil {
		return out, err
	}
	for membershipRows.Next() {
		var record IAMMembershipRecord
		var groupIDsJSON string
		var effectiveFrom string
		var effectiveUntil sql.NullString
		if err := membershipRows.Scan(&record.ID, &record.TenantID, &record.UserID, &record.UserLabel, &groupIDsJSON, &effectiveFrom, &effectiveUntil); err != nil {
			membershipRows.Close()
			return out, err
		}
		if groupIDsJSON != "" {
			if err := json.Unmarshal([]byte(groupIDsJSON), &record.GroupIDs); err != nil {
				membershipRows.Close()
				return out, err
			}
		}
		parsedFrom, err := time.Parse(time.RFC3339, effectiveFrom)
		if err != nil {
			membershipRows.Close()
			return out, err
		}
		record.EffectiveFrom = parsedFrom.UTC()
		if effectiveUntil.Valid && strings.TrimSpace(effectiveUntil.String) != "" {
			parsedUntil, err := time.Parse(time.RFC3339, effectiveUntil.String)
			if err != nil {
				membershipRows.Close()
				return out, err
			}
			parsedUntil = parsedUntil.UTC()
			record.EffectiveUntil = &parsedUntil
		}
		out.Memberships = append(out.Memberships, record)
	}
	if err := membershipRows.Err(); err != nil {
		membershipRows.Close()
		return out, err
	}
	membershipRows.Close()

	inviteRows, err := p.db.Query(`SELECT token, id, tenant_id, tenant_code, invitee_email, invitee_phone, role_hint, invite_link, created_at, expires_at, status FROM iam_invites`)
	if err != nil {
		return out, err
	}
	for inviteRows.Next() {
		var record IAMInviteRecord
		var createdAt string
		var expiresAt string
		if err := inviteRows.Scan(&record.Token, &record.ID, &record.TenantID, &record.TenantCode, &record.InviteeEmail, &record.InviteePhone, &record.RoleHint, &record.InviteLink, &createdAt, &expiresAt, &record.Status); err != nil {
			inviteRows.Close()
			return out, err
		}
		parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			inviteRows.Close()
			return out, err
		}
		parsedExpiresAt, err := time.Parse(time.RFC3339, expiresAt)
		if err != nil {
			inviteRows.Close()
			return out, err
		}
		record.CreatedAt = parsedCreatedAt.UTC()
		record.ExpiresAt = parsedExpiresAt.UTC()
		out.Invites = append(out.Invites, record)
	}
	if err := inviteRows.Err(); err != nil {
		inviteRows.Close()
		return out, err
	}
	inviteRows.Close()

	sessionRows, err := p.db.Query(`SELECT token, user_json, available_json, active_tenant_id FROM auth_sessions`)
	if err != nil {
		return out, err
	}
	for sessionRows.Next() {
		var record AuthSessionRecord
		if err := sessionRows.Scan(&record.Token, &record.UserJSON, &record.AvailableJSON, &record.ActiveTenantID); err != nil {
			sessionRows.Close()
			return out, err
		}
		out.Sessions = append(out.Sessions, record)
	}
	if err := sessionRows.Err(); err != nil {
		sessionRows.Close()
		return out, err
	}
	sessionRows.Close()

	return out, nil
}

func (p *sqliteIAMPersistence) ReplaceGroups(records []IAMGroupRecord) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM iam_groups`); err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, record := range records {
		permissionsJSON, _ := json.Marshal(record.Permissions)
		if _, err := tx.Exec(`INSERT INTO iam_groups(id, tenant_id, name, description, permissions_json) VALUES(?, ?, ?, ?, ?)`, record.ID, record.TenantID, record.Name, record.Description, string(permissionsJSON)); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (p *sqliteIAMPersistence) ReplaceMemberships(records []IAMMembershipRecord) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM iam_memberships`); err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, record := range records {
		groupIDsJSON, _ := json.Marshal(record.GroupIDs)
		effectiveUntil := ""
		if record.EffectiveUntil != nil {
			effectiveUntil = record.EffectiveUntil.UTC().Format(time.RFC3339)
		}
		if _, err := tx.Exec(`INSERT INTO iam_memberships(id, tenant_id, user_id, user_label, group_ids_json, effective_from, effective_until) VALUES(?, ?, ?, ?, ?, ?, ?)`, record.ID, record.TenantID, record.UserID, record.UserLabel, string(groupIDsJSON), record.EffectiveFrom.UTC().Format(time.RFC3339), effectiveUntil); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (p *sqliteIAMPersistence) ReplaceInvites(records []IAMInviteRecord) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM iam_invites`); err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, record := range records {
		if _, err := tx.Exec(`INSERT INTO iam_invites(token, id, tenant_id, tenant_code, invitee_email, invitee_phone, role_hint, invite_link, created_at, expires_at, status) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, record.Token, record.ID, record.TenantID, record.TenantCode, record.InviteeEmail, record.InviteePhone, record.RoleHint, record.InviteLink, record.CreatedAt.UTC().Format(time.RFC3339), record.ExpiresAt.UTC().Format(time.RFC3339), record.Status); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (p *sqliteIAMPersistence) ReplaceSessions(records []AuthSessionRecord) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM auth_sessions`); err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, record := range records {
		if _, err := tx.Exec(`INSERT INTO auth_sessions(token, user_json, available_json, active_tenant_id) VALUES(?, ?, ?, ?)`, record.Token, record.UserJSON, record.AvailableJSON, record.ActiveTenantID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
