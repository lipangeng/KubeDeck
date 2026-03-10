# KubeDeck Task A Code Change Plan

## 1. Purpose

This document defines the first code-planning step for V1 Task A: Shared Working Context Model.

It does not implement code. It defines the initial file creation order, the file responsibility boundary, what existing code should remain untouched at first, and when old code may be replaced entirely.

## 2. Planning Rule

This plan follows the new product model, not the current historical code shape.

If existing code conflicts with the new model:
- the new model wins
- historical code may be reduced, bypassed, or fully replaced

Compatibility with old structure is not a goal by itself.

## 3. Task A Goal

Task A should produce the first canonical shared working context foundation for:
- active cluster
- namespace scope
- current workflow domain
- list continuity
- action continuity

Task A is complete when these concepts have one coherent state source and are no longer spread across unrelated page-local ownership patterns.

## 4. First File Creation Set

The first code change set should create these files:

- `frontend/src/state/work-context/types.ts`
- `frontend/src/state/work-context/clusterContext.ts`
- `frontend/src/state/work-context/namespaceContext.ts`
- `frontend/src/state/work-context/workflowContext.ts`
- `frontend/src/state/work-context/actionContext.ts`
- `frontend/src/state/work-context/listContext.ts`
- `frontend/src/state/work-context/events.ts`
- `frontend/src/state/work-context/selectors.ts`

This set should exist before deeper page rewrites begin.

## 5. Responsibility Of Each New File

### `types.ts`

Should define:
- shared working context types
- root state shape
- key enum-like unions for V1

Should not define:
- page rendering logic
- fetch logic

### `clusterContext.ts`

Should define:
- cluster state slice
- cluster update logic
- switch lifecycle rules

Should not define:
- namespace defaults beyond explicit coordination contracts
- list state

### `namespaceContext.ts`

Should define:
- namespace scope state slice
- browsing-scope update logic
- restore/reset rules after cluster switch

Should not define:
- create/apply execution target ownership

### `workflowContext.ts`

Should define:
- current workflow domain state
- workflow entry and return rules

Should not define:
- list filters
- action result details

### `actionContext.ts`

Should define:
- action lifecycle state
- execution target state
- result summary state

Should not define:
- namespace browsing source of truth

### `listContext.ts`

Should define:
- continuity-relevant list state

Should not define:
- table presentation-only details

### `events.ts`

Should define:
- the public event/update entry points for the shared working context

Should not define:
- page layout logic

### `selectors.ts`

Should define:
- selectors for Homepage, Workloads, Create/Apply, and result feedback consumers

Should not define:
- mutation logic

## 6. Existing Code Handling Strategy

At the first step, existing code should be handled conservatively but not treated as authoritative.

### Keep Untouched Initially

These categories may stay untouched during the first file-creation step:
- theme and theme mode files
- SDK parsing files unless required by context shape
- plugin host files
- page-shell presentational helpers

### Do Not Use As Design Source

These categories must not be treated as the source of truth for the new model:
- current top-level page composition
- current local state layout in the main app entry
- historical state helper shape if it conflicts with the new context model

### Replace If Necessary

If old files block the new model, it is acceptable to:
- stop extending them
- bypass them
- or fully replace them later

Task A should not contort the new architecture to preserve them.

## 7. Recommended Execution Order

The first implementation sequence for Task A should be:

1. create `types.ts`
2. create `clusterContext.ts`
3. create `namespaceContext.ts`
4. create `workflowContext.ts`
5. create `actionContext.ts`
6. create `listContext.ts`
7. create `events.ts`
8. create `selectors.ts`

Only after these exist should page integration begin.

## 8. What Task A Should Not Do Yet

Task A should not yet:
- redesign Homepage visually
- build Workloads page UI
- implement Create / Apply UI
- expand plugin integration
- introduce AI-related context
- optimize visual polish

Task A is about shared state ownership, not about visible feature completeness.

## 9. Destructive-Change Rule

If a current file fundamentally prevents the shared working context model from landing cleanly, it is acceptable to clear or replace that code in a later implementation step.

But this should happen only after:
- the new replacement structure exists,
- the replacement responsibility is explicit,
- and the removed code is clearly outside the accepted model.

In short:
- do not preserve bad ownership for safety theatre
- do not delete old code before replacement structure is ready

## 10. Ready To Implement Task A

Task A is ready for real code work when:
- the new file set is accepted
- file responsibilities are accepted
- historical code is explicitly treated as optional, not binding
- and the team agrees that replacement is allowed when the old structure conflicts with the new model
