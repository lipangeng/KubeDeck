package storage

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	ExpiresAt      time.Time
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

func NewIAMPersistence(driver string, dsn string) (IAMPersistence, error) {
	driver = strings.ToLower(strings.TrimSpace(driver))
	switch driver {
	case "", "sqlite":
		return newSQLIAMPersistence("sqlite", dsn)
	case "mysql":
		return newSQLIAMPersistence("mysql", dsn)
	case "postgres":
		return newSQLIAMPersistence("pgx", dsn)
	default:
		return nil, fmt.Errorf("unsupported iam persistence driver %q", driver)
	}
}

type sqlIAMPersistence struct {
	db      *sql.DB
	dialect string
}

func newSQLIAMPersistence(driver string, dsn string) (IAMPersistence, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		switch driver {
		case "sqlite":
			dsn = "kubedeck.sqlite"
		case "mysql":
			dsn = "root:root@tcp(127.0.0.1:3306)/kubedeck?parseTime=true"
		case "pgx":
			dsn = "postgres://postgres:postgres@127.0.0.1:5432/kubedeck"
		}
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	p := &sqlIAMPersistence{db: db, dialect: driver}
	if err := p.ensureSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return p, nil
}

func (p *sqlIAMPersistence) Close() error {
	return p.db.Close()
}

func (p *sqlIAMPersistence) ensureSchema() error {
	return applyMigrations(p.db, p.dialect, iamMigrationsForDialect(p.dialect))
}

func (p *sqlIAMPersistence) Load() (IAMStateSnapshot, error) {
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

	sessionRows, err := p.db.Query(`SELECT token, user_json, available_json, active_tenant_id, expires_at FROM auth_sessions`)
	if err != nil {
		return out, err
	}
	for sessionRows.Next() {
		var record AuthSessionRecord
		var expiresAt string
		if err := sessionRows.Scan(&record.Token, &record.UserJSON, &record.AvailableJSON, &record.ActiveTenantID, &expiresAt); err != nil {
			sessionRows.Close()
			return out, err
		}
		trimmed := strings.TrimSpace(expiresAt)
		if trimmed != "" {
			parsedExpiresAt, err := time.Parse(time.RFC3339, trimmed)
			if err != nil {
				sessionRows.Close()
				return out, err
			}
			record.ExpiresAt = parsedExpiresAt.UTC()
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

func (p *sqlIAMPersistence) ReplaceGroups(records []IAMGroupRecord) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM iam_groups`); err != nil {
		_ = tx.Rollback()
		return err
	}
	query := insertQuery(p.dialect, "iam_groups", []string{"id", "tenant_id", "name", "description", "permissions_json"})
	for _, record := range records {
		permissionsJSON, _ := json.Marshal(record.Permissions)
		if _, err := tx.Exec(query, record.ID, record.TenantID, record.Name, record.Description, string(permissionsJSON)); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (p *sqlIAMPersistence) ReplaceMemberships(records []IAMMembershipRecord) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM iam_memberships`); err != nil {
		_ = tx.Rollback()
		return err
	}
	query := insertQuery(p.dialect, "iam_memberships", []string{"id", "tenant_id", "user_id", "user_label", "group_ids_json", "effective_from", "effective_until"})
	for _, record := range records {
		groupIDsJSON, _ := json.Marshal(record.GroupIDs)
		effectiveUntil := ""
		if record.EffectiveUntil != nil {
			effectiveUntil = record.EffectiveUntil.UTC().Format(time.RFC3339)
		}
		if _, err := tx.Exec(query, record.ID, record.TenantID, record.UserID, record.UserLabel, string(groupIDsJSON), record.EffectiveFrom.UTC().Format(time.RFC3339), effectiveUntil); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (p *sqlIAMPersistence) ReplaceInvites(records []IAMInviteRecord) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM iam_invites`); err != nil {
		_ = tx.Rollback()
		return err
	}
	query := insertQuery(p.dialect, "iam_invites", []string{"token", "id", "tenant_id", "tenant_code", "invitee_email", "invitee_phone", "role_hint", "invite_link", "created_at", "expires_at", "status"})
	for _, record := range records {
		if _, err := tx.Exec(query, tokenForStorage(record.Token), record.ID, record.TenantID, record.TenantCode, record.InviteeEmail, record.InviteePhone, record.RoleHint, record.InviteLink, record.CreatedAt.UTC().Format(time.RFC3339), record.ExpiresAt.UTC().Format(time.RFC3339), record.Status); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (p *sqlIAMPersistence) ReplaceSessions(records []AuthSessionRecord) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM auth_sessions`); err != nil {
		_ = tx.Rollback()
		return err
	}
	query := insertQuery(p.dialect, "auth_sessions", []string{"token", "user_json", "available_json", "active_tenant_id", "expires_at"})
	for _, record := range records {
		expiresAt := ""
		if !record.ExpiresAt.IsZero() {
			expiresAt = record.ExpiresAt.UTC().Format(time.RFC3339)
		}
		if _, err := tx.Exec(query, tokenForStorage(record.Token), record.UserJSON, record.AvailableJSON, record.ActiveTenantID, expiresAt); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func tokenForStorage(token string) string {
	trimmed := strings.TrimSpace(token)
	if isSHA256Hex(trimmed) {
		return trimmed
	}
	return hashToken(trimmed)
}

func isSHA256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

func insertQuery(dialect string, table string, columns []string) string {
	columnList := strings.Join(columns, ", ")
	placeholders := make([]string, 0, len(columns))
	for idx := range columns {
		if dialect == "pgx" {
			placeholders = append(placeholders, fmt.Sprintf("$%d", idx+1))
		} else {
			placeholders = append(placeholders, "?")
		}
	}
	return fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, columnList, strings.Join(placeholders, ", "))
}
