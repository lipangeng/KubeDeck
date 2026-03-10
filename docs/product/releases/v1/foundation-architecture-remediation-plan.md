# KubeDeck Foundation Architecture Remediation Plan

> **For agentic workers:** REQUIRED: Use `test-driven-development` before implementation and keep changes aligned with `docs/product/development-mode.md`. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the current architecture gaps between the implemented microkernel skeleton and the current product runtime model so V1 can resume on the correct foundation.

**Architecture:** The plan starts from the already-working kernel contracts and plugin discovery path, then adds the missing product runtime layers in order: composed menus, shared working context, shared resource-page shell, and finally the first real workflow on top of those systems. Each task stays architecture-first and must avoid rebuilding shell-only paths.

**Tech Stack:** Go backend, React + TypeScript frontend, Vite, MUI, Vitest, Go test

---

## 1. Scope

This plan only covers architecture remediation work required before feature expansion.

It does not include:

- AI features,
- broad UI polish,
- marketplace or plugin management,
- broad block-level page extension,
- or wide resource-domain expansion.

## 2. Current Gap Summary

The current codebase already has:

- frontend and backend kernel contracts,
- plugin discovery,
- backend kernel snapshot output,
- frontend runtime hydration,
- and minimal i18n copy access.

The missing product-runtime layers are:

1. composed menu runtime,
2. shared working context runtime,
3. resource-page shell runtime,
4. first workflow rebuilt on top of those systems.

## 3. File Structure

### Backend areas

- Modify: `backend/pkg/sdk/menu.go`
  - expand menu descriptor shape so it can represent blueprint or mount-driven menu composition results
- Modify: `backend/internal/plugins/menu_composer.go`
  - stop treating menus as a flat sorted list and start composing from blueprint and mount inputs
- Create: `backend/internal/plugins/menu_blueprint.go`
  - define the system-owned default menu skeleton
- Create: `backend/internal/plugins/menu_mounts.go`
  - convert built-in, CRD, and plugin capabilities into uniform menu mounts
- Create: `backend/internal/plugins/menu_resolution.go`
  - resolve availability states and final composed menu entries
- Modify: `backend/internal/plugins/kernel_snapshot.go`
  - expose the composed menu result shape
- Modify: `backend/internal/api/kernel_handler.go`
  - expose the new composed menu payload through kernel snapshot

### Frontend areas

- Modify: `frontend/src/kernel/contracts/menuContribution.ts`
  - align frontend menu types with the composed menu result instead of flat registrations
- Create: `frontend/src/kernel/runtime/menu/types.ts`
  - frontend menu runtime result types
- Create: `frontend/src/kernel/runtime/menu/resolveMenuState.ts`
  - translate backend menu composition into UI-ready navigation groups and states
- Modify: `frontend/src/kernel/runtime/composeKernelNavigation.ts`
  - stop sorting a flat list and instead render grouped composed navigation
- Modify: `frontend/src/App.tsx`
  - consume grouped menu results instead of flat buttons

### Shared context areas

- Create: `frontend/src/kernel/runtime/context/types.ts`
  - active cluster, namespace scope, workflow domain, resource identity, and continuity types
- Create: `frontend/src/kernel/runtime/context/reducer.ts`
  - context transitions for navigation, resource entry, and action start or finish
- Create: `frontend/src/kernel/runtime/context/selectors.ts`
  - UI selectors for active cluster, namespace scope, current domain, and current resource
- Modify: `frontend/src/kernel/runtime/KernelRuntimeContext.tsx`
  - host shared working context state together with kernel snapshot state

### Resource page areas

- Create: `frontend/src/kernel/resource-pages/types.ts`
  - resource capability, tab capability, tab replacement, and page takeover types
- Create: `frontend/src/kernel/resource-pages/resolveResourcePage.ts`
  - final page resolution for one resource identity
- Create: `frontend/src/kernel/resource-pages/ResourcePageShell.tsx`
  - the product-level shared resource-page shell
- Create: `frontend/src/kernel/resource-pages/DefaultOverviewTab.tsx`
  - default overview tab
- Create: `frontend/src/kernel/resource-pages/DefaultYamlTab.tsx`
  - default YAML tab
- Create: `frontend/src/kernel/resource-pages/tabs.ts`
  - default tab registration and ordering rules
- Modify: `frontend/src/components/page-shell/ResourcePageShell.tsx`
  - either narrow to generic layout-only use or remove once the product-level shell is in place

### I18n minimum-runtime areas

- Modify: `frontend/src/i18n/types.ts`
  - define locale state shape for product-level usage
- Modify: `frontend/src/i18n/copy.ts`
  - stop assuming only direct `'en'` access in runtime usage
- Create: `frontend/src/i18n/localeContext.tsx`
  - minimal product locale boundary
