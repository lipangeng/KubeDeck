# KubeDeck Task A First Patch Scope

## 1. Purpose

This document defines the minimum first patch scope for Task A: Shared Working Context Model.

The goal is to keep the first code change intentionally small, structural, and low-risk while establishing the new ownership model.

## 2. Patch Scope Principle

The first patch should establish the new shared working context foundation without trying to finish the workflow UI.

This patch should:
- create the new state structure
- avoid broad page rewrites
- avoid migrating old code prematurely
- and prepare the ground for later integration patches

## 3. What The First Patch Should Include

The first patch should include only:

1. new shared working context file set
2. root shared state types
3. minimal event/update interfaces
4. minimal selectors
5. baseline unit tests for state rules

## 4. Files To Create In The First Patch

Create these files:

- `frontend/src/state/work-context/types.ts`
- `frontend/src/state/work-context/clusterContext.ts`
- `frontend/src/state/work-context/namespaceContext.ts`
- `frontend/src/state/work-context/workflowContext.ts`
- `frontend/src/state/work-context/actionContext.ts`
- `frontend/src/state/work-context/listContext.ts`
- `frontend/src/state/work-context/events.ts`
- `frontend/src/state/work-context/selectors.ts`

Create matching tests for the new state modules where meaningful:

- `frontend/src/state/work-context/clusterContext.test.ts`
- `frontend/src/state/work-context/namespaceContext.test.ts`
- `frontend/src/state/work-context/workflowContext.test.ts`
- `frontend/src/state/work-context/actionContext.test.ts`
- `frontend/src/state/work-context/listContext.test.ts`
- `frontend/src/state/work-context/events.test.ts`

Not every file needs a separate test if coverage is better consolidated, but the state rules must be tested.

## 5. What The First Patch Must Not Include

The first patch must not include:
- Homepage UI rewrite
- Workloads page rewrite
- Create / Apply surface implementation
- routing overhaul
- plugin integration changes
- backend contract changes
- visual redesign work
- AI-related code

## 6. Existing Files To Leave Untouched In The First Patch

Do not modify these categories in the first patch unless absolutely required for type import wiring:
- `frontend/src/App.tsx`
- page-shell components
- theme files
- plugin host files
- existing SDK transport/parsing files
- backend files

Reason:
- the first patch should establish new ownership, not entangle itself with migration work

## 7. Existing Files That May Receive Minimal Touches

Only minimal touches are acceptable in the first patch for:
- barrel or import wiring if needed
- test configuration if required by the new state module tests

These touches must stay narrow and must not start UI migration implicitly.

## 8. Behavioral Scope Of The First Patch

The first patch should prove only these behaviors:
- cluster state can be represented cleanly
- namespace browsing scope can be represented cleanly
- workflow domain can be represented cleanly
- action execution target is structurally separate from namespace browsing scope
- event/update entry points preserve the model constraints

The first patch does not need to prove:
- end-to-end UI continuity
- real screen navigation
- visible task entry behavior

## 9. Testing Scope Of The First Patch

The first patch should test:
- cluster switch state transitions
- namespace scope validity rules
- action execution target resolution shape
- event ownership boundaries
- preservation of unrelated state during event handling

The first patch should not yet test:
- page rendering outcomes
- visual layout behavior
- browser navigation flows

## 10. Destructive Change Rule For The First Patch

The first patch should not clear old code yet.

Even though later replacement is allowed, the first patch should focus on creating the new structure first.

Allowed in first patch:
- coexistence with old code

Not allowed in first patch:
- large deletion of historical page/state code before the replacement structure is in place

## 11. Success Criteria

The first patch is successful when:
- the new state/work-context structure exists
- the core V1 context concepts have one canonical structural home
- tests verify the core state rules
- and no unnecessary UI migration or legacy cleanup was mixed into the patch

## 12. Next Patch After This One

The next patch after the first one should focus on:
- wiring Homepage and Workloads to consume the new context foundation

That second patch, not this one, should begin controlled migration away from old ownership patterns.
