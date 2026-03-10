# KubeDeck V1 First Microkernel Patch Code Mapping Draft

## 1. Purpose

This document defines the first code landing plan for the microkernel architecture work.

It does not implement code. It maps the accepted kernel contribution contract and implementation mapping into concrete code ownership areas and patch scope.

## 2. First-Patch Principle

The first microkernel patch should establish contribution-oriented structure without attempting a full platform rewrite.

It should:
- introduce canonical kernel contract types,
- introduce built-in contribution registration,
- introduce a shell composition entry,
- and stop future feature work from deepening shell-only ownership.

It should not yet:
- implement dynamic plugin loading,
- redesign all backend APIs,
- or migrate every existing page fully into plugin-style modules.

## 3. Current Reality To Work Around

Today, the frontend still concentrates too much workflow ownership in the app shell, while plugin contracts are thin and mostly collection-oriented.

Today, the backend plugin layer is still identity-only and does not expose capability registration or execution-oriented contracts.

This means the first patch should not try to "finish plugins". It should create the architectural baseline that later patches can build on.

## 4. Frontend First-Patch Landing Areas

### 4.1 New Kernel Contract Files

Create a new frontend kernel contract area:

- `frontend/src/kernel/contracts/types.ts`
- `frontend/src/kernel/contracts/pageContribution.ts`
- `frontend/src/kernel/contracts/menuContribution.ts`
- `frontend/src/kernel/contracts/actionContribution.ts`
- `frontend/src/kernel/contracts/slotContribution.ts`

Purpose:
- define the canonical contribution contract family
- stop overloading `sdk/types.ts` with mixed transport and kernel concerns

### 4.2 Built-In Contribution Registration

Create a built-in capability registration area:

- `frontend/src/kernel/builtins/registerBuiltInPages.ts`
- `frontend/src/kernel/builtins/registerBuiltInMenus.ts`
- `frontend/src/kernel/builtins/registerBuiltInActions.ts`
- `frontend/src/kernel/builtins/registerBuiltInSlots.ts`

Purpose:
- model `Homepage`, `Workloads`, `Create`, and `Apply` as built-in contributions
- move the product toward capability-style ownership without requiring third-party plugins yet

### 4.3 Shell Composition Entry

Create a shell composition area:

- `frontend/src/kernel/runtime/kernelRegistry.ts`
- `frontend/src/kernel/runtime/composeKernelNavigation.ts`
- `frontend/src/kernel/runtime/resolveWorkflowActions.ts`
- `frontend/src/kernel/runtime/renderSlots.ts`

Purpose:
- provide one shell-owned composition entry
- keep registration separate from rendering and workflow state

### 4.4 I18n Minimum Boundary

Create the minimum copy-access boundary:

- `frontend/src/i18n/copy.ts`
- `frontend/src/i18n/messages/en.ts`

If the team wants an explicit locale type immediately, also add:

- `frontend/src/i18n/types.ts`

Purpose:
- stop new user-facing copy from continuing as arbitrary inline strings
- create a future-localizable access path without requiring full i18n rollout

## 5. Frontend Existing Files To Leave Mostly Untouched

The first patch should avoid broad rewrites in:

- `frontend/src/App.tsx`
- `frontend/src/pages/homepage/HomepageView.tsx`
- `frontend/src/pages/workloads/WorkloadsPage.tsx`
- `frontend/src/features/actions/ActionDrawer.tsx`
- `frontend/src/components/page-shell/*`

These files may receive minimal wiring later, but they should not be the place where kernel contract design is invented.

## 6. Frontend Existing Files That May Receive Narrow Changes

The first patch may update these files narrowly:

- `frontend/src/core/pluginHost.ts`
  Purpose:
  align the current host shape with the new contract family or route it behind the new kernel registry

- `frontend/src/sdk/types.ts`
  Purpose:
  reduce direct overlap between transport DTOs and kernel contribution types

- `frontend/src/App.test.tsx`
  Purpose:
  only if minimal composition wiring needs test adaptation

These changes should stay narrow and should not start broad UI migration.

## 7. Backend First-Patch Landing Areas

### 7.1 Backend Capability Contract Files

Create a backend kernel contract area:

- `backend/pkg/sdk/capability.go`
- `backend/pkg/sdk/menu.go`
- `backend/pkg/sdk/page.go`
- `backend/pkg/sdk/action.go`

Purpose:
- define capability registration beyond plugin identity
- establish the backend-side canonical contract family

### 7.2 Backend Kernel Composition Area

Create a backend composition area:

- `backend/internal/plugins/capability_registry.go`
- `backend/internal/plugins/menu_composer.go`
- `backend/internal/plugins/page_composer.go`
- `backend/internal/plugins/action_composer.go`

Purpose:
- move backend plugin logic beyond ID-only storage
- prepare cluster-aware capability metadata composition

### 7.3 Backend Built-In Capability Registration

Create a built-in capability registration area:

- `backend/internal/core/builtins/workloads_capability.go`
- `backend/internal/core/builtins/homepage_capability.go`

If action separation is preferred:

- `backend/internal/core/builtins/create_action.go`
- `backend/internal/core/builtins/apply_action.go`

Purpose:
- make built-in capabilities first-class kernel inputs

## 8. Backend Existing Files To Leave Mostly Untouched

The first patch should avoid broad changes in:

- `backend/internal/api/meta_handler.go`
- `backend/internal/api/resource_handler.go`
- `backend/internal/registry/*`
- `backend/internal/auth/*`

These areas may consume the new capability composition later, but the first patch should not mix capability contracts with full API behavior redesign.

## 9. Existing Files That Must Not Be The Design Source

The first patch must not treat these files as architecture truth:

- `frontend/src/App.tsx`
- `frontend/src/core/pluginHost.ts`
- `backend/internal/plugins/manager.go`
- `backend/pkg/sdk/plugin.go`

They may be adapted, wrapped, or replaced, but the new contract family should be driven by the accepted docs, not by preserving these current shapes.

## 10. Recommended Implementation Order

The first microkernel patch should be built in this order:

1. frontend kernel contract types
2. backend capability contract types
3. frontend built-in contribution registration
4. backend built-in capability registration
5. frontend shell composition entry
6. backend capability composition entry
7. minimum i18n copy-access boundary
8. narrow compatibility wiring in old host or shell files if required

## 11. First-Patch Test Landing

The first patch should add narrow tests for:

- contribution registration stability
- composition ordering
- built-in contribution exposure
- and minimum copy-access behavior if a lookup helper is introduced

Recommended new test areas:

- `frontend/src/kernel/runtime/kernelRegistry.test.ts`
- `frontend/src/kernel/runtime/composeKernelNavigation.test.ts`
- `backend/internal/plugins/capability_registry_test.go`
- `backend/internal/plugins/menu_composer_test.go`

## 12. First-Patch Success Criteria

The first microkernel patch is successful when:

- built-in workflow areas have a visible registration path outside the shell core
- kernel contribution contracts exist on both frontend and backend
- the shell has one clear composition entry
- new UI text no longer requires direct inline hard-coding as the only path
- and no broad plugin platform or full i18n rollout scope was mixed in
