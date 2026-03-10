# KubeDeck Shared Working Context Model Spec

## 1. Purpose

This document defines the shared working context model for KubeDeck.

It is a product-level specification, not an implementation description.

Its purpose is to define which pieces of state must persist across the first workflow, which pieces are page-local, and which transitions are allowed.

## 2. Design Principle

The shared working context should represent the user's active operating context, not the current UI implementation.

It should answer:
- where the user is operating,
- what scope they are operating within,
- what workflow area they are in,
- and what context must survive navigation and actions.

It should not absorb every temporary UI detail.

## 3. Context Hierarchy

The shared working context should follow this hierarchy:

1. active cluster
2. namespace scope
3. current workflow domain
4. workflow continuity state

This hierarchy means:
- cluster is the top-level anchor,
- namespace is subordinate to cluster,
- page or resource domain is subordinate to cluster and namespace,
- and continuity state exists only to preserve workflow progress.

## 4. Shared Context Fields

The shared working context should contain these fields.

### 4.1 Active Cluster

Field:
- `activeCluster`

Meaning:
- the cluster the user is currently operating against

Why shared:
- every page and action depends on it

### 4.2 Namespace Scope

Field:
- `namespaceScope`

Meaning:
- the current namespace browsing scope under the active cluster

Expected shape:
- `single`
- `multiple`
- `all`

V1 requirement:
- support `single` and `all` in actual product behavior
- retain model compatibility for `multiple`

Why shared:
- it must persist across Homepage, Workloads, details, and action surfaces

### 4.3 Current Workflow Domain

Field:
- `currentWorkflowDomain`

Meaning:
- the current primary work area, such as `workloads`

Why shared:
- it defines where the user is in the first workflow
- it helps preserve task continuity on return

### 4.4 List Context

Field:
- `listContext`

Meaning:
- the minimal browsing state needed to continue the current workflow after navigation or action

May include:
- search text
- selected status filters
- selected subtype filters
- sort mode

Why shared:
- users should be able to return to their working list context after create/apply

Constraint:
- only continuity-relevant list state belongs here
- purely cosmetic or temporary UI state does not

### 4.5 Action Context

Field:
- `actionContext`

Meaning:
- the minimal persistent context needed for in-progress or just-finished task actions

May include:
- action type such as `create` or `apply`
- originating workflow domain
- resolved execution target summary
- last action result summary

Why shared:
- success/failure feedback and return flow depend on it

Constraint:
- draft content itself does not have to live in shared context unless explicitly required for recovery

## 5. What Is Not Shared Context

The following should not be part of shared working context by default:
- modal open/close flags
- drawer width or visual layout state
- hover state
- row expansion state
- purely local form formatting state
- diagnostics probe timestamps
- shell-only status widgets unrelated to workflow continuity

These belong to page-local or component-local state unless they become necessary for cross-page recovery.

## 6. Page Responsibilities Against Shared Context

### 6.1 Homepage

Homepage may read:
- `activeCluster`
- `namespaceScope` summary
- `currentWorkflowDomain` only if needed for resume semantics

Homepage may update:
- `activeCluster`
- entry into the first workflow domain

Homepage should not own:
- detailed list state
- operation draft state

### 6.2 Workloads Page

Workloads may read:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- relevant action result summary

Workloads may update:
- `currentWorkflowDomain`
- `listContext`
- namespace scope
- start of action context

### 6.3 Create / Apply Surface

Create / Apply may read:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `actionContext`

Create / Apply may update:
- resolved action target
- action status
- action result summary

Create / Apply should not overwrite:
- the user's browsing context unless the user explicitly changes it

### 6.4 Result Feedback

Result feedback may read:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `actionContext`

Result feedback may update:
- only acknowledgement or completion markers needed for return flow

## 7. Persistence Rules

These fields must persist across the first workflow:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- continuity-relevant `listContext`

These fields may persist only for the duration of one action cycle:
- `actionContext`
- action result summary

These fields may reset freely:
- purely local UI presentation state

## 8. Transition Rules

### 8.1 Cluster Change

When `activeCluster` changes:
- namespace scope must be revalidated or safely reset
- current workflow domain may remain the same if the target cluster supports the same workflow path
- incompatible action context must be cleared

### 8.2 Namespace Scope Change

When `namespaceScope` changes:
- browsing context remains valid if the page supports the new scope
- action context may need revalidation before submit

### 8.3 Workflow Domain Change

When `currentWorkflowDomain` changes:
- cluster and namespace remain
- page-specific list context may reset if the new domain is different

### 8.4 Action Start

When create/apply starts:
- browsing context remains intact
- action context is created
- namespace resolution must move toward a concrete execution target

### 8.5 Action Completion

When create/apply completes:
- action result summary becomes available
- user returns to the same workflow domain
- browsing context is preserved unless the user explicitly changes it

## 9. Namespace Requirements In The Model

The model must distinguish clearly between:
- browsing scope
- execution target

This is critical because:
- browsing may use `all`
- execution may not use ambiguous `all`

Therefore:
- `namespaceScope` represents browsing context
- execution target belongs in `actionContext`

These two must not be collapsed into one field.

## 10. V1 Model Requirements

For V1, the shared model must be sufficient to support:
- Homepage -> Workloads entry
- Workloads browsing with context continuity
- Create/apply with explicit target resolution
- return to Workloads without losing cluster and namespace context

If the model cannot support those four outcomes, it is not acceptable for V1.

## 11. Implementation Mapping Rule

Only after this model is accepted should the implementation map be created.

Implementation must adapt to this model.
The model must not adapt to accidental current code structure.
