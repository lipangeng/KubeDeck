# KubeDeck Development Mode

## 1. Purpose

This document defines how to enter active development using the current product and architecture set.

It replaces the earlier draft-heavy workflow where planning, patch scope, and exploratory notes all lived on the main path.

## 2. Current Working Rule

Development must now proceed from the current architecture set, not from legacy shell-first code paths and not from archived planning drafts.

The active flow is:

1. read the current core documents,
2. make changes that fit the microkernel runtime,
3. align UI changes with menu composition and resource-page extension rules,
4. deliver V1 in bounded increments,
5. keep older drafts out of the implementation decision path.

## 3. Required Reading

Before starting feature work, read these documents in order:

1. `docs/PLAN.md`
2. `docs/product/product-requirements.md`
3. `docs/product/architecture/microkernel-plugin-runtime.md`
4. `docs/product/architecture/cluster-aware-menu-composition.md`
5. `docs/product/architecture/frontend-resource-page-extension-model.md`
6. `docs/product/releases/v1/implementation-boundary.md`
7. `docs/product/releases/v1/v1-development-plan.md`

## 4. Development Guardrails

All new development must follow these rules:

### 4.1 Do Not Rebuild The Old Shell

Do not grow the application by adding more hard-coded pages, menus, or workflow branches outside the microkernel runtime.

### 4.2 Menus Must Be Composed

Do not treat the left navigation as a simple render of registered menu items.

Menu work must fit the blueprint, mount, availability, and override model.

### 4.3 Resource Pages Must Use The Shared Model

Do not build unrelated page systems for built-in resources and CRDs.

Resource UI must fit the shared shell and tab-first extension model unless takeover is explicitly justified.

### 4.4 V1 Must Stay Bounded

Do not expand into AI, broad observability, marketplace features, or generalized customization before the first workflow is stable.

### 4.5 I18n Must Stay Ready

Full i18n delivery is deferred, but new user-facing copy must use a locale boundary instead of uncontrolled hard-coded strings.

## 5. Development Sequence

The current implementation sequence is:

1. stabilize the microkernel and plugin runtime,
2. implement cluster-aware menu composition,
3. implement resource-page shell and tab capabilities,
4. reconnect the first real workflow on top of those systems,
5. expand plugins and richer resource specialization after V1 is proven.

## 6. What Counts As Current Truth

Current implementation truth comes from:

- `PLAN`
- current product requirements
- current architecture documents
- current V1 boundary and development plan

Archived documents do not define current truth.

## 7. What To Do With Archived Documents

Archived documents can still be used for:

- historical rationale,
- old option comparisons,
- previous draft terminology,
- or migration context.

They must not be used as the primary basis for new implementation decisions unless their content is brought back into a current core document.
