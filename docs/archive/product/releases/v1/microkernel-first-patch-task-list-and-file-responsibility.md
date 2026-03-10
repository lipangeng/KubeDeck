# KubeDeck V1 First Microkernel Patch Task List And File Responsibility

## 1. Purpose

This document turns the first microkernel patch code mapping into an execution-ready pre-development checklist.

It defines:
- the first patch task list
- and the file-level responsibility boundary for each planned file

## 2. First Patch Task List

The first microkernel patch should be executed in this order.

### Task 1: Define Frontend Kernel Contract Types

Goal:
- establish the canonical frontend contribution contract family

Files:
- `frontend/src/kernel/contracts/types.ts`
- `frontend/src/kernel/contracts/pageContribution.ts`
- `frontend/src/kernel/contracts/menuContribution.ts`
- `frontend/src/kernel/contracts/actionContribution.ts`
- `frontend/src/kernel/contracts/slotContribution.ts`

Done when:
- page/menu/action/slot contribution shapes are explicit
- transport DTOs are no longer the default place for kernel contract design

### Task 2: Define Backend Capability Contract Types

Goal:
- establish the canonical backend capability contract family

Files:
- `backend/pkg/sdk/capability.go`
- `backend/pkg/sdk/menu.go`
- `backend/pkg/sdk/page.go`
- `backend/pkg/sdk/action.go`

Done when:
- backend capability registration is no longer identity-only
- capability metadata types exist outside API handlers

### Task 3: Register Built-In Frontend Contributions

Goal:
- give built-in workflow areas a contribution-style registration path

Files:
- `frontend/src/kernel/builtins/registerBuiltInPages.ts`
- `frontend/src/kernel/builtins/registerBuiltInMenus.ts`
- `frontend/src/kernel/builtins/registerBuiltInActions.ts`
- `frontend/src/kernel/builtins/registerBuiltInSlots.ts`

Done when:
- `Homepage`, `Workloads`, `Create`, and `Apply` can be described as built-in contributions

### Task 4: Register Built-In Backend Capabilities

Goal:
- make built-in capability metadata explicit on the backend side

Files:
- `backend/internal/core/builtins/homepage_capability.go`
- `backend/internal/core/builtins/workloads_capability.go`
- optional:
  - `backend/internal/core/builtins/create_action.go`
  - `backend/internal/core/builtins/apply_action.go`

Done when:
- built-in backend capabilities exist as explicit kernel inputs

### Task 5: Add Frontend Kernel Runtime Composition

Goal:
- provide one kernel-owned composition entry for shell consumption

Files:
- `frontend/src/kernel/runtime/kernelRegistry.ts`
- `frontend/src/kernel/runtime/composeKernelNavigation.ts`
- `frontend/src/kernel/runtime/resolveWorkflowActions.ts`
- `frontend/src/kernel/runtime/renderSlots.ts`

Done when:
- the shell has one clear place to consume built-in and future plugin contributions

### Task 6: Add Backend Capability Composition

Goal:
- move backend plugin logic beyond ID storage into capability composition

Files:
- `backend/internal/plugins/capability_registry.go`
- `backend/internal/plugins/menu_composer.go`
- `backend/internal/plugins/page_composer.go`
- `backend/internal/plugins/action_composer.go`

Done when:
- backend capability metadata can be composed independently of API handlers

### Task 7: Add Minimum I18n Copy Boundary

Goal:
- stop new product UI text from deepening inline hard-coded copy sprawl

Files:
- `frontend/src/i18n/copy.ts`
- `frontend/src/i18n/messages/en.ts`
- optional:
  - `frontend/src/i18n/types.ts`

Done when:
- new UI copy has one consistent access path

### Task 8: Add Narrow Compatibility Wiring

Goal:
- allow the new kernel baseline to coexist with current code without broad migration

Files that may receive narrow changes:
- `frontend/src/core/pluginHost.ts`
- `frontend/src/sdk/types.ts`
- `frontend/src/App.test.tsx`
- `backend/internal/plugins/manager.go`
- `backend/pkg/sdk/plugin.go`

Done when:
- new kernel structures can compile or be exercised without forcing full UI/API migration

## 3. File-Level Responsibility

### Frontend Kernel Contract Files