- Modify: `frontend/src/i18n/messages/en.ts`
  - add keys required by menu states, resource-page tabs, and workflow context UI

## 4. Chunk 1: Composed Menu Runtime

### Task 1: Define menu composition result contracts

**Files:**
- Modify: `backend/pkg/sdk/menu.go`
- Modify: `frontend/src/kernel/contracts/menuContribution.ts`
- Test: `backend/internal/plugins/menu_composer_test.go`
- Test: `frontend/src/kernel/runtime/composeKernelNavigation.test.ts`

- [ ] **Step 1: Write failing backend test for grouped composed menu descriptors**
- [ ] **Step 2: Run `mise exec go@1.25.1 -- go test ./backend/internal/plugins/...` and verify it fails for missing grouped menu fields**
- [ ] **Step 3: Write minimal backend descriptor changes for menu group, availability state, and fallback-entry support**
- [ ] **Step 4: Write failing frontend test for grouped navigation input**
- [ ] **Step 5: Run `cd frontend && npm test -- --run composeKernelNavigation` and verify it fails**
- [ ] **Step 6: Write minimal frontend type updates**
- [ ] **Step 7: Re-run affected backend and frontend tests until green**
- [ ] **Step 8: Commit**

### Task 2: Implement backend menu blueprint and mount resolution

**Files:**
- Create: `backend/internal/plugins/menu_blueprint.go`
- Create: `backend/internal/plugins/menu_mounts.go`
- Create: `backend/internal/plugins/menu_resolution.go`
- Modify: `backend/internal/plugins/menu_composer.go`
- Modify: `backend/internal/plugins/kernel_snapshot.go`
- Test: `backend/internal/plugins/menu_composer_test.go`

- [ ] **Step 1: Write failing tests for blueprint groups, mount placement, disabled-unavailable entries, and `CRDs` fallback**
- [ ] **Step 2: Run `mise exec go@1.25.1 -- go test ./backend/internal/plugins/...` and verify the failures match the missing composition rules**
- [ ] **Step 3: Implement the smallest backend blueprint, mount conversion, and resolution path**
- [ ] **Step 4: Re-run backend plugin tests until green**
- [ ] **Step 5: Commit**

### Task 3: Render composed menu results in the frontend shell

**Files:**
- Create: `frontend/src/kernel/runtime/menu/types.ts`
- Create: `frontend/src/kernel/runtime/menu/resolveMenuState.ts`
- Modify: `frontend/src/kernel/runtime/composeKernelNavigation.ts`
- Modify: `frontend/src/App.tsx`
- Test: `frontend/src/kernel/runtime/composeKernelNavigation.test.ts`
- Test: `frontend/src/App.test.tsx`

- [ ] **Step 1: Write failing tests for grouped rendering, disabled unavailable entries, and stable `CRDs` fallback display**
- [ ] **Step 2: Run `cd frontend && npm test -- --run` and verify the failures**
- [ ] **Step 3: Implement the smallest grouped navigation resolution and shell rendering**
- [ ] **Step 4: Re-run frontend tests until green**
- [ ] **Step 5: Commit**

## 5. Chunk 2: Shared Working Context Runtime

### Task 4: Reintroduce product-level working context

**Files:**
- Create: `frontend/src/kernel/runtime/context/types.ts`
- Create: `frontend/src/kernel/runtime/context/reducer.ts`
- Create: `frontend/src/kernel/runtime/context/selectors.ts`
- Modify: `frontend/src/kernel/runtime/KernelRuntimeContext.tsx`
- Test: `frontend/src/kernel/runtime/context/reducer.test.ts`

- [ ] **Step 1: Write failing reducer tests for cluster switch, namespace-scope update, workflow entry, and resource entry**
- [ ] **Step 2: Run `cd frontend && npm test -- --run reducer` and verify it fails**
- [ ] **Step 3: Implement the minimal shared context reducer and selectors**
- [ ] **Step 4: Re-run reducer tests until green**
- [ ] **Step 5: Commit**

### Task 5: Connect menu navigation to working context

**Files:**
- Modify: `frontend/src/kernel/runtime/KernelRuntimeContext.tsx`
- Modify: `frontend/src/App.tsx`
- Test: `frontend/src/App.test.tsx`

- [ ] **Step 1: Write failing app tests for active workflow-domain continuity and namespace continuity during menu navigation**
- [ ] **Step 2: Run `cd frontend && npm test -- --run App` and verify it fails for missing context transitions**
- [ ] **Step 3: Implement the smallest navigation-to-context connection**
- [ ] **Step 4: Re-run frontend app tests until green**
- [ ] **Step 5: Commit**

## 6. Chunk 3: Resource-Page Shell Runtime

### Task 6: Define resource-page capability types

