# KubeDeck Shared Working Context Implementation Mapping Draft

## 1. Purpose

This document maps the shared working context model, state schema, and event rules into an implementation-oriented structure.

It is still pre-code. Its role is to define what kinds of state modules, update entry points, and page consumers should exist before implementation starts.

## 2. Mapping Principle

Implementation must follow the product model, not the current accidental UI structure.

This mapping must:
- keep cluster, namespace, workflow, list continuity, and action continuity separate
- avoid one oversized global state bucket
- avoid duplicating the same context in unrelated pages
- support the V1 first workflow before broader expansion

## 3. Recommended State Module Split

The shared working context should be implemented as a small set of cooperating state modules, not one undifferentiated store.

Recommended logical modules:
- cluster context module
- namespace context module
- workflow context module
- list continuity module
- action continuity module

## 4. Module Responsibilities

### 4.1 Cluster Context Module

Owns:
- `activeCluster.id`
- `activeCluster.status`
- `activeCluster.lastStableId`

Handles events:
- `request_cluster_switch`
- `complete_cluster_switch`
- `fail_cluster_switch`

Must not own:
- namespace browsing rules
- list filters
- create/apply result state

### 4.2 Namespace Context Module

Owns:
- `namespaceScope.mode`
- `namespaceScope.values`
- `namespaceScope.source`

Handles events:
- `update_namespace_scope`
- cluster-switch-related namespace restore or reset

Must not own:
- execution target for create/apply
- page-local namespace UI presentation details

### 4.3 Workflow Context Module

Owns:
- `currentWorkflowDomain.id`
- `currentWorkflowDomain.source`

Handles events:
- `enter_homepage`
- `enter_workloads`
- `return_to_workloads`

Must not own:
- list filters themselves
- action result details

### 4.4 List Continuity Module

Owns:
- `listContext.searchText`
- `listContext.statusFilters`
- `listContext.subtypeFilters`
- `listContext.sortKey`
- `listContext.sortDirection`

Handles events:
- `update_list_context`
- workflow-domain-compatible restore on return

Must not own:
- purely visual table state
- row hover or row expansion details

### 4.5 Action Continuity Module

Owns:
- `actionContext.actionType`
- `actionContext.originDomain`
- `actionContext.status`
- `actionContext.executionTarget`
- `actionContext.resultSummary`

Handles events:
- `start_action`
- `validate_action`
- `fail_action_validation`
- `resolve_execution_target`
- `submit_action`
- `complete_action_success`
- `complete_action_partial_failure`
- `complete_action_failure`
- `acknowledge_action_result`

Must not own:
- namespace browsing scope
- unrelated list browsing state

## 5. Recommended Update Entry Points

The implementation should expose explicit update entry points grouped by intent, not arbitrary direct mutation.

Recommended update entry groups:
- cluster updates
- namespace scope updates
- workflow navigation updates
- list continuity updates
- action lifecycle updates

These groups should correspond to the event model, even if the final implementation uses reducers, stores, or controller functions.

## 6. Recommended Page Consumption Map

### 6.1 Homepage

Homepage should consume:
- cluster context
- namespace context summary
- workflow context for primary entry semantics

Homepage should trigger:
- cluster updates
- workflow entry updates

Homepage should not consume:
- detailed list continuity state unless resume behavior is intentionally added
- active action draft internals

### 6.2 Workloads Page

Workloads should consume:
- cluster context
- namespace context
- workflow context
- list continuity context
- action result summary when returning from an action

Workloads should trigger:
- namespace scope updates
- list continuity updates
- action start
- workflow re-entry if needed

### 6.3 Create / Apply Surface

Create / Apply should consume:
- cluster context
- namespace browsing context
- action continuity context

Create / Apply should trigger:
- action lifecycle updates
- execution target resolution

Create / Apply should not trigger:
- direct browsing-scope rewrites without explicit user action

### 6.4 Result Feedback Surface

Result feedback should consume:
- action continuity context
- cluster context
- namespace browsing context

Result feedback should trigger:
- acknowledgement
- return to Workloads

## 7. Cross-Module Coordination Rules

The modules must coordinate under these rules:

- cluster module may force namespace module to restore or reset
- namespace module may require action module revalidation
- workflow module may preserve or clear list continuity depending on domain compatibility
- action module must not rewrite namespace browsing context
- list continuity module must survive action completion and return flow

## 8. V1 Implementation Recommendation

For V1, these modules should be considered required:
- cluster context module
- namespace context module
- workflow context module
- action continuity module

The list continuity module may begin minimal, but its structural place must still exist.

## 9. What Should Not Be Mapped Yet

Do not map these into the first implementation structure yet:
- plugin extension state
- AI interaction state
- resume-memory systems
- full multi-domain navigation coordination
- advanced analytics or diagnostics state

## 10. Implementation-Readiness Check

This mapping is ready for code-level design when:
- the module split is accepted
- the event ownership is accepted
- the page consumption map is accepted
- and V1 scope is still limited to the first workflow
