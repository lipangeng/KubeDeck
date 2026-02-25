package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	iamPersistenceOnce sync.Once
	iamPersistenceDB   *sql.DB
	iamPersistenceErr  error
)

func resetIAMPersistenceForTest() {
	if iamPersistenceDB != nil {
		_ = iamPersistenceDB.Close()
	}
	iamPersistenceDB = nil
	iamPersistenceErr = nil
	iamPersistenceOnce = sync.Once{}
}

func ensureIAMPersistence() {
	iamPersistenceOnce.Do(func() {
		if !iamPersistenceEnabled() {
			return
		}
		dsn := strings.TrimSpace(os.Getenv("KUBEDECK_SQLITE_DSN"))
		if dsn == "" {
			dsn = "kubedeck.sqlite"
		}
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
			iamPersistenceErr = err
			return
		}
		if err := ensureIAMSchema(db); err != nil {
			_ = db.Close()
			iamPersistenceErr = err
			return
		}
		iamPersistenceDB = db
		if err := loadIAMPersistentState(db); err != nil {
			iamPersistenceErr = err
			return
		}
	})
	if iamPersistenceErr != nil {
		log.Printf("iam persistence disabled due to init error: %v", iamPersistenceErr)
	}
}

func iamPersistenceEnabled() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("KUBEDECK_IAM_PERSIST")), "0") {
		return false
	}
	argv0 := os.Args[0]
	if strings.HasSuffix(argv0, ".test") && !strings.EqualFold(strings.TrimSpace(os.Getenv("KUBEDECK_IAM_PERSIST_IN_TEST")), "1") {
		return false
	}
	return true
}

