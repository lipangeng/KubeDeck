# KubeDeck PLAN (SSOT)

This document is the single source of truth for scope, architecture direction, and execution rules.

## 1. Scope and Stack

KubeDeck is a multi-cluster, plugin-extensible Kubernetes web control plane.
- Frontend: Vite + TypeScript + MUI
- Backend: Go
- Architecture: frontend shell + backend core, both microkernel + plugin model

## 2. Non-Negotiable Principles

1. Microkernel-first: core provides framework capability; business is pluginized.
2. Feature-branch workflow only; no direct development on `main/master`.
3. Backend authorization is authoritative.
4. Built-in resources and CRDs share unified abstraction.
5. UI extension supports both replacement and slots.
6. Product supports i18n (ZH/EN).

## 3. Current Implementation Baseline (2026-02-25)

Completed and merged:
- Backend module skeleton (`api/auth/core/plugins/registry/storage/webui`)
- Multi-database IAM persistence (default sqlite, mysql/postgres implemented in storage repository)
- Metadata/resource stub APIs and health probes (`/api/healthz`, `/api/readyz`)
- Local auth session APIs (`login/me/logout/switch-tenant`) with `tenant_code` targeting
- Membership validity domain model (`effective_from/effective_to`) and validation hooks
- Tenant-scoped IAM APIs (group CRUD, permission binding, membership-group binding)
- Invite onboarding APIs (`create/list/accept`) with notification provider abstraction
- Structured audit pipeline and tenant-scoped audit events API
- Backend-sensitive write endpoint authz enforcement (`/api/resources/apply`)
- Frontend shell baseline with MUI
- Theme preference (`system/light/dark`) with persistence
- Sidebar menu composition and grouped rendering (`system/user/dynamic`)
- Reusable page shell components (`ListPageShell`, `DetailPageShell`)
- Plugin templates and manifest validation tests
- Single-executable backend mode (embedded static + optional `--static-dir` override)
- Runtime persistence flags (`--db-driver`, `--db-dsn`, `--disable-persist`)
- Storage regression tests for IAM persistence (dialect placeholders + sqlite round-trip)
- IAM repo-first fallback chain for auth/groups/memberships/invites (cache miss -> persistence reload)
- Route-policy auth middleware table with per-method permission requirements
- Group-aware authorization (membership-group permission inheritance) and default tenant groups (`tenant-owner/admin/viewer`)
- Versioned IAM schema migrations (`schema_migrations` + baseline migration runner)
- OAuth MVP endpoints (`/api/auth/oauth/url`, `/api/auth/oauth/callback`) with provider stub wiring
- Configurable notification provider for invites (stub/webhook via env)
- OIDC provider foundation via open-source libraries (`go-oidc` + `oauth2`) with state validation
- OIDC claim mapping configurability (subject/username/roles claims + role map/default role)
- OIDC role allowlist mode with optional strict deny-on-empty behavior

## 4. Next Priority (Phase 1 continuation)

- (Completed for IAM MVP) repository-first fallback chain for IAM auth/group/member/invite/users flows
- (Completed for IAM MVP) auth middleware abstraction and route policy table (instead of per-handler checks)
- (Completed for IAM MVP) production migration strategy baseline (runbook + migration runner foundation)
- Add OAuth provider integration for production IdPs and RBAC administration UI
- Add OAuth config diagnostic endpoint (`GET /api/auth/oauth/config`) for operations troubleshooting
  - Security boundary: return effective non-sensitive config only; never return raw `client_secret`
- Expand invite delivery providers (email production adapter + SMS adapter)

## 5. Canonical Documents

- Architecture design: `docs/plans/2026-02-25-kubedeck-architecture-design.md`
- Implementation plan: `docs/plans/2026-02-25-kubedeck-microkernel-baseline-implementation.md`
- Runtime persistence guide: `docs/runbooks/backend-persistence.md`
- IAM migration strategy (mysql/postgres): `docs/runbooks/iam-migrations-mysql-postgres.md`
- Notification provider guide: `docs/runbooks/notification-provider.md`
- OIDC provider guide: `docs/runbooks/oidc-provider.md`

## 6. Execution Rule

- One task per feature branch/worktree.
- Each task must pass relevant tests before PR.
- Integration branch (`main`) must remain runnable.
