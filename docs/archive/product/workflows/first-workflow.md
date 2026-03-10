# KubeDeck First Core Workflow Draft

## 1. Purpose

This document defines the first core workflow that must become real before broader UI expansion, plugin growth, or AI enhancement.

## 2. Target User

The first workflow should primarily serve:
- platform engineers,
- SRE / DevOps operators,
- and delivery engineers who perform real operations across multiple clusters.

This workflow is not designed first for passive viewers or low-frequency users.

## 3. Workflow Goal

The goal of the first workflow is:
- enter the system,
- confirm the correct cluster context,
- enter a real resource domain,
- perform one standard resource operation,
- receive a clear result,
- and continue working without losing context.

In short, the workflow must prove that KubeDeck is usable as a control plane, not only presentable as a shell.

## 4. Workflow Definition

The first core workflow is:

1. Enter homepage.
2. Confirm or change active cluster.
3. Enter `Workloads`.
4. View a real workload list.
5. Start a standard action such as `Create` or `Apply`.
6. Submit the action.
7. Receive success, partial-failure, or failure feedback.
8. Return to the workload context and continue work.

## 5. Page Inventory

The minimum page set should be:
- Homepage
- Workloads page
- Create / Apply surface
- Result feedback state inside the same task flow

The create/apply surface may be a page, drawer, or dialog. For MVP, the key requirement is task continuity rather than UI form.

## 6. Single Responsibility Per Page

### Homepage

Unique responsibility:
- establish working context,
- and route the user into the primary task flow quickly.

Homepage should not be the main place for diagnostics, framework metadata, or implementation-status display.

### Workloads Page

Unique responsibility:
- act as the main workspace for browsing and selecting resources in the current context.

This page should make cluster, namespace, and resource scope obvious.

### Create / Apply Surface

Unique responsibility:
- collect input and complete one standard operation.

It should not force the user to rebuild context or navigate elsewhere for unrelated information.

### Result Feedback State

Unique responsibility:
- explain what happened after submit,
- show what succeeded or failed,
- and direct the user to the next immediate action.

## 7. Page Flow

The intended page flow is:

1. Homepage
2. Workloads page
3. Create / Apply surface
4. Submission state
5. Result feedback
6. Return to Workloads page with preserved context

If the user fails during submission, the flow should remain inside the same task path rather than sending the user back to homepage.

## 8. Critical State

The workflow must preserve these states:
- `activeCluster`
- `namespace`
- current resource domain
- current list filters
- current operation mode
- submission status
- operation result summary

The two most important anchors are `activeCluster` and `namespace`.

## 9. Success Feedback Rules

Success feedback must:
- clearly indicate the action succeeded,
- identify which object or objects were affected,
- preserve the user context,
- and let the user continue working immediately.

Success feedback should not disappear without leaving evidence in the current working view.

## 10. Failure Feedback Rules

Failure feedback must:
- clearly indicate the action failed,
- identify which step or object failed,
- preserve the user input context,
- and tell the user whether to retry, edit, or return.

If partial failure exists, successful and failed items must be distinguishable.

Failure feedback should not be only a raw backend error string.

## 11. Homepage Information Architecture Draft

Homepage should keep:
- active cluster selector,
- primary task entry,
- current working context summary,
- and only the most relevant blocking status hints.

Homepage should demote or remove from primary focus:
- runtime health as headline content,
- raw API target information,
- registry type counts,
- failure summaries for shell diagnostics,
- and navigation grouped by implementation source.

Homepage should answer these three questions first:
- Which cluster am I working in?
- What is the next likely task?
- Can I enter that task immediately?

## 12. Scope Boundary

This first workflow does not require:
- AI in the task loop,
- a full plugin ecosystem,
- broad resource coverage,
- or a complete control-plane dashboard.

It only requires one real, repeatable, task-complete operating path.

## 13. Page Module Inventory

### Homepage Modules

Required modules:
- global header
- active cluster selector
- primary task entry card
- current context summary
- recent or default task shortcut
- blocking status notice area

Optional but low priority modules:
- lightweight announcements
- recent activity summary

Homepage should not treat diagnostics widgets as primary modules.

### Workloads Page Modules

Required modules:
- page header with current cluster and namespace context
- resource domain navigation or title
- namespace selector or namespace scope display
- filter and search area
- workload list table or list
- primary action entry such as `Create` or `Apply`
- refresh or reload control

Optional but low priority modules:
- empty state guidance
- secondary summary metrics

### Create / Apply Surface Modules

Required modules:
- operation title and current context summary
- namespace confirmation field
- input area for YAML or structured form
- validation and warning region
- submit action
- cancel / back action

Optional but low priority modules:
- examples or templates
- advanced options section

### Result Feedback Modules

Required modules:
- result status headline
- affected object summary
- success / failure / partial-failure breakdown
- actionable next step area
- return to workload context action

Optional but low priority modules:
- raw technical detail expansion
- copyable diagnostic payload

## 14. State Transition Rules

### Primary Workflow States

The first workflow should use these high-level states:
- `homepage_idle`
- `cluster_switching`
- `workloads_loading`
- `workloads_ready`
- `operation_editing`
- `operation_validating`
- `operation_submitting`
- `operation_success`
- `operation_partial_failure`
- `operation_failure`

### Transition Rules

1. Homepage entry
- Default state: `homepage_idle`
- If user changes cluster, move to `cluster_switching`

2. Cluster change
- From `cluster_switching`, load the target workload context
- On success, transition to `workloads_loading` and then `workloads_ready`
- On failure, remain in the current known-safe context and show blocking feedback

3. Enter workload workspace
- From homepage primary task entry, transition to `workloads_loading`
- Once list data and context are usable, transition to `workloads_ready`

4. Start create/apply
- From `workloads_ready`, opening the task surface transitions to `operation_editing`

5. Validate input
- From `operation_editing`, user-triggered checks or submit prechecks transition to `operation_validating`
- If validation succeeds, return to `operation_editing` with ready-to-submit status
- If validation fails, remain in `operation_editing` and preserve user input

6. Submit operation
- From `operation_editing`, submit transitions to `operation_submitting`
- Input must remain recoverable while submit is in progress

7. Submit result
- On full success, transition to `operation_success`
- On mixed result, transition to `operation_partial_failure`
- On failure, transition to `operation_failure`

8. Return to workspace
- From any result state, returning transitions back to `workloads_loading`
- Once refreshed data is available, transition to `workloads_ready`
- Cluster, namespace, and relevant filters must be preserved

### Preservation Rules

These values must persist across the workflow unless the user intentionally changes them:
- active cluster
- namespace
- current resource domain
- list filter state

These values may reset after a completed operation:
- draft input, only if the user explicitly exits or starts a new operation
- temporary validation messages

### Error Handling Rules

- A failed submit must not reset cluster or namespace context
- A failed validation must not discard user input
- A failed workload reload after submit must still preserve result feedback until the user acknowledges it
- A cluster switch failure must not leave the user in an undefined mixed context
