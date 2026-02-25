package storage

import (
	"database/sql"
	"time"
)

type sqlMigration struct {
	version    string
	statements []string
}

func iamMigrationsForDialect(_ string) []sqlMigration {
	return []sqlMigration{
		{
			version: "20260225_001_iam_baseline",
			statements: []string{
				`CREATE TABLE IF NOT EXISTS iam_groups (
					id VARCHAR(255) PRIMARY KEY,
					tenant_id VARCHAR(255) NOT NULL,
					name VARCHAR(255) NOT NULL,
					description TEXT NOT NULL,
					permissions_json TEXT NOT NULL
				);`,
				`CREATE TABLE IF NOT EXISTS iam_memberships (
					id VARCHAR(255) PRIMARY KEY,
					tenant_id VARCHAR(255) NOT NULL,
					user_id VARCHAR(255) NOT NULL,
					user_label VARCHAR(255) NOT NULL,
					group_ids_json TEXT NOT NULL,
					effective_from VARCHAR(64) NOT NULL,
					effective_until VARCHAR(64)
				);`,
				`CREATE TABLE IF NOT EXISTS iam_invites (
					token VARCHAR(255) PRIMARY KEY,
					id VARCHAR(255) NOT NULL,
					tenant_id VARCHAR(255) NOT NULL,
					tenant_code VARCHAR(255) NOT NULL,
					invitee_email VARCHAR(255) NOT NULL,
					invitee_phone VARCHAR(255) NOT NULL,
					role_hint VARCHAR(255) NOT NULL,
					invite_link TEXT NOT NULL,
					created_at VARCHAR(64) NOT NULL,
					expires_at VARCHAR(64) NOT NULL,
					status VARCHAR(64) NOT NULL
				);`,
				`CREATE TABLE IF NOT EXISTS auth_sessions (
					token VARCHAR(255) PRIMARY KEY,
					user_json TEXT NOT NULL,
					available_json TEXT NOT NULL,
					active_tenant_id VARCHAR(255) NOT NULL
				);`,
				`CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_groups_tenant_name ON iam_groups(tenant_id, name);`,
				`CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_memberships_tenant_user ON iam_memberships(tenant_id, user_id);`,
			},
		},
	}
}

func applyMigrations(db *sql.DB, dialect string, migrations []sqlMigration) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at VARCHAR(64) NOT NULL
	);`); err != nil {
		return err
	}

	applied := map[string]struct{}{}
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			rows.Close()
			return err
		}
		applied[version] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, migration := range migrations {
		if _, ok := applied[migration.version]; ok {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}
		for _, statement := range migration.statements {
			if _, err := tx.Exec(statement); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		if _, err := tx.Exec(
			insertQuery(dialect, "schema_migrations", []string{"version", "applied_at"}),
			migration.version,
			time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
