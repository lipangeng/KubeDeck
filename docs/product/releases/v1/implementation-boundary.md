# KubeDeck V1 Implementation Boundary

## 1. Purpose

This document defines what the first implementation of KubeDeck must include and what it must explicitly defer.

The goal is to prevent scope expansion before the first core workflow becomes real and usable.

## 2. Boundary Principle

V1 is complete only when one real user workflow is usable end-to-end.

V1 is not complete because:
- the shell renders,
- metadata endpoints exist,
- or extension points are declared in architecture documents.

## 3. V1 Must Deliver

### 3.1 Core Workflow

V1 must deliver one complete workflow:

1. Enter Homepage
2. Confirm or change cluster
3. Enter `Workloads`
4. View a real workload list
5. Start `Create` or `Apply`
6. Submit with valid target resolution
7. Receive clear result feedback
8. Return to the workload context without losing state

### 3.2 Required Pages

V1 must include:
- Homepage
- Workloads page
- Create / Apply interaction surface
- contextual result feedback in the same task flow

### 3.3 Required Context

V1 must preserve:
- active cluster
- namespace context
- current resource domain
- key list filters needed for task continuity

### 3.4 Required Namespace Support

V1 must support:
- `single` namespace scope for browsing and action defaults
- `all` namespace scope for browsing

V1 must also ensure:
- create/apply resolves to a concrete namespace target when required
- cluster-scoped resources are handled explicitly
- namespace context persists across Homepage, Workloads, and Create / Apply

### 3.5 Required Feedback

V1 must provide:
- success feedback with affected objects
- failure feedback with preserved input context
- partial-failure distinction when multi-document or mixed outcomes exist

### 3.6 Required Homepage Behavior

Homepage must:
- prioritize task entry over diagnostics
- show active cluster clearly
- provide a direct entry into `Workloads`
- avoid making technical source categories part of the primary UX

### 3.7 Required Workloads Behavior

Workloads page must:
- show current cluster and namespace scope clearly
- present a real resource list
- expose `Create` or `Apply` prominently
- preserve context after actions

## 4. V1 Should Deliver If Low Cost

These items are useful but not required for V1 completion:
- last-used namespace restoration per cluster
- recent task or resume shortcut on Homepage
- lightweight empty-state guidance
- secondary refresh status messaging
- delayed support for `multiple` namespace selection in the UI

These should be added only if they do not delay the first usable workflow.

## 5. V1 Must Not Include

V1 must not expand into:
- AI-assisted workflows
- generalized chat or assistant surfaces
- broad plugin-driven workflow delivery
- large dashboard-style observability surfaces
- multi-resource-domain expansion beyond the first real workflow
- detailed visual polish work before workflow completion
- advanced filtering and sorting that do not materially support the first task

## 6. Deferred After V1

The following are valid post-V1 directions:
- AI as a later task-enhancement layer
- broader plugin-hosted workflows
- richer workload detail views
- `multiple` namespace selection in full UI form
- broader resource domains beyond `Workloads`
- homepage resume intelligence and task memory

## 7. Completion Check

V1 is in bounds when all answers below are yes:
- Can the user reach `Workloads` directly from Homepage?
- Can the user understand current cluster and namespace context immediately?
- Can the user perform one real create/apply operation?
- Can the operation resolve to a valid concrete target?
- Can the user understand success and failure without leaving the task path?
- Can the user continue working without losing context?

If any answer is no, V1 scope should not expand.
