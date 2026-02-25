# IAM Foundation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Deliver production-usable Phase-1 IAM baseline for local login, multi-tenant RBAC, membership validity, invite onboarding, and audit trail.

**Architecture:** Extend backend microkernel with IAM domain services behind storage interfaces. Keep API router thin, move auth/authorization and audit write logic into dedicated middleware/service components. Use SQLite implementation first while keeping repository interfaces portable to MySQL/PostgreSQL.

**Tech Stack:** Go, net/http, existing backend internal modules (`api/auth/storage/core`), SQLite-first repositories, table-driven tests.

---

### Task 1: Define IAM Storage Contracts

**Files:**
- Modify: `backend/internal/storage/repo_interfaces.go`
- Test: `backend/internal/storage/factory_test.go`

**Step 1: Write failing storage contract tests**
- Add assertions for new repos: users/tenants/memberships/groups/permissions/sessions/invites/audit.

**Step 2: Run test to verify it fails**
- Run: `cd backend && go test ./internal/storage -run TestFactory -v`
- Expected: missing methods/fields compile failure.

**Step 3: Add minimal interface definitions + Store accessors**
- Extend `Store` with typed repo getters.
- Add stub repo structs/variables for all new repos.

**Step 4: Run test to verify it passes**
- Run: `cd backend && go test ./internal/storage -run TestFactory -v`

**Step 5: Commit**
- `git add backend/internal/storage/repo_interfaces.go backend/internal/storage/factory_test.go`
- `git commit -m "feat(storage): add iam repository interfaces"`

### Task 2: Add IAM Domain Types and Membership Validity Logic

**Files:**
- Create: `backend/internal/auth/tenant_membership.go`
- Modify: `backend/internal/auth/types.go`
- Test: `backend/internal/auth/rbac_eval_test.go`

**Step 1: Write failing tests for membership validity window**
- Cases: valid window, pre-effective, expired, no-expiry.

**Step 2: Run test to verify it fails**
- `cd backend && go test ./internal/auth -run Membership -v`

**Step 3: Implement validity helpers and integrate in access decision input model**
- Add `EffectiveFrom/EffectiveTo` and helper `IsMembershipActive(now)`.

**Step 4: Run test to verify it passes**
- `cd backend && go test ./internal/auth -run Membership -v`

**Step 5: Commit**
- `git add backend/internal/auth/tenant_membership.go backend/internal/auth/types.go backend/internal/auth/rbac_eval_test.go`
- `git commit -m "feat(auth): add tenant membership validity model"`

### Task 3: Add Auth Session APIs (login/me/logout/switch-tenant)

**Files:**
- Create: `backend/internal/api/auth_handler.go`
- Modify: `backend/internal/api/router.go`
- Test: `backend/internal/api/meta_handler_test.go`

**Step 1: Write failing API tests**
- Test routes and status contracts for:
  - `POST /api/auth/login` (tenant_code accepted)
  - `GET /api/auth/me`
  - `POST /api/auth/logout`
  - `POST /api/auth/switch-tenant`

**Step 2: Run tests to verify failure**
- `cd backend && go test ./internal/api -run Auth -v`

**Step 3: Implement minimal handlers + request/response DTOs**
- Stub repository-backed flows with explicit errors.
- Validate `tenant_code` branch and membership validity checks.

**Step 4: Run tests to verify pass**
- `cd backend && go test ./internal/api -run Auth -v`

**Step 5: Commit**
- `git add backend/internal/api/auth_handler.go backend/internal/api/router.go backend/internal/api/meta_handler_test.go`
- `git commit -m "feat(api): add auth session endpoints"`

### Task 4: Add RBAC Permission Dictionary and Group Binding APIs

**Files:**
- Create: `backend/internal/api/iam_handler.go`
- Modify: `backend/internal/api/router.go`
- Test: `backend/internal/api/meta_handler_test.go`

**Step 1: Write failing tests for IAM APIs**
- Cover group CRUD, permission list, group-permission replace, membership-group replace.

**Step 2: Run test to verify failure**
- `cd backend && go test ./internal/api -run IAM -v`

**Step 3: Implement minimal APIs and validation**
- Enforce tenant scoping and permission checks (`iam:read/write`).

**Step 4: Run tests to verify pass**
- `cd backend && go test ./internal/api -run IAM -v`

**Step 5: Commit**
- `git add backend/internal/api/iam_handler.go backend/internal/api/router.go backend/internal/api/meta_handler_test.go`
- `git commit -m "feat(api): add iam management endpoints"`

