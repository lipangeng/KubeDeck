# KubeDeck Microkernel Baseline Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the Phase 1 baseline skeleton for KubeDeck with frontend/backend microkernel structure, plugin templates, registry contracts, and minimal stub APIs.

**Architecture:** Use contract-first structure: backend core is authority for metadata/registry/auth/storage contracts; frontend shell consumes metadata and composes menus/routes. All business capability is represented as built-in plugins. Default storage uses sqlite with swappable mysql/postgres drivers through repository interfaces.

**Tech Stack:** Go 1.22+, net/http, Kubernetes client-go (stub phase), Vite + TypeScript + MUI, Vitest, npm workspaces.

---

### Task 1: Scaffold backend module and entrypoint

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/kubedeck/main.go`
- Create: `backend/internal/core/app.go`
- Test: `backend/internal/core/app_test.go`

**Step 1: Write the failing test**

```go
package core

import "testing"

func TestNewApp(t *testing.T) {
    app := NewApp()
    if app == nil {
        t.Fatal("expected app instance")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/core -run TestNewApp -v`
Expected: FAIL with undefined `NewApp`.

**Step 3: Write minimal implementation**

```go
type App struct{}

func NewApp() *App { return &App{} }
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/core -run TestNewApp -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add backend/go.mod backend/cmd/kubedeck/main.go backend/internal/core/app.go backend/internal/core/app_test.go
git commit -m "build(backend): scaffold service entrypoint"
```

### Task 2: Add backend plugin SDK contracts

**Files:**
- Create: `backend/pkg/sdk/plugin.go`
- Create: `backend/internal/plugins/manager.go`
- Test: `backend/internal/plugins/manager_test.go`

**Step 1: Write the failing test**

```go
func TestManagerRegistersPlugin(t *testing.T) {
    m := NewManager()
    if err := m.Register(mockPlugin{id: "p1"}); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/plugins -run TestManagerRegistersPlugin -v`
Expected: FAIL with undefined manager types.

**Step 3: Write minimal implementation**

Implement `Plugin` interface and a `Manager` with in-memory map + duplicate id check.

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/plugins -run TestManagerRegistersPlugin -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add backend/pkg/sdk/plugin.go backend/internal/plugins/manager.go backend/internal/plugins/manager_test.go
git commit -m "feat(backend): add plugin manager and sdk contract"
```

### Task 3: Define registry models and snapshot builder stub

**Files:**
- Create: `backend/internal/registry/types.go`
- Create: `backend/internal/registry/builder.go`
- Test: `backend/internal/registry/builder_test.go`

**Step 1: Write the failing test**

Add test that builds snapshot from system + dynamic inputs and validates output lengths and source tags.

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/registry -run TestBuildSnapshot -v`
Expected: FAIL with missing types/functions.

**Step 3: Write minimal implementation**

Create structs for `ResourceType`, `PageMeta`, `SlotMeta`, `MenuItem`, `Snapshot`, and `BuildSnapshot` merge function.

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/registry -run TestBuildSnapshot -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add backend/internal/registry/types.go backend/internal/registry/builder.go backend/internal/registry/builder_test.go
git commit -m "feat(registry): add metadata models and snapshot builder"
```

### Task 4: Define storage abstraction with sqlite default and driver stubs

**Files:**
- Create: `backend/internal/storage/repo_interfaces.go`
- Create: `backend/internal/storage/factory.go`
- Create: `backend/internal/storage/driver_sqlite.go`
- Create: `backend/internal/storage/driver_mysql.go`
- Create: `backend/internal/storage/driver_postgres.go`
- Test: `backend/internal/storage/factory_test.go`

**Step 1: Write the failing test**

Add table-driven test: drivers `sqlite/mysql/postgres` are selectable; unknown driver returns error.

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/storage -run TestNewStore -v`
Expected: FAIL with missing factory.

**Step 3: Write minimal implementation**

Add repository interfaces and `NewStore(driver, dsn)` factory with sqlite concrete + mysql/postgres placeholder constructors.

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/storage -run TestNewStore -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add backend/internal/storage/*.go
git commit -m "feat(storage): add multi-driver abstraction with sqlite default"
```

### Task 5: Add auth contracts (local + oauth-ready)

**Files:**
- Create: `backend/internal/auth/types.go`
- Create: `backend/internal/auth/provider_local.go`
- Create: `backend/internal/auth/provider_oauth.go`
- Create: `backend/internal/auth/rbac_eval.go`
- Test: `backend/internal/auth/rbac_eval_test.go`

**Step 1: Write the failing test**

Add matrix test for allow/deny by cluster + namespace allowlist.

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/auth -run TestEvaluateAccess -v`
Expected: FAIL with undefined evaluator.

**Step 3: Write minimal implementation**

Implement `EvaluateAccess` and provider interfaces; local provider stub returns test user.

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/auth -run TestEvaluateAccess -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add backend/internal/auth/*.go
git commit -m "feat(auth): add local auth and oauth-ready contracts"
```

### Task 6: Expose minimal backend metadata/resource APIs

**Files:**
- Create: `backend/internal/api/router.go`
- Create: `backend/internal/api/meta_handler.go`
- Create: `backend/internal/api/resource_handler.go`
- Test: `backend/internal/api/meta_handler_test.go`

**Step 1: Write the failing test**

Create HTTP test for `GET /api/meta/registry?cluster=dev` expecting `200` and JSON object with `resourceTypes` key.

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/api -run TestRegistryEndpoint -v`
Expected: FAIL with route/handler missing.

**Step 3: Write minimal implementation**

Wire net/http mux and return stub JSON payloads for metadata endpoints and apply endpoint.

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/api -run TestRegistryEndpoint -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add backend/internal/api/*.go
git commit -m "feat(api): add metadata and resource stub endpoints"
```

### Task 7: Scaffold frontend shell workspace

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/shell/package.json`
- Create: `frontend/shell/tsconfig.json`
- Create: `frontend/shell/vite.config.ts`
- Create: `frontend/shell/src/main.tsx`
- Create: `frontend/shell/src/App.tsx`
- Test: `frontend/shell/src/App.test.tsx`

**Step 1: Write the failing test**

Create test asserting shell renders "KubeDeck" header.

**Step 2: Run test to verify it fails**

Run: `cd frontend/shell && npm run test -- App.test.tsx`
Expected: FAIL before app component exists.

**Step 3: Write minimal implementation**

Add Vite+TS app skeleton and basic app shell frame.

**Step 4: Run test to verify it passes**

Run: `cd frontend/shell && npm run test -- App.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add frontend/package.json frontend/shell
git commit -m "build(frontend): scaffold shell app workspace"
```

### Task 8: Add frontend plugin host and menu composer contracts

**Files:**
- Create: `frontend/shell/src/sdk/types.ts`
- Create: `frontend/shell/src/core/pluginHost.ts`
- Create: `frontend/shell/src/core/menuComposer.ts`
- Create: `frontend/shell/src/core/routeResolver.ts`
- Test: `frontend/shell/src/core/menuComposer.test.ts`

**Step 1: Write the failing test**

Test system/user/dynamic menus merge with order/visibility rules.

**Step 2: Run test to verify it fails**

Run: `cd frontend/shell && npm run test -- src/core/menuComposer.test.ts`
Expected: FAIL with missing composer.

**Step 3: Write minimal implementation**

Implement typed contracts and pure merge functions.

**Step 4: Run test to verify it passes**

Run: `cd frontend/shell && npm run test -- src/core/menuComposer.test.ts`
Expected: PASS.

**Step 5: Commit**

```bash
git add frontend/shell/src/sdk frontend/shell/src/core
git commit -m "feat(shell): add plugin host and menu composition contracts"
```

### Task 9: Add frontend state for cluster switch and namespace defaulting

**Files:**
- Create: `frontend/shell/src/state/clusterContext.ts`
- Create: `frontend/shell/src/state/namespaceFilter.ts`
- Test: `frontend/shell/src/state/namespaceFilter.test.ts`

**Step 1: Write the failing test**

Add tests for rules:
- list filter ns selected -> create dialog defaults to same ns
- list filter all -> fallback to last-used ns -> otherwise `default`

**Step 2: Run test to verify it fails**

Run: `cd frontend/shell && npm run test -- src/state/namespaceFilter.test.ts`
Expected: FAIL with undefined rule function.

**Step 3: Write minimal implementation**

Implement deterministic namespace default resolver and cluster-switch reset helper.

**Step 4: Run test to verify it passes**

Run: `cd frontend/shell && npm run test -- src/state/namespaceFilter.test.ts`
Expected: PASS.

**Step 5: Commit**

```bash
git add frontend/shell/src/state
git commit -m "feat(shell-state): add cluster and namespace context rules"
```

### Task 10: Add plugin templates (frontend/backend)

**Files:**
- Create: `plugins/templates/frontend-plugin-template/plugin.manifest.json`
- Create: `plugins/templates/frontend-plugin-template/README.md`
- Create: `plugins/templates/frontend-plugin-template/src/index.ts`
- Create: `plugins/templates/backend-plugin-template/plugin.manifest.json`
- Create: `plugins/templates/backend-plugin-template/README.md`
- Create: `plugins/templates/backend-plugin-template/src/index.go`
- Test: `backend/internal/plugins/template_manifest_test.go`

**Step 1: Write the failing test**

Add backend test validating template manifest has required fields (`pluginId`, `version`, `displayName`, `contributions`).

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/plugins -run TestTemplateManifestShape -v`
Expected: FAIL because template/manifests missing.

**Step 3: Write minimal implementation**

Add both template trees with manifest + README explaining replacement/slots/menu contributions.

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/plugins -run TestTemplateManifestShape -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add plugins/templates backend/internal/plugins/template_manifest_test.go
git commit -m "feat(templates): add frontend and backend plugin templates"
```

### Task 11: Publish SSOT PLAN and architecture links

**Files:**
- Create: `docs/PLAN.md`
- Modify: `README.md`
- Test: `docs/plans/2026-02-25-kubedeck-architecture-design.md`

**Step 1: Write the failing check**

Define a doc checklist in README requiring both files exist (`docs/PLAN.md`, architecture design plan).

**Step 2: Run check to verify it fails**

Run: `test -f docs/PLAN.md && echo ok`
Expected: command fails until PLAN is added.

**Step 3: Write minimal implementation**

Add SSOT PLAN content and README pointers to planning docs.

**Step 4: Run check to verify it passes**

Run: `test -f docs/PLAN.md && echo ok`
Expected: prints `ok`.

**Step 5: Commit**

```bash
git add docs/PLAN.md README.md docs/plans/2026-02-25-kubedeck-architecture-design.md
git commit -m "docs(plan): add ssot plan and architecture references"
```

### Task 12: Verification checkpoint before PR

**Files:**
- Modify: none (verification only)
- Test: backend and frontend test suites

**Step 1: Run backend tests**

Run: `cd backend && go test ./...`
Expected: all PASS.

**Step 2: Run frontend tests**

Run: `cd frontend/shell && npm test -- --run`
Expected: all PASS.

**Step 3: Run frontend build**

Run: `cd frontend/shell && npm run build`
Expected: build success.

**Step 4: Summarize branch deliverables**

Run:
- `git branch --show-current`
- `git log --oneline --decorate -n 12`
- `git status`

Expected:
- on feature branch
- clean working tree
- commit history aligned with tasks

**Step 5: Commit final verification note (if needed)**

```bash
git commit --allow-empty -m "chore(verify): record green baseline checks"
```

