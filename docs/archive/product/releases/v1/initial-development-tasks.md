# KubeDeck V1 Initial Development Tasks

## 1. Purpose

This document converts the V1 implementation task breakdown into the first concrete batch of development tasks.

The goal is to define the smallest practical starting set that can move the product from documentation to implementation without reopening scope.

## 2. Selection Rule

The first batch should:
- unlock the shared workflow foundation,
- avoid UI polish work,
- avoid parallelizing before core context rules are stable,
- and create the minimum base required for Homepage -> Workloads -> Create / Apply continuity.

## 3. First Batch Scope

The first batch should include only Phase 1 and the minimum entry from Phase 2:

1. shared working context model
2. cluster switch lifecycle
3. namespace context baseline
4. Homepage primary task entry handoff

Everything else should wait until these four items are stable.

## 4. Initial Development Tasks

### Task A: Shared Working Context Model

Objective:
- define and implement one shared state model for the first workflow.

Must include:
- active cluster
- namespace scope
- current resource domain
- task continuity fields needed between pages

Acceptance:
- one canonical state model exists
- Homepage and Workloads can depend on the same context source
- context fields are not duplicated in unrelated local state without reason

Why first:
- this is the base contract for every later page and action

### Task B: Cluster Switch Lifecycle

Objective:
- make cluster switching deterministic and safe before expanding the UI flow.

Must include:
- loading state during cluster switch
- safe fallback on failure
- prevention of mixed old/new context after failure

Acceptance:
- cluster switch either completes cleanly or reverts safely
- post-switch context is unambiguous
- failure does not leave the user in a broken task path

Why second:
- namespace and task entry rules depend on stable cluster behavior

### Task C: Namespace Context Baseline

Objective:
- implement V1 namespace behavior for workflow continuity.

Must include:
- `single` browsing scope
- `all` browsing scope
- inheritance from Homepage to Workloads
- action-time resolution to explicit execution target

Acceptance:
- browsing and action semantics are not mixed
- create/apply cannot submit against ambiguous namespace target
- namespace context persists across the first workflow

Why third:
- this is the highest-risk context rule after cluster itself

### Task D: Homepage Primary Task Entry

Objective:
- make Homepage a real entry into the first workflow.

Must include:
- active cluster visibility
- direct entry into `Workloads`
- demotion of shell diagnostics from primary focus

Acceptance:
- Homepage clearly shows current cluster
- user can enter `Workloads` directly
- the first screen is task-led rather than diagnostics-led

Why fourth:
- once context rules are stable, Homepage can become the real start of the workflow

## 5. Second Batch Preview

Only after the first batch is stable should the next batch begin:
- Workloads page shell
- real workload list
- Create / Apply interaction surface

## 6. Explicit Non-Goals For First Batch

The first batch must not include:
- AI
- plugin workflow expansion
- detailed visual polish
- advanced list filtering
- resume/recent task UX
- multi-namespace UI for `multiple`
- workload detail page expansion

## 7. Ready-To-Start Checklist

The first batch is ready to begin when:
- shared context fields are agreed,
- namespace rules are accepted,
- Homepage and Workloads are confirmed as the first path,
- and no additional V1 scope is added to this batch.
