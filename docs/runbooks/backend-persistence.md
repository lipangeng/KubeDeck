# Backend Persistence Runbook

This document describes backend IAM persistence configuration for `kubedeck`.

## 1. Defaults

- Default driver: `sqlite`
- Default sqlite file: `kubedeck.sqlite` (working directory)
- Persistence can be disabled with flag or env.

## 2. Startup Flags

- `--db-driver=sqlite|mysql|postgres`
- `--db-dsn=<dsn-or-path>`
- `--disable-persist=true|false`

Example (sqlite):

```bash
go run ./cmd/kubedeck --db-driver=sqlite --db-dsn=./kubedeck.sqlite
```

Example (mysql):

```bash
go run ./cmd/kubedeck \
  --db-driver=mysql \
  --db-dsn='user:pass@tcp(127.0.0.1:3306)/kubedeck?parseTime=true'
```

Example (postgres):

```bash
go run ./cmd/kubedeck \
  --db-driver=postgres \
  --db-dsn='postgres://postgres:postgres@127.0.0.1:5432/kubedeck?sslmode=disable'
```

## 3. Environment Variables

Flags take precedence over env vars.

- `KUBEDECK_DB_DRIVER`
- `KUBEDECK_SQLITE_DSN`
- `KUBEDECK_IAM_PERSIST` (`0` disables persistence)

Example:

```bash
export KUBEDECK_DB_DRIVER=sqlite
export KUBEDECK_SQLITE_DSN=./kubedeck.sqlite
go run ./cmd/kubedeck
```

## 4. Current Scope

Persisted IAM domains:

- auth sessions
- iam groups
- iam memberships
- iam invites

Non-IAM Kubernetes resource data remains runtime state and should be backed by Kubernetes APIs.

## 5. Verification

```bash
cd backend && go test ./... && go build ./...
cd frontend && npm test -- --run && npm run build
```