**Files:**
- Create: `frontend/src/kernel/resource-pages/types.ts`
- Create: `frontend/src/kernel/resource-pages/tabs.ts`
- Test: `frontend/src/kernel/resource-pages/tabs.test.ts`

- [ ] **Step 1: Write failing tests for default `Overview` and `YAML` tab resolution**
- [ ] **Step 2: Run `cd frontend && npm test -- --run tabs` and verify it fails**
- [ ] **Step 3: Implement the minimal capability and default-tab definitions**
- [ ] **Step 4: Re-run tab tests until green**
- [ ] **Step 5: Commit**

### Task 7: Implement the shared resource-page shell

**Files:**
- Create: `frontend/src/kernel/resource-pages/ResourcePageShell.tsx`
- Create: `frontend/src/kernel/resource-pages/DefaultOverviewTab.tsx`
- Create: `frontend/src/kernel/resource-pages/DefaultYamlTab.tsx`
- Create: `frontend/src/kernel/resource-pages/resolveResourcePage.ts`
- Test: `frontend/src/kernel/resource-pages/ResourcePageShell.test.tsx`

- [ ] **Step 1: Write failing tests for shell rendering and default tab visibility**
- [ ] **Step 2: Run `cd frontend && npm test -- --run ResourcePageShell` and verify it fails**
- [ ] **Step 3: Implement the smallest shared shell and default tab path**
- [ ] **Step 4: Re-run resource-page tests until green**
- [ ] **Step 5: Commit**

### Task 8: Connect one resource path to the shared shell

**Files:**
- Modify: `frontend/src/kernel/builtins/pages/WorkloadsPage.tsx`
- Modify: `frontend/src/App.tsx`
- Test: `frontend/src/App.test.tsx`

- [ ] **Step 1: Write failing tests for entering one resource page from the workload list and seeing `Overview` plus `YAML`**
- [ ] **Step 2: Run `cd frontend && npm test -- --run App` and verify it fails**
- [ ] **Step 3: Implement the smallest resource-entry path through the shared shell**
- [ ] **Step 4: Re-run frontend app tests until green**
- [ ] **Step 5: Commit**

## 7. Chunk 4: Minimum I18n Runtime Completion

### Task 9: Add minimal product-level locale state

**Files:**
- Modify: `frontend/src/i18n/types.ts`
- Modify: `frontend/src/i18n/copy.ts`
- Create: `frontend/src/i18n/localeContext.tsx`
- Test: `frontend/src/i18n/copy.test.ts`

- [ ] **Step 1: Write failing tests for locale-context-driven copy access**
- [ ] **Step 2: Run `cd frontend && npm test -- --run copy` and verify it fails**
- [ ] **Step 3: Implement the smallest locale boundary without expanding into full i18n productization**
- [ ] **Step 4: Re-run i18n tests until green**
- [ ] **Step 5: Commit**

## 8. Chunk 5: Reconnect The First Real Workflow

### Task 10: Restore the first workflow on the corrected foundation

**Files:**
- Modify: `frontend/src/kernel/builtins/pages/HomepagePage.tsx`
- Modify: `frontend/src/kernel/builtins/pages/WorkloadsPage.tsx`
- Modify: `frontend/src/kernel/runtime/KernelRuntimeContext.tsx`
- Modify: `backend/internal/plugins/workload_provider.go`
- Modify: `backend/internal/plugins/action_executor.go`
- Test: `frontend/src/App.test.tsx`
- Test: `backend/internal/api/kernel_handler_test.go`

- [ ] **Step 1: Write failing integration tests for `Homepage -> Workloads -> Resource Page -> Action -> Result -> Return`**
- [ ] **Step 2: Run affected frontend and backend tests and verify the failures**
- [ ] **Step 3: Implement the smallest workflow recovery on top of the new menu and resource-page systems**
- [ ] **Step 4: Re-run affected frontend and backend tests until green**
- [ ] **Step 5: Commit**

## 9. Verification Commands

Use these commands at the end of each chunk:

- `cd frontend && npm test -- --run`
- `cd frontend && npm run build`
- `export GOPATH=/tmp/go GOMODCACHE=/tmp/go/pkg/mod GOCACHE=/tmp/go/build-cache GOTOOLCHAIN=local && mise exec go@1.25.1 -- go test ./...`
- `export GOPATH=/tmp/go GOMODCACHE=/tmp/go/pkg/mod GOCACHE=/tmp/go/build-cache GOTOOLCHAIN=local && mise exec go@1.25.1 -- go build ./...`

## 10. Exit Condition

This remediation plan is complete when:

- the menu runtime is composed rather than flat,
- shared working context is real rather than hard-coded,
- resource-page shell and default tabs exist,
- minimum i18n runtime state exists,
- and the first workflow runs on top of those layers instead of bypassing them.