`frontend/src/kernel/contracts/types.ts`
- owns shared contract utility types
- does not own product page details

`frontend/src/kernel/contracts/pageContribution.ts`
- owns page contribution shape
- does not own page rendering logic

`frontend/src/kernel/contracts/menuContribution.ts`
- owns menu contribution shape
- does not own menu grouping policy implementation

`frontend/src/kernel/contracts/actionContribution.ts`
- owns action contribution shape
- does not own action form behavior

`frontend/src/kernel/contracts/slotContribution.ts`
- owns slot contribution shape
- does not own slot UI content

### Frontend Built-In Registration Files

`frontend/src/kernel/builtins/registerBuiltInPages.ts`
- registers built-in page contributions
- does not render pages

`frontend/src/kernel/builtins/registerBuiltInMenus.ts`
- registers built-in menu contributions
- does not compose final navigation

`frontend/src/kernel/builtins/registerBuiltInActions.ts`
- registers built-in workflow actions
- does not execute actions

`frontend/src/kernel/builtins/registerBuiltInSlots.ts`
- registers built-in slot contributions
- does not render slot contents directly

### Frontend Runtime Files

`frontend/src/kernel/runtime/kernelRegistry.ts`
- owns contribution aggregation and lookup
- does not own page-local state

`frontend/src/kernel/runtime/composeKernelNavigation.ts`
- owns navigation composition from registered contributions
- does not own sidebar rendering

`frontend/src/kernel/runtime/resolveWorkflowActions.ts`
- owns action lookup by workflow domain
- does not own action execution results

`frontend/src/kernel/runtime/renderSlots.ts`
- owns slot contribution resolution rules
- does not own slot content implementations

### Frontend I18n Files

`frontend/src/i18n/copy.ts`
- owns the minimum copy lookup/access pattern
- does not own component behavior

`frontend/src/i18n/messages/en.ts`
- owns the initial message catalog
- does not own translation logic beyond catalog shape

`frontend/src/i18n/types.ts`
- owns locale-related type definitions if introduced
- does not own preference persistence

### Backend Capability Contract Files

`backend/pkg/sdk/capability.go`
- owns backend capability registration shape
- does not own HTTP transport

`backend/pkg/sdk/menu.go`
- owns backend menu capability metadata shape
- does not own menu visibility execution

`backend/pkg/sdk/page.go`
- owns backend page capability metadata shape
- does not own page rendering

`backend/pkg/sdk/action.go`
- owns backend action capability metadata and execution contract shape
- does not own business action implementation

### Backend Composition Files

`backend/internal/plugins/capability_registry.go`
- owns capability aggregation and lookup
- does not own API output rendering

`backend/internal/plugins/menu_composer.go`
- owns backend menu composition
- does not own HTTP handlers

`backend/internal/plugins/page_composer.go`
- owns backend page metadata composition
- does not own frontend routes

`backend/internal/plugins/action_composer.go`
- owns backend action metadata composition
- does not own action execution internals

### Backend Built-In Capability Files

`backend/internal/core/builtins/homepage_capability.go`
- owns homepage capability registration
- does not own homepage UI

`backend/internal/core/builtins/workloads_capability.go`
- owns workloads capability registration
- does not own workload API response implementation

`backend/internal/core/builtins/create_action.go`
- owns built-in create action capability registration if introduced
- does not own generic kernel composition

`backend/internal/core/builtins/apply_action.go`
- owns built-in apply action capability registration if introduced
- does not own generic kernel composition

## 4. Existing Files That Should Stay Out Of Scope

These files should not become implementation centers for the first microkernel patch:

- `frontend/src/App.tsx`
- `frontend/src/pages/homepage/HomepageView.tsx`
- `frontend/src/pages/workloads/WorkloadsPage.tsx`
- `frontend/src/features/actions/ActionDrawer.tsx`
- `backend/internal/api/meta_handler.go`
- `backend/internal/api/resource_handler.go`

They may consume the new structures later, but they should not define them.

## 5. First-Patch Completion Check

The first patch is ready to implement when:
- the task order is accepted
- each file has one clear ownership statement
- existing page and API files are explicitly outside the first patch center
- and the patch still avoids full plugin platform and full i18n rollout scope
