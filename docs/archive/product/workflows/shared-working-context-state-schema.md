# KubeDeck Shared Working Context State Schema Draft

## 1. Purpose

This document turns the shared working context model into a structured state schema draft.

It does not define code types yet. It defines the intended state structure, required fields, optional fields, and V1 expectations.

## 2. Schema Principles

The schema must satisfy these rules:
- represent user operating context, not incidental UI structure
- distinguish browsing scope from execution target
- preserve continuity across Homepage, Workloads, and Create / Apply
- remain small enough to avoid becoming a dump for all UI state

## 3. Root State

The shared working context root should contain:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- `actionContext`

## 4. Field Schema

### 4.1 `activeCluster`

Meaning:
- the currently active operating cluster

Recommended shape:

```text
activeCluster:
  id: string
  status: ready | switching | failed
  lastStableId?: string
```

Required in V1:
- `id`
- `status`

Optional in V1:
- `lastStableId`

### 4.2 `namespaceScope`

Meaning:
- the current browsing scope under the active cluster

Recommended shape:

```text
namespaceScope:
  mode: single | multiple | all
  values: string[]
  source: default | restored | user_selected
```

Rules:
- `single` should have exactly one value
- `multiple` should have more than one value
- `all` should have an empty or ignored `values` list

Required in V1:
- `mode`
- `source`

Required V1 behavior:
- support `single`
- support `all`

Optional in V1:
- full `multiple` behavior in the UI

### 4.3 `currentWorkflowDomain`

Meaning:
- the current top-level workflow area

Recommended shape:

```text
currentWorkflowDomain:
  id: workloads | other_future_domain
  source: homepage_entry | direct_navigation | return_flow
```

Required in V1:
- `id`

Optional in V1:
- `source`

V1 expectation:
- `workloads` is the only required domain

### 4.4 `listContext`

Meaning:
- the browsing context required for continuity when returning to the current workflow domain

Recommended shape:

```text
listContext:
  searchText?: string
  statusFilters?: string[]
  subtypeFilters?: string[]
  sortKey?: string
  sortDirection?: asc | desc
```

Required in V1:
- none at schema level

Required V1 capability:
- the structure must exist even if the initial field set is minimal

V1 guidance:
- keep only continuity-relevant fields
- do not store temporary presentation state

### 4.5 `actionContext`

Meaning:
- the context needed while an operation is in progress or immediately after it completes

Recommended shape:

```text
actionContext:
  actionType?: create | apply
  originDomain?: workloads
  status?: idle | editing | validating | submitting | success | partial_failure | failure
  executionTarget?:
    kind: namespace | cluster_scoped
    namespace?: string
  resultSummary?:
    outcome: success | partial_failure | failure
    affectedObjects?: string[]
    failedObjects?: string[]
```

Required in V1:
- `status` once an action begins
- `executionTarget` before submit

Required V1 capability:
- explicit execution target resolution
- result summary after completion

Optional in V1:
- rich object metadata inside result summary

## 5. V1 Required Fields

For V1, the minimum required field set is:

```text
activeCluster.id
activeCluster.status
namespaceScope.mode
namespaceScope.source
currentWorkflowDomain.id
actionContext.status
actionContext.executionTarget
```

The rest may start minimal and expand only when needed for workflow continuity.

## 6. Field Ownership

Fields owned by shared working context:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- continuity-relevant `listContext`
- continuity-relevant `actionContext`

Fields not owned by shared working context:
- modal visibility flags
- drawer visibility flags
- transient hover/focus state
- non-blocking diagnostics
- purely local draft formatting state

## 7. Transition Expectations

### 7.1 Cluster Change

On cluster change:
- `activeCluster.status` becomes `switching`
- `namespaceScope` is restored or safely reset
- incompatible `actionContext` clears
- `currentWorkflowDomain` may remain if still valid

### 7.2 Workflow Entry

On Homepage -> Workloads:
- `currentWorkflowDomain.id` becomes `workloads`
- shared cluster and namespace stay intact

### 7.3 Action Start

On Create / Apply open:
- `actionContext.actionType` is set
- `actionContext.status` becomes `editing`

### 7.4 Validation

Before submit:
- `actionContext.status` becomes `validating`
- `executionTarget` must become concrete before transition to `submitting`

### 7.5 Submission

On submit:
- `actionContext.status` becomes `submitting`

On completion:
- `actionContext.status` becomes `success`, `partial_failure`, or `failure`
- `resultSummary` becomes available

### 7.6 Return Flow

On return to Workloads:
- `currentWorkflowDomain.id` remains `workloads`
- `namespaceScope` stays preserved
- `listContext` stays preserved
- `actionContext` may be cleared after acknowledgement

## 8. Validation Constraints

The schema must support these product constraints:
- browsing scope may be `all`
- execution target may not be ambiguous
- `all` is never a write target
- cluster switch must not leave undefined mixed state
- action failure must not destroy browsing continuity

## 9. Deferred Schema Capacity

The schema should leave room for later additions without changing the root model:
- richer `multiple` namespace behavior
- workload detail context
- more workflow domains
- resume state and remembered task history
- later AI-related context, if product scope changes after V1

## 10. Ready For Code Mapping

This schema is ready for implementation mapping when:
- V1 field minimums are accepted
- namespace scope semantics are accepted
- execution target semantics are accepted
- and no one is trying to use shared context as a container for all local UI state
