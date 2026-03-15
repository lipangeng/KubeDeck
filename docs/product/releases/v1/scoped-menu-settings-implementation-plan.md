# Scoped Menu Settings Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add first-wave scoped menu settings so KubeDeck can switch between work, system, and cluster menu spaces while supporting scoped `pin`, `hide`, and `reset` operations.

**Architecture:** Keep one dynamic menu system and extend it with scope-aware composition. The backend remains the authority for blueprint, mount, override, and resolved menu groups; the frontend only switches scope, loads resolved metadata for the active cluster, and sends minimal menu-setting mutations back through the preferences API.

**Tech Stack:** Go backend, React + TypeScript frontend, Vitest, Go test, MUI, existing kernel runtime and menu composition model.

---

## File Structure

### Backend

- Modify: `backend/internal/plugins/menu_model.go`
  Add first-class scope fields for menu overrides and ordering support.
- Modify: `backend/internal/plugins/menu_composer.go`
  Route composition through scoped override inputs.
- Modify: `backend/internal/plugins/menu_resolution.go`
  Parse `groupOrderOverrides` and `itemOrderOverrides` in addition to `pin` and `hide`.
- Modify: `backend/internal/plugins/kernel_snapshot.go`
  Emit scoped menu results for work, system, and cluster spaces.
- Modify: `backend/internal/api/kernel_handler.go`
  Accept `scope` in menu preference GET/PUT and snapshot requests.
- Modify: `backend/internal/api/router.go`
  Keep the menu preference API mounted on the same route family.
- Modify: `backend/internal/storage/repo_interfaces.go`
  Store menu overrides keyed by user, cluster, and scope.
- Test: `backend/internal/plugins/menu_composer_test.go`
  Lock scoped composition rules.
- Test: `backend/internal/api/kernel_handler_test.go`
  Lock scoped preference API behavior.

### Frontend

- Modify: `frontend/src/kernel/runtime/transport.ts`
  Extend remote contracts with scoped preference payloads and scope-aware menu metadata.
- Modify: `frontend/src/kernel/runtime/fetchKernelMetadata.ts`
  Send both cluster and active menu scope.
- Modify: `frontend/src/kernel/runtime/KernelRuntimeContext.tsx`
  Add current menu scope, scope switching, scoped metadata reload, and scoped preference mutation entry points.
- Modify: `frontend/src/App.tsx`
  Add the visible scope entry points:
  - top-right system settings
  - bottom-left cluster settings
  - back-to-work entry in config scopes
- Create: `frontend/src/kernel/runtime/updateMenuPreferences.ts`
  One focused transport helper for saving scoped menu settings.
- Create: `frontend/src/features/menu-settings/MenuSettingsPanel.tsx`
  First-wave settings UI for `pin`, `hide`, and `reset current scope`.
- Create: `frontend/src/features/menu-settings/MenuSettingsPanel.test.tsx`
  Verify the settings UI behavior in isolation.
- Modify: `frontend/src/App.test.tsx`
  Verify scope switching and menu refresh.
- Modify: `frontend/src/i18n/messages/en.ts`
  Add new copy keys for menu settings and scope switching.

### Documentation

- Modify: `docs/product/architecture/cluster-aware-menu-composition.md`
- Modify: `docs/product/architecture/cluster-aware-menu-composition.zh.md`
  Keep the architecture spec aligned with implementation details if behavior changes during execution.

## Chunk 1: Extend Scoped Override Contracts

### Task 1: Add failing backend tests for scoped override fields

**Files:**
- Modify: `backend/internal/plugins/menu_composer_test.go`

- [ ] **Step 1: Write the failing test**

Add tests that assert:
- `MenuOverride` accepts `scope`
- `groupOrderOverrides` is honored
- `itemOrderOverrides` is honored inside one group
- `system` and `cluster` scopes are treated as valid composition targets

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd backend && GOPATH=/tmp/go GOMODCACHE=/tmp/go/pkg/mod GOCACHE=/tmp/go/build-cache GOTOOLCHAIN=local mise exec go@1.25.1 -- go test ./internal/plugins -run TestComposeMenus
```

Expected: FAIL because the override model does not yet support the new scoped ordering fields.

- [ ] **Step 3: Write minimal implementation**

Update:
- `backend/internal/plugins/menu_model.go`
- `backend/internal/plugins/menu_resolution.go`

Add:
- scoped override shape
- `groupOrderOverrides`
- `itemOrderOverrides`
- ordering application after blueprint and mount resolution

- [ ] **Step 4: Run test to verify it passes**

Run the same command and confirm PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/plugins/menu_model.go backend/internal/plugins/menu_resolution.go backend/internal/plugins/menu_composer_test.go
git commit -m "feat(menu): extend scoped override model"
```

## Chunk 2: Add Scoped Preference Persistence And Snapshot Resolution

### Task 2: Add failing backend API tests for `scope`

**Files:**
- Modify: `backend/internal/api/kernel_handler_test.go`
- Modify: `backend/internal/storage/repo_interfaces.go`

- [ ] **Step 1: Write the failing test**

Add tests that assert:
- `GET /api/preferences/menu?scope=system`
- `PUT /api/preferences/menu?scope=cluster&cluster=prod-eu1`
- `GET /api/meta/kernel?scope=cluster&cluster=prod-eu1`

