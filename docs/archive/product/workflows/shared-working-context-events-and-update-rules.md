# KubeDeck Shared Working Context Events And Update Rules

## 1. Purpose

This document defines which events are allowed to update the shared working context, which fields each event may change, which fields it must not change, and which state constraints must hold after the event.

It exists to make the shared context operationally safe before code mapping begins.

## 2. Event Design Rules

All events must follow these rules:
- an event may update only the minimum fields needed for its purpose
- an event must not silently reset unrelated continuity state
- cluster, namespace, browsing scope, and execution target must remain logically separated
- action events must not rewrite browsing context unless the user explicitly chose that change

## 3. Event List

The first workflow should allow these events:

1. `enter_homepage`
2. `request_cluster_switch`
3. `complete_cluster_switch`
4. `fail_cluster_switch`
5. `enter_workloads`
6. `update_namespace_scope`
7. `update_list_context`
8. `start_action`
9. `validate_action`
10. `fail_action_validation`
11. `resolve_execution_target`
12. `submit_action`
13. `complete_action_success`
14. `complete_action_partial_failure`
15. `complete_action_failure`
16. `acknowledge_action_result`
17. `return_to_workloads`

## 4. Event Rules

### 4.1 `enter_homepage`

May modify:
- `currentWorkflowDomain`

Must not modify:
- `activeCluster`
- `namespaceScope`
- `listContext`
- `actionContext`

Post-event constraints:
- shared context remains valid
- no task continuity state is lost just because Homepage became active

### 4.2 `request_cluster_switch`

May modify:
- `activeCluster.status`

Must not modify:
- `activeCluster.id`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- `actionContext`

Post-event constraints:
- `activeCluster.status` becomes `switching`
- previous stable context remains available until switch completes or fails

### 4.3 `complete_cluster_switch`

May modify:
- `activeCluster.id`
- `activeCluster.status`
- `activeCluster.lastStableId`
- `namespaceScope`
- `actionContext`

Must not modify:
- unrelated list context for the same valid workflow domain unless it becomes invalid under the new cluster

Post-event constraints:
- `activeCluster.status` becomes `ready`
- `namespaceScope` is restored or safely reset
- incompatible action context is cleared
- no mixed old/new cluster context remains

### 4.4 `fail_cluster_switch`

May modify:
- `activeCluster.status`

Must not modify:
- `activeCluster.id`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- `actionContext`

Post-event constraints:
- `activeCluster.status` becomes `failed` or reverts to `ready` with the last stable cluster
- last stable working context remains usable
- the user is not left in undefined mixed state

### 4.5 `enter_workloads`

May modify:
- `currentWorkflowDomain`

Must not modify:
- `activeCluster`
- `namespaceScope`
- `actionContext`

May preserve:
- `listContext`

Post-event constraints:
- `currentWorkflowDomain.id` is `workloads`
- cluster and namespace continuity remain intact

### 4.6 `update_namespace_scope`

May modify:
- `namespaceScope`

Must not modify:
- `activeCluster`
- `currentWorkflowDomain`
- `actionContext.executionTarget`

May require revalidation of:
- `actionContext`

Post-event constraints:
- namespace browsing scope becomes the newly selected scope
- browsing scope and execution target remain distinct
- if action context exists, it is marked for revalidation rather than silently rewritten

### 4.7 `update_list_context`

May modify:
- `listContext`

Must not modify:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `actionContext`

Post-event constraints:
- list context reflects the current workflow domain only
- temporary visual-only state does not enter shared context

### 4.8 `start_action`

May modify:
- `actionContext.actionType`
- `actionContext.originDomain`
- `actionContext.status`

Must not modify:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`

Post-event constraints:
- `actionContext.status` becomes `editing`
- browsing context remains unchanged

### 4.9 `validate_action`

May modify:
- `actionContext.status`

Must not modify:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- `actionContext.executionTarget`

Post-event constraints:
- `actionContext.status` becomes `validating`
- execution target remains unchanged until explicitly resolved

### 4.10 `fail_action_validation`

May modify:
- `actionContext.status`
- validation-related metadata inside `actionContext` if later added

Must not modify:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- already valid browsing scope

Post-event constraints:
- `actionContext.status` returns to `editing`
- user input remains recoverable
- no browsing continuity is lost

### 4.11 `resolve_execution_target`

May modify:
- `actionContext.executionTarget`

Must not modify:
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- `activeCluster`

Post-event constraints:
- execution target becomes concrete
- if namespaced, target namespace is explicit
- `all` and unresolved `multiple` are never stored as execution target values

### 4.12 `submit_action`

May modify:
- `actionContext.status`

Must not modify:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- `actionContext.executionTarget`

Post-event constraints:
- `actionContext.status` becomes `submitting`
- execution target is already concrete

### 4.13 `complete_action_success`

May modify:
- `actionContext.status`
- `actionContext.resultSummary`

Must not modify:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`

Post-event constraints:
- `actionContext.status` becomes `success`
- result summary is available
- browsing continuity remains intact

### 4.14 `complete_action_partial_failure`

May modify:
- `actionContext.status`
- `actionContext.resultSummary`

Must not modify:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`

Post-event constraints:
- `actionContext.status` becomes `partial_failure`
- successful and failed outcomes are distinguishable
- browsing continuity remains intact

### 4.15 `complete_action_failure`

May modify:
- `actionContext.status`
- `actionContext.resultSummary`

Must not modify:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`

Post-event constraints:
- `actionContext.status` becomes `failure`
- result summary is available
- failure does not destroy browsing continuity

### 4.16 `acknowledge_action_result`

May modify:
- `actionContext`

Must not modify:
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`

Post-event constraints:
- acknowledgement may clear completed action result state
- browsing continuity remains intact

### 4.17 `return_to_workloads`

May modify:
- `currentWorkflowDomain`
- `actionContext`

Must not modify:
- `activeCluster`
- `namespaceScope`
- `listContext`

Post-event constraints:
- user is back in `workloads`
- cluster and namespace remain unchanged
- list context remains usable
- action context is either preserved for feedback display or safely cleared after acknowledgement

## 5. Forbidden Update Patterns

The following are forbidden:
- cluster switch silently changing list filters without validity reason
- namespace scope update silently changing execution target
- action submit changing browsing scope
- action failure clearing cluster or namespace context
- Homepage entry resetting workflow continuity
- storing shell diagnostics as shared working context updates

## 6. Global State Constraints

These constraints must always hold after any event:
- `activeCluster` is always defined
- `namespaceScope` is always subordinate to the active cluster
- `currentWorkflowDomain` never contradicts the active task path
- `executionTarget` is never ambiguous when submit is allowed
- browsing scope and execution target are never represented by the same field

## 7. V1 Minimum Event Support

For V1, the minimum required implemented event set is:
- `request_cluster_switch`
- `complete_cluster_switch`
- `fail_cluster_switch`
- `enter_workloads`
- `update_namespace_scope`
- `update_list_context`
- `start_action`
- `resolve_execution_target`
- `submit_action`
- `complete_action_success`
- `complete_action_failure`
- `return_to_workloads`

The remaining events may exist conceptually but can be simplified if V1 continuity is preserved.

## 8. Ready For Code Mapping

This event model is ready for implementation mapping when:
- the event list is accepted,
- forbidden update patterns are accepted,
- and V1 minimum event support is accepted.
