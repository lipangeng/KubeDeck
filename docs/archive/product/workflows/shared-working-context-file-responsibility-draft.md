# KubeDeck Shared Working Context File Responsibility Draft

## 1. Purpose

This document defines the intended file-level responsibility split for implementing the shared working context.

It is intentionally derived from product documents and implementation mapping, not from the current historical code layout.

## 2. Design Rule

File responsibilities must follow workflow and context boundaries.

They must not be organized around:
- temporary convenience,
- one oversized `App` container,
- or historical mixing of layout, fetch logic, and state logic.

## 3. Recommended Frontend Responsibility Areas

The frontend should be split into these responsibility areas:
- app shell composition
- shared working context state
- workflow pages
- action surfaces
- SDK / API access
- page-local presentation state

## 4. Recommended File Groups

### 4.1 Shared Working Context State Files

Recommended group:
- `frontend/src/state/work-context/`

Recommended files:
- `clusterContext.ts`
- `namespaceContext.ts`
- `workflowContext.ts`
- `listContext.ts`
- `actionContext.ts`
- `events.ts`
- `selectors.ts`
- `types.ts`

Responsibility:
- define the shared state model
- define update entry points
- define selectors for page consumption

Must not contain:
- page layout code
- direct component rendering
- arbitrary fetch orchestration unrelated to context ownership

### 4.2 Workflow Page Files

Recommended group:
- `frontend/src/pages/`

Recommended files:
- `Homepage.tsx`
- `WorkloadsPage.tsx`

Responsibility:
- compose page zones
- consume shared context
- trigger workflow-relevant events

Must not contain:
- the source of truth for shared state
- cross-domain state ownership

### 4.3 Action Surface Files

Recommended group:
- `frontend/src/features/actions/`

Recommended files:
- `CreateApplySurface.tsx`
- `CreateApplyValidation.ts`
- `CreateApplyResult.tsx`

Responsibility:
- handle create/apply task UI
- consume action continuity state
- trigger action lifecycle events

Must not contain:
- cluster ownership
- namespace browsing ownership

### 4.4 Context-Aware UI Components

Recommended group:
- `frontend/src/components/context/`

Recommended files:
- `ClusterSelector.tsx`
- `NamespaceScopeSelector.tsx`
- `ContextSummary.tsx`

Responsibility:
- render reusable context controls and summaries

Must not contain:
- global workflow business rules

### 4.5 Workflow-Oriented Page Components

Recommended group:
- `frontend/src/components/workflows/`

Recommended files:
- `HomepagePrimaryTaskCard.tsx`
- `WorkloadsList.tsx`
- `WorkloadsToolbar.tsx`
- `ActionResultBanner.tsx`

Responsibility:
- render workflow-specific UI blocks
- remain driven by props or selectors

Must not contain:
- ownership of shared context source

### 4.6 SDK / API Access Files

Recommended group:
- `frontend/src/sdk/`

Recommended files:
- keep API parsing and transport concerns here
- add workflow-specific clients only if needed, such as:
  - `clustersApi.ts`
  - `workloadsApi.ts`
  - `applyApi.ts`

Responsibility:
- data fetching, parsing, transport contracts

Must not contain:
- shared context state ownership
- page navigation rules

## 5. File Responsibility Rules By Domain

### 5.1 Cluster Responsibility

Should live in:
- work-context state files
- context-aware selector and control components

Should not live in:
- page-local component state as the source of truth

### 5.2 Namespace Responsibility

Should live in:
- work-context state files

May be rendered in:
- Homepage summary
- Workloads scope controls
- Create/Apply target confirmation

Should not live in:
- isolated page-only state with no continuity contract

### 5.3 Workflow Domain Responsibility

Should live in:
- workflow context state files

May be consumed in:
- Homepage entry logic
- Workloads page
- return flow handling

Should not live in:
- arbitrary layout containers

### 5.4 List Continuity Responsibility

Should live in:
- list context state files

May be consumed in:
- Workloads page and its direct child components

Should not live in:
- global shell layout
- unrelated pages

### 5.5 Action Continuity Responsibility

Should live in:
- action context state files

May be consumed in:
- Create/Apply surface
- result feedback surface
- Workloads return state

Should not live in:
- unrelated page UI

## 6. Recommended V1 File Creation Order

For V1, files should be introduced in this order:

1. `frontend/src/state/work-context/types.ts`
2. `frontend/src/state/work-context/clusterContext.ts`
3. `frontend/src/state/work-context/namespaceContext.ts`
4. `frontend/src/state/work-context/workflowContext.ts`
5. `frontend/src/state/work-context/actionContext.ts`
6. `frontend/src/state/work-context/listContext.ts`
7. `frontend/src/state/work-context/events.ts`
8. `frontend/src/state/work-context/selectors.ts`
9. `frontend/src/pages/Homepage.tsx`
10. `frontend/src/pages/WorkloadsPage.tsx`

The Create/Apply surface can follow only after the shared context files are accepted.

## 7. What This Draft Explicitly Avoids

This draft does not assume:
- the current `App.tsx` should remain the main orchestration surface
- the current state files are the final shape
- the existing file split is the right ownership model

If any current file conflicts with this responsibility model, the model should win.

## 8. Ready For Code Planning

This file responsibility draft is ready for implementation planning when:
- the state-module split is accepted,
- the page responsibility split is accepted,
- and the team agrees to let workflow boundaries drive file ownership.
