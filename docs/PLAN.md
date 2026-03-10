# KubeDeck PLAN

This document is the single source of truth for product direction, scope boundaries, and delivery priorities.

## 1. Product Definition

KubeDeck is a multi-cluster, plugin-extensible Kubernetes web control plane for platform teams.

It is not primarily intended to replace `kubectl` feature-for-feature. Its value is to provide:
- a consistent multi-cluster operating context,
- a unified control surface for built-in resources and CRDs,
- and a plugin-capable shell for team-specific workflows.

## 2. Product Principles

1. User task first: UI and APIs must serve real operating workflows before exposing framework capability.
2. Multi-cluster first: cluster context is a primary user concern, not a secondary filter.
3. Unified resource model: built-in resources and CRDs must share one mental model and extension path.
4. Plugin extensibility is product value: plugins must extend meaningful workflows, not only register metadata.
5. Backend authority: authorization and policy enforcement belong to the backend.
6. Bilingual documentation: product and contributor documents must be maintained in separate EN and ZH versions.

## 3. Target Users

Primary users:
- platform engineers,
- DevOps / SRE operators,
- and delivery engineers who work across multiple Kubernetes clusters.

These users need one web control plane that can carry both standard resource operations and team-specific workflows.

## 4. Core Product Scenarios

KubeDeck should first succeed in these scenarios:
- switch into the correct cluster context quickly,
- enter a real resource or workflow area immediately,
- browse and act on Kubernetes resources with consistent context,
- and extend that workflow with team-specific pages, panels, actions, or menus.

## 5. MVP Workflow

The first workflow that must become real is:

1. User enters the system.
2. User confirms or changes the active cluster.
3. User enters a concrete resource domain.
4. User sees a real resource list.
5. User performs one standard operation such as create or apply.
6. User receives a clear result.
7. User remains in the same cluster and namespace context for follow-up work.

If this workflow does not work end-to-end, the product is still a shell rather than a usable control plane.

## 6. Homepage Responsibility

The homepage should primarily help the user:
- confirm current operating context,
- identify the next likely task,
- and enter that workflow immediately.

The homepage should not primarily behave like a runtime diagnostics screen.

## 7. Current Reality

As of the current baseline:
- backend metadata and resource APIs are still stubs,
- frontend navigation is present but not yet task-complete,
- plugin contracts and templates exist,
- but the product does not yet deliver the first required user workflow end-to-end.

This means the main gap is not visual polish. The main gap is task completion.

AI may become a later enhancement layer after the first real workflow is complete, but it is not part of the current MVP.

## 8. Priority Order

Near-term priorities must follow this order:

1. clarify and preserve product requirements,
2. make the first user workflow complete,
3. align homepage and navigation with that workflow,
4. harden cluster context and state lifecycle,
5. then expand plugin-driven workflows.

## 9. Canonical Documents

- Product requirements clarification: `docs/product/product-requirements.md`
- Product requirements clarification (ZH): `docs/product/product-requirements.zh.md`
- First core workflow draft: `docs/product/workflows/first-workflow.md`
- First core workflow draft (ZH): `docs/product/workflows/first-workflow.zh.md`
- Shared working context model spec: `docs/product/workflows/shared-working-context-model.md`
- Shared working context model spec (ZH): `docs/product/workflows/shared-working-context-model.zh.md`
- Shared working context state schema draft: `docs/product/workflows/shared-working-context-state-schema.md`
- Shared working context state schema draft (ZH): `docs/product/workflows/shared-working-context-state-schema.zh.md`
- Shared working context events and update rules: `docs/product/workflows/shared-working-context-events-and-update-rules.md`
- Shared working context events and update rules (ZH): `docs/product/workflows/shared-working-context-events-and-update-rules.zh.md`
- Shared working context implementation mapping draft: `docs/product/workflows/shared-working-context-implementation-mapping.md`
- Shared working context implementation mapping draft (ZH): `docs/product/workflows/shared-working-context-implementation-mapping.zh.md`
- Shared working context file responsibility draft: `docs/product/workflows/shared-working-context-file-responsibility-draft.md`
- Shared working context file responsibility draft (ZH): `docs/product/workflows/shared-working-context-file-responsibility-draft.zh.md`
- Homepage and Workloads field-level IA draft: `docs/product/workflows/field-level-ia.md`
- Homepage and Workloads field-level IA draft (ZH): `docs/product/workflows/field-level-ia.zh.md`
- Namespace context decision draft: `docs/product/workflows/namespace-context-decision.md`
- Namespace context decision draft (ZH): `docs/product/workflows/namespace-context-decision.zh.md`
- V1 implementation boundary: `docs/product/releases/v1/implementation-boundary.md`
- V1 implementation boundary (ZH): `docs/product/releases/v1/implementation-boundary.zh.md`
- V1 implementation task breakdown: `docs/product/releases/v1/implementation-task-breakdown.md`
- V1 implementation task breakdown (ZH): `docs/product/releases/v1/implementation-task-breakdown.zh.md`
- V1 initial development tasks: `docs/product/releases/v1/initial-development-tasks.md`
- V1 initial development tasks (ZH): `docs/product/releases/v1/initial-development-tasks.zh.md`
- Task A code change plan: `docs/product/releases/v1/task-a-code-change-plan.md`
- Task A code change plan (ZH): `docs/product/releases/v1/task-a-code-change-plan.zh.md`
- Task A first patch scope: `docs/product/releases/v1/task-a-first-patch-scope.md`
- Task A first patch scope (ZH): `docs/product/releases/v1/task-a-first-patch-scope.zh.md`
- Task A first patch checklist: `docs/product/releases/v1/task-a-first-patch-checklist.md`
- Task A first patch checklist (ZH): `docs/product/releases/v1/task-a-first-patch-checklist.zh.md`
- Architecture design: `docs/plans/2026-02-25-kubedeck-architecture-design.md`
- Implementation plan: `docs/plans/2026-02-25-kubedeck-microkernel-baseline-implementation.md`
- PLAN (ZH): `docs/PLAN.zh.md`

## 10. Execution Rules

- One task per feature branch or worktree.
- Keep the repository runnable after each merged change.
- Validate affected backend and frontend tests before PR.
- Update both EN and ZH documents together when requirements or plans change.