### Task 5: Add Invite and Accept-Invite APIs with Notification Abstraction

**Files:**
- Create: `backend/internal/core/notification/provider.go`
- Create: `backend/internal/core/notification/provider_email_stub.go`
- Modify: `backend/internal/api/iam_handler.go`
- Modify: `backend/internal/api/auth_handler.go`
- Test: `backend/internal/api/meta_handler_test.go`

**Step 1: Write failing tests for invite flows**
- Create invite with email/phone validation.
- Accept invite success/failure/expired scenarios.

**Step 2: Run tests to verify failure**
- `cd backend && go test ./internal/api -run Invite -v`

**Step 3: Implement invite endpoints + notification provider interface usage**
- Email provider called when email exists.
- SMS path logged as stub.

**Step 4: Run tests to verify pass**
- `cd backend && go test ./internal/api -run Invite -v`

**Step 5: Commit**
- `git add backend/internal/core/notification backend/internal/api/iam_handler.go backend/internal/api/auth_handler.go backend/internal/api/meta_handler_test.go`
- `git commit -m "feat(iam): add invite onboarding and notification abstraction"`

### Task 6: Add Audit Event Writer and Integrate Critical Paths

**Files:**
- Create: `backend/internal/core/audit/writer.go`
- Modify: `backend/internal/api/auth_handler.go`
- Modify: `backend/internal/api/iam_handler.go`
- Modify: `backend/internal/api/resource_handler.go`
- Test: `backend/internal/api/meta_handler_test.go`

**Step 1: Write failing tests for audit side effects**
- Assert auth/login and IAM mutate endpoints trigger audit write attempts.

**Step 2: Run tests to verify failure**
- `cd backend && go test ./internal/api -run Audit -v`

**Step 3: Implement writer interface and best-effort integration**
- Ensure business response is not blocked by audit write errors.

**Step 4: Run tests to verify pass**
- `cd backend && go test ./internal/api -run Audit -v`

**Step 5: Commit**
- `git add backend/internal/core/audit backend/internal/api/auth_handler.go backend/internal/api/iam_handler.go backend/internal/api/resource_handler.go backend/internal/api/meta_handler_test.go`
- `git commit -m "feat(audit): add structured audit event pipeline"`

### Task 7: End-to-End API Verification + Docs Sync

**Files:**
- Modify: `docs/PLAN.md`
- Modify: `README.md`
- Test: `backend/internal/api/meta_handler_test.go`

**Step 1: Add integration-style test cases for full auth/tenant/iam path**
- login(tenant_code) -> me -> switch-tenant -> iam mutation authorization.

**Step 2: Run full backend test suite**
- `cd backend && go test ./...`

**Step 3: Run backend build verification**
- `cd backend && go build ./...`

**Step 4: Update docs for new APIs and IAM scope**
- Add endpoint list and membership validity behavior.

**Step 5: Commit**
- `git add docs/PLAN.md README.md backend/internal/api/meta_handler_test.go`
- `git commit -m "docs: sync iam baseline APIs and tenancy semantics"`

### Task 8: Frontend IAM Bootstrap (MVP)

**Files:**
- Create: `frontend/src/sdk/authApi.ts`
- Modify: `frontend/src/sdk/types.ts`
- Modify: `frontend/src/App.tsx`
- Test: `frontend/src/App.test.tsx`

**Step 1: Write failing tests for login bootstrap and tenant switch UX hooks**
- Include auth state loading and unauthorized fallback.

**Step 2: Run frontend tests to verify failure**
- `cd frontend && npm test -- --run`

**Step 3: Implement minimal API client + state integration**
- Read `me`, expose active tenant, wire switch-tenant action.

**Step 4: Run frontend test and build**
- `cd frontend && npm test -- --run`
- `cd frontend && npm run build`

**Step 5: Commit**
- `git add frontend/src/sdk/authApi.ts frontend/src/sdk/types.ts frontend/src/App.tsx frontend/src/App.test.tsx`
- `git commit -m "feat(frontend): add iam auth bootstrap and tenant context"`

### Task 9: Final Verification Gate

**Files:**
- No code changes required unless failures appear.

**Step 1: Run full checks**
- `cd backend && go test ./... && go build ./...`
- `cd frontend && npm test -- --run && npm run build`

**Step 2: Validate branch hygiene**
- `git status --short`

**Step 3: Commit if final fixups are needed**
- `git add ...`
- `git commit -m "chore: final verification fixups"`