func ensureIAMSchema(db *sql.DB) error {
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
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func loadIAMPersistentState(db *sql.DB) error {
	loadGroups := map[string]iamGroup{}
	rows, err := db.Query(`SELECT id, tenant_id, name, description, permissions_json FROM iam_groups`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var g iamGroup
		var permissionsJSON string
		if err := rows.Scan(&g.ID, &g.TenantID, &g.Name, &g.Description, &permissionsJSON); err != nil {
			rows.Close()
			return err
		}
		if permissionsJSON != "" {
			if err := json.Unmarshal([]byte(permissionsJSON), &g.Permissions); err != nil {
				rows.Close()
				return err
			}
		}
		loadGroups[g.ID] = g
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	loadMemberships := map[string]iamMembership{}
	rows, err = db.Query(`SELECT id, tenant_id, user_id, user_label, group_ids_json, effective_from, effective_until FROM iam_memberships`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var m iamMembership
		var groupIDsJSON string
		var effectiveFrom string
		var effectiveUntil sql.NullString
		if err := rows.Scan(&m.ID, &m.TenantID, &m.UserID, &m.UserLabel, &groupIDsJSON, &effectiveFrom, &effectiveUntil); err != nil {
			rows.Close()
			return err
		}
		if groupIDsJSON != "" {
			if err := json.Unmarshal([]byte(groupIDsJSON), &m.GroupIDs); err != nil {
				rows.Close()
				return err
			}
		}
		parsedFrom, err := time.Parse(time.RFC3339, effectiveFrom)
		if err != nil {
			rows.Close()
			return err
		}
		m.EffectiveFrom = parsedFrom.UTC()
		if effectiveUntil.Valid && strings.TrimSpace(effectiveUntil.String) != "" {
			parsedUntil, err := time.Parse(time.RFC3339, effectiveUntil.String)
			if err != nil {
				rows.Close()
				return err
			}
			parsedUntil = parsedUntil.UTC()
			m.EffectiveUntil = &parsedUntil
		}
		loadMemberships[m.ID] = m
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	loadInvites := map[string]iamInvite{}
	rows, err = db.Query(`SELECT token, id, tenant_id, tenant_code, invitee_email, invitee_phone, role_hint, invite_link, created_at, expires_at, status FROM iam_invites`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var invite iamInvite
		var createdAt string
		var expiresAt string
		if err := rows.Scan(&invite.Token, &invite.ID, &invite.TenantID, &invite.TenantCode, &invite.InviteeEmail, &invite.InviteePhone, &invite.RoleHint, &invite.InviteLink, &createdAt, &expiresAt, &invite.Status); err != nil {
			rows.Close()
			return err
		}
		parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			rows.Close()
			return err
		}
		parsedExpiresAt, err := time.Parse(time.RFC3339, expiresAt)
		if err != nil {
			rows.Close()
			return err
		}
		invite.CreatedAt = parsedCreatedAt.UTC()
		invite.ExpiresAt = parsedExpiresAt.UTC()
		loadInvites[invite.Token] = invite
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	loadSessions := map[string]authSession{}
	rows, err = db.Query(`SELECT token, user_json, available_json, active_tenant_id FROM auth_sessions`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var token string
		var userJSON string
		var availableJSON string
		var activeTenantID string
		if err := rows.Scan(&token, &userJSON, &availableJSON, &activeTenantID); err != nil {
			rows.Close()
			return err
		}
		var session authSession
		if err := json.Unmarshal([]byte(userJSON), &session.User); err != nil {
			rows.Close()
			return err
		}
		if err := json.Unmarshal([]byte(availableJSON), &session.Available); err != nil {
			rows.Close()
			return err
		}
		session.Token = token
		session.ActiveTenantID = activeTenantID
		loadSessions[token] = session
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	iamGroupsMu.Lock()
	iamGroups = loadGroups
	iamGroupsMu.Unlock()

	iamMembershipsMu.Lock()
	iamMemberships = loadMemberships
	iamMembershipsMu.Unlock()

	invitesMu.Lock()
	invites = loadInvites
	invitesMu.Unlock()

	authSessionsMu.Lock()
	authSessions = loadSessions
	authSessionsMu.Unlock()

	return nil
}

func persistIAMGroups() {
	ensureIAMPersistence()
	if iamPersistenceDB == nil {
		return
	}
	iamGroupsMu.RLock()
	snapshot := make([]iamGroup, 0, len(iamGroups))
	for _, item := range iamGroups {
		snapshot = append(snapshot, item)
	}
	iamGroupsMu.RUnlock()

	tx, err := iamPersistenceDB.Begin()
	if err != nil {
		log.Printf("persist iam groups begin failed: %v", err)
		return
	}
	if _, err := tx.Exec(`DELETE FROM iam_groups`); err != nil {
		_ = tx.Rollback()
		log.Printf("persist iam groups clear failed: %v", err)
		return
	}
	for _, g := range snapshot {
		permissionsJSON, _ := json.Marshal(g.Permissions)
		if _, err := tx.Exec(
			`INSERT INTO iam_groups(id, tenant_id, name, description, permissions_json) VALUES(?, ?, ?, ?, ?)`,
			g.ID,
			g.TenantID,
			g.Name,
			g.Description,
			string(permissionsJSON),
		); err != nil {
			_ = tx.Rollback()
			log.Printf("persist iam groups insert failed: %v", err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		log.Printf("persist iam groups commit failed: %v", err)
	}
}

func persistIAMMemberships() {
	ensureIAMPersistence()
	if iamPersistenceDB == nil {
		return
	}
	iamMembershipsMu.RLock()
	snapshot := make([]iamMembership, 0, len(iamMemberships))
	for _, item := range iamMemberships {
		snapshot = append(snapshot, item)
	}
	iamMembershipsMu.RUnlock()

	tx, err := iamPersistenceDB.Begin()
	if err != nil {
		log.Printf("persist iam memberships begin failed: %v", err)
		return
	}
	if _, err := tx.Exec(`DELETE FROM iam_memberships`); err != nil {
		_ = tx.Rollback()
		log.Printf("persist iam memberships clear failed: %v", err)
		return
	}
	for _, m := range snapshot {
		groupIDsJSON, _ := json.Marshal(m.GroupIDs)
		effectiveUntil := ""
		if m.EffectiveUntil != nil {
			effectiveUntil = m.EffectiveUntil.UTC().Format(time.RFC3339)
		}
		if _, err := tx.Exec(
			`INSERT INTO iam_memberships(id, tenant_id, user_id, user_label, group_ids_json, effective_from, effective_until) VALUES(?, ?, ?, ?, ?, ?, ?)`,
			m.ID,
			m.TenantID,
			m.UserID,
			m.UserLabel,
			string(groupIDsJSON),
			m.EffectiveFrom.UTC().Format(time.RFC3339),
			effectiveUntil,
		); err != nil {
			_ = tx.Rollback()
			log.Printf("persist iam memberships insert failed: %v", err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		log.Printf("persist iam memberships commit failed: %v", err)
	}
}

func persistIAMInvites() {
	ensureIAMPersistence()
	if iamPersistenceDB == nil {
		return
	}
	invitesMu.RLock()
	snapshot := make([]iamInvite, 0, len(invites))
	for _, item := range invites {
		snapshot = append(snapshot, item)
	}
	invitesMu.RUnlock()

	tx, err := iamPersistenceDB.Begin()
	if err != nil {
		log.Printf("persist iam invites begin failed: %v", err)
		return
	}
	if _, err := tx.Exec(`DELETE FROM iam_invites`); err != nil {
		_ = tx.Rollback()
		log.Printf("persist iam invites clear failed: %v", err)
		return
	}
	for _, i := range snapshot {
		if _, err := tx.Exec(
			`INSERT INTO iam_invites(token, id, tenant_id, tenant_code, invitee_email, invitee_phone, role_hint, invite_link, created_at, expires_at, status) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			i.Token,
			i.ID,
			i.TenantID,
			i.TenantCode,
			i.InviteeEmail,
			i.InviteePhone,
			i.RoleHint,
			i.InviteLink,
			i.CreatedAt.UTC().Format(time.RFC3339),
			i.ExpiresAt.UTC().Format(time.RFC3339),
			i.Status,
		); err != nil {
			_ = tx.Rollback()
			log.Printf("persist iam invites insert failed: %v", err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		log.Printf("persist iam invites commit failed: %v", err)
	}
}

func persistAuthSessions() {
	ensureIAMPersistence()
	if iamPersistenceDB == nil {
		return
	}
	authSessionsMu.RLock()
	snapshot := make([]authSession, 0, len(authSessions))
	for _, item := range authSessions {
		snapshot = append(snapshot, item)
	}
	authSessionsMu.RUnlock()

	tx, err := iamPersistenceDB.Begin()
	if err != nil {
		log.Printf("persist auth sessions begin failed: %v", err)
		return
	}
	if _, err := tx.Exec(`DELETE FROM auth_sessions`); err != nil {
		_ = tx.Rollback()
		log.Printf("persist auth sessions clear failed: %v", err)
		return
	}
	for _, s := range snapshot {
		userJSON, _ := json.Marshal(s.User)
		availableJSON, _ := json.Marshal(s.Available)
		if _, err := tx.Exec(
			`INSERT INTO auth_sessions(token, user_json, available_json, active_tenant_id) VALUES(?, ?, ?, ?)`,
			s.Token,
			string(userJSON),
			string(availableJSON),
			s.ActiveTenantID,
		); err != nil {
			_ = tx.Rollback()
			log.Printf("persist auth sessions insert failed: %v", err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		log.Printf("persist auth sessions commit failed: %v", err)
	}
}