Each test should verify that the response reflects the requested scope instead of the previous hard-coded work-menu path.

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd backend && GOPATH=/tmp/go GOMODCACHE=/tmp/go/pkg/mod GOCACHE=/tmp/go/build-cache GOTOOLCHAIN=local mise exec go@1.25.1 -- go test ./internal/api -run 'TestKernelHandler(MenuPreferences|Snapshot)'
```

Expected: FAIL because the API does not yet understand scoped menu settings.

- [ ] **Step 3: Write minimal implementation**

Update:
- `backend/internal/api/kernel_handler.go`
- `backend/internal/storage/repo_interfaces.go`
- `backend/internal/plugins/kernel_snapshot.go`

Add:
- scope-aware preference lookup
- scope-aware snapshot composition
- per-scope persistence keyed by user + cluster + scope

- [ ] **Step 4: Run test to verify it passes**

Run the same command and confirm PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/api/kernel_handler.go backend/internal/api/kernel_handler_test.go backend/internal/plugins/kernel_snapshot.go backend/internal/storage/repo_interfaces.go
git commit -m "feat(menu): persist scoped menu preferences"
```

## Chunk 3: Add Frontend Scope Switching Runtime

### Task 3: Add failing frontend runtime tests for scope switching

**Files:**
- Modify: `frontend/src/App.test.tsx`
- Modify: `frontend/src/kernel/runtime/KernelRuntimeContext.tsx`

- [ ] **Step 1: Write the failing test**

Add tests that assert:
- clicking the top-right system settings entry switches the menu to system scope
- clicking the bottom-left cluster settings entry switches the menu to cluster scope
- a `Back to Work` entry returns to `work-*` scope
- metadata fetch includes the active scope

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd frontend && npm test -- --run src/App.test.tsx
```

Expected: FAIL because the runtime currently has no scoped menu-space switching.

- [ ] **Step 3: Write minimal implementation**

Update:
- `frontend/src/kernel/runtime/fetchKernelMetadata.ts`
- `frontend/src/kernel/runtime/KernelRuntimeContext.tsx`
- `frontend/src/App.tsx`

Add:
- current menu scope state
- scope switch actions
- top-right `System Settings`
- bottom-left `Cluster Settings`
- `Back to Work`

- [ ] **Step 4: Run test to verify it passes**

Run the same command and confirm PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/kernel/runtime/fetchKernelMetadata.ts frontend/src/kernel/runtime/KernelRuntimeContext.tsx frontend/src/App.tsx frontend/src/App.test.tsx
git commit -m "feat(menu): add scoped menu space switching"
```

## Chunk 4: Add First-Wave Menu Settings Panel

### Task 4: Add failing UI tests for `pin`, `hide`, and `reset`

**Files:**
- Create: `frontend/src/features/menu-settings/MenuSettingsPanel.test.tsx`
- Create: `frontend/src/features/menu-settings/MenuSettingsPanel.tsx`
- Create: `frontend/src/kernel/runtime/updateMenuPreferences.ts`
- Modify: `frontend/src/i18n/messages/en.ts`

- [ ] **Step 1: Write the failing test**

Add tests that assert:
- the panel renders the current resolved menu for the selected scope
- `Pin` sends an updated scoped override payload
- `Hide` sends an updated scoped override payload
- `Reset current scope` clears the scoped override payload

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd frontend && npm test -- --run src/features/menu-settings/MenuSettingsPanel.test.tsx
```

Expected: FAIL because the settings panel and update transport do not yet exist.

- [ ] **Step 3: Write minimal implementation**

Add:
- `updateMenuPreferences.ts`
- `MenuSettingsPanel.tsx`
- new copy keys in `en.ts`

Wire the panel into system and cluster configuration scopes only.

- [ ] **Step 4: Run test to verify it passes**

Run the same command and confirm PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/menu-settings/MenuSettingsPanel.tsx frontend/src/features/menu-settings/MenuSettingsPanel.test.tsx frontend/src/kernel/runtime/updateMenuPreferences.ts frontend/src/i18n/messages/en.ts
git commit -m "feat(menu): add first scoped menu settings panel"
```

## Chunk 5: Full Verification And Documentation Sync

### Task 5: Verify and sync docs

**Files:**
- Modify: `docs/product/architecture/cluster-aware-menu-composition.md`
- Modify: `docs/product/architecture/cluster-aware-menu-composition.zh.md`

- [ ] **Step 1: Run backend verification**

Run:

```bash
cd backend && GOPATH=/tmp/go GOMODCACHE=/tmp/go/pkg/mod GOCACHE=/tmp/go/build-cache GOTOOLCHAIN=local mise exec go@1.25.1 -- go test ./... && mise exec go@1.25.1 -- go build ./...
```

Expected: PASS.

- [ ] **Step 2: Run frontend verification**

Run:

```bash
cd frontend && npm test -- --run && npm run build
```

Expected: PASS.

- [ ] **Step 3: Sync docs if behavior changed**

Update the architecture document if the implemented scope names or UI entry paths changed during execution.

- [ ] **Step 4: Commit**

```bash
git add docs/product/architecture/cluster-aware-menu-composition.md docs/product/architecture/cluster-aware-menu-composition.zh.md
git commit -m "docs(menu): sync scoped menu settings implementation"
```
