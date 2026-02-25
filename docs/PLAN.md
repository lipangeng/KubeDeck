# KubeDeck PLAN (SSOT)

This file is the single source of truth for KubeDeck planning and phased implementation.

## Scope

KubeDeck is a multi-cluster, plugin-extensible Kubernetes web control plane.

- Frontend: Vite + TypeScript + MUI
- Backend: Go
- Architecture: Microkernel + Plugin (frontend and backend)

## Non-Negotiable Principles

1. Microkernel-first: core provides framework capability only.
2. Feature-branch workflow only.
3. Backend authorization is authoritative.
4. Built-in resources and CRDs share one abstraction model.
5. UI extension supports both replacement and slots.
6. Product supports i18n (ZH/EN).

## MVP Direction

Phase 1 baseline focuses on:

- Frontend/backend skeleton and plugin contracts
- Registry data model and metadata API stubs
- Generic resource workflows and menu composition baseline
- Plugin templates for frontend/backend

## Phase Roadmap

- Phase 1: baseline skeleton + contracts + templates
- Phase 2: multi-cluster context, page replacement, slot injection
- Phase 3: global search, resource graph, terminal, OAuth/RBAC management UI

## Canonical Design Documents

- Architecture design:
  - `docs/plans/2026-02-25-kubedeck-architecture-design.md`
- Implementation plan:
  - `docs/plans/2026-02-25-kubedeck-microkernel-baseline-implementation.md`

## Execution Rule

Implementation must follow task isolation:

- one task per branch/worktree
- each task merged back to integration validation branch after passing checks
- integration validation branch must stay runnable
