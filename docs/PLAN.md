# KubeDeck PLAN

This document is the current top-level index for product direction, architecture focus, and development entry.

## 1. Current Product Definition

KubeDeck is a multi-cluster, plugin-extensible Kubernetes web control plane for platform teams.

Its current product direction is defined by four ideas:

1. multi-cluster operating context is a first-class concern,
2. built-in resources and CRDs share one resource model,
3. frontend and backend must follow a microkernel extension architecture,
4. navigation and resource pages must be composed rather than hard-coded.

KubeDeck is not trying to be another generic Kubernetes dashboard.

## 2. Current Development Stage

The project is now in development-mode preparation.

That means:

- product and architecture intent have been re-clarified,
- legacy planning drafts have been removed from the main reading path,
- and the next implementation work must follow the current architecture set instead of older shell-first assumptions.

## 3. Current Priorities

Development priority order is:

1. preserve the microkernel and plugin runtime shape,
2. implement cluster-aware menu composition,
3. implement the resource-page shell and tab-based extension model,
4. restore the first real user workflow on top of those foundations,
5. expand plugin-powered capabilities after the first workflow is stable.

## 4. Current Non-Priorities

The following are intentionally not current priorities:

- AI-assisted workflow delivery,
- generalized chat surfaces,
- full i18n product delivery,
- large dashboard-style observability surfaces,
- broad visual polish before workflow completion.

## 5. Development Entry

Start here when entering development mode:

- Development mode guide: `docs/product/development-mode.md`
- Development mode guide (ZH): `docs/product/development-mode.zh.md`

## 6. Current Core Documents

### Product

- Product requirements: `docs/product/product-requirements.md`
- Product requirements (ZH): `docs/product/product-requirements.zh.md`

### Architecture

- Microkernel and plugin runtime: `docs/product/architecture/microkernel-plugin-runtime.md`
- Microkernel and plugin runtime (ZH): `docs/product/architecture/microkernel-plugin-runtime.zh.md`
- Cluster-aware menu composition: `docs/product/architecture/cluster-aware-menu-composition.md`
- Cluster-aware menu composition (ZH): `docs/product/architecture/cluster-aware-menu-composition.zh.md`
- Frontend resource page extension model: `docs/product/architecture/frontend-resource-page-extension-model.md`
- Frontend resource page extension model (ZH): `docs/product/architecture/frontend-resource-page-extension-model.zh.md`

### V1 Delivery

- V1 implementation boundary: `docs/product/releases/v1/implementation-boundary.md`
- V1 implementation boundary (ZH): `docs/product/releases/v1/implementation-boundary.zh.md`
- V1 development plan: `docs/product/releases/v1/v1-development-plan.md`
- V1 development plan (ZH): `docs/product/releases/v1/v1-development-plan.zh.md`
- Foundation architecture remediation plan: `docs/product/releases/v1/foundation-architecture-remediation-plan.md`
- Foundation architecture remediation plan (ZH): `docs/product/releases/v1/foundation-architecture-remediation-plan.zh.md`

### Archive

- Archive index: `docs/archive/README.md`
- Archive index (ZH): `docs/archive/README.zh.md`

## 7. Reading Order

Recommended reading order for active development:

1. `docs/PLAN.md`
2. `docs/product/product-requirements.md`
3. `docs/product/development-mode.md`
4. `docs/product/architecture/microkernel-plugin-runtime.md`
5. `docs/product/architecture/cluster-aware-menu-composition.md`
6. `docs/product/architecture/frontend-resource-page-extension-model.md`
7. `docs/product/releases/v1/implementation-boundary.md`
8. `docs/product/releases/v1/v1-development-plan.md`

## 8. Documentation Rule

Only the documents listed in this PLAN are considered the current development path.

Older planning drafts remain available in `docs/archive/` for reference, but they are not the active source of implementation decisions.
