# KubeDeck V1 Development Plan

## 1. Purpose

This document is the current execution plan for V1.

It replaces earlier task-breakdown, first-patch, checklist, and code-mapping drafts with one development-facing plan.

## 2. Plan Principle

V1 development must follow architecture-first sequencing:

1. runtime foundation,
2. menu composition,
3. resource-page shell,
4. first real workflow,
5. controlled specialization.

Do not reverse this order by rebuilding page flows outside the current architecture.

## 3. Stage 1: Stabilize Runtime Foundation

Goal:

- preserve the microkernel and plugin runtime as the base for all later work.

Deliverables:

- clean frontend and backend kernel contracts,
- built-in and plugin capability registration through the same runtime family,
- metadata and execution entry points that fit capability composition,
- minimum i18n copy boundary for new UI text.

Exit criteria:

- built-in features no longer depend on shell-only assumptions,
- plugin capability flow remains intact,
- new UI text can enter through an i18n boundary.

## 4. Stage 2: Implement Menu Composition

Goal:

- move the left navigation from raw registrations to composed menu results.

Deliverables:

- default menu blueprint,
- mount model for built-in workflows, CRDs, and plugin routes,
- cluster-aware availability resolution,
- disabled state for configured-but-unavailable entries,
- `CRDs` fallback entry,
- room for future global and cluster-local overrides.

Exit criteria:

- the rendered navigation is a composed result,
- explicit CRD and plugin mounts fit the same runtime family,
- unavailable configured entries remain stable and visible.
- work, system, and cluster menu spaces share one dynamic menu model.

For the next implementation sequence inside this stage, use:

- `docs/product/releases/v1/scoped-menu-settings-implementation-plan.md`

## 5. Stage 3: Implement Resource Page Shell

Goal:

- establish one shared resource-page model for built-in resources and CRDs.

Deliverables:

- `ResourcePageShell`,
- default `Overview` and `YAML` tabs,
- baseline resource identity and context handling,
- extension capability types for `tab`, `tab-replace`, and `page-takeover`,
- the first resource views delivered through that model.

Exit criteria:

- first-wave resources resolve through the shell,
- YAML is a real baseline tab,
- the model allows later specialization without breaking the shell.

## 6. Stage 4: Reconnect The First Real Workflow

Goal:

- deliver the first usable workflow through the new menu and page systems.

Deliverables:

- homepage that emphasizes task entry,
- composed entry into `Workloads`,
- real workload list,
- one real action path such as create, apply, or edit,
- clear result handling,
- continuity across cluster, namespace, and prior path.

Exit criteria:

- a user can complete the first workflow end-to-end on the current runtime shape,
- no critical path depends on deprecated shell patterns.

## 7. Stage 5: Controlled Specialization

Goal:

- prove the architecture can support richer capability-specific experiences without collapsing back into special cases.

Deliverables:

- one or more specialized tabs for a built-in resource,
- one CRD-backed path proving the shared model,
- one plugin-backed entry or extension that fits the same system.

Exit criteria:

- specialization happens through approved extension types,
- no new parallel page model is introduced.

## 8. Explicit Deferred Work

These items remain out of scope for this plan:

- AI workflow support,
- generalized chat surfaces,
- plugin marketplace flows,
- complete menu customization productization,
- broad block-level resource-page extension,
- broad multi-domain expansion before the first workflow is stable.

## 9. Verification Rule

Each stage should close with:

- affected frontend tests passing,
- affected backend tests passing,
- build verification,
- and documentation updated if the architecture or boundary changed.

## 10. Ready For Implementation Check

Work can proceed directly from this plan when the following remain true:

- current implementation still targets the runtime defined in `microkernel-plugin-runtime`,
- menu work still targets `cluster-aware-menu-composition`,
- resource-page work still targets `frontend-resource-page-extension-model`,
- and V1 scope still matches `implementation-boundary`.

For the immediate remediation sequence before broader V1 feature work, use:

- `docs/product/releases/v1/foundation-architecture-remediation-plan.md`
