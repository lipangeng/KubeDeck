# IAM Migration Strategy (MySQL/PostgreSQL)

This runbook defines the production migration strategy for IAM tables when `kubedeck` runs on MySQL or PostgreSQL.

## Goals

- Keep SQLite as default zero-config runtime for local/dev.
- Ensure MySQL/PostgreSQL schema evolution is explicit, repeatable, and rollback-aware.
- Avoid breaking changes in auth/session/IAM/audit data during upgrades.

## Baseline Rules

- Use versioned SQL migrations under `backend/internal/storage/migrations/<dialect>/`.
- Track applied versions in `schema_migrations(version, applied_at)`.
- Every release must include:
  - `up` migration for new schema/data changes.
  - `down` migration when rollback is practical and safe.
- Do not auto-apply destructive migrations without manual confirmation in production.

## Execution Flow

1. Startup checks DB driver (`sqlite`/`mysql`/`postgres`).
2. For MySQL/PostgreSQL:
   - Load pending migrations ordered by version.
   - Run each migration in a transaction when supported by the dialect/statement type.
   - Record success in `schema_migrations`.
3. On failure:
   - Stop startup and log failed version + statement summary.
   - Keep service in not-ready state until migration is resolved.

## Rollout Guidance

- Stage first: run migrations against a production-like snapshot.
- Backup before production migration.
- Deploy in order:
  1. apply migration
  2. deploy app
  3. run IAM smoke tests (`login`, `me`, `groups`, `invites`, `audit/events`)

## Current Status

- Migration framework and versioned SQL files are planned but not yet implemented in runtime code.
- This document is the contract for the next implementation task.
