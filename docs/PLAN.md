# KubeDeck PLAN (SSOT)

This document is the single source of truth for scope, architecture direction, and execution rules.

## 1. Scope and Stack

KubeDeck is a multi-cluster, plugin-extensible Kubernetes web control plane.
- Frontend: Vite + TypeScript + MUI
- Backend: Go
- Architecture: frontend shell + backend core, both microkernel + plugin model

## 2. Non-Negotiable Principles

1. Microkernel-first: core provides framework capability; business is pluginized.
2. Feature-branch workflow only; no direct development on `main/master`.
3. Backend authorization is authoritative.
4. Built-in resources and CRDs share unified abstraction.
5. UI extension supports both replacement and slots.
6. Product supports i18n (ZH/EN).

## 3. Current Implementation Baseline (2026-02-25)

Completed and merged:
- Backend module skeleton (`api/auth/core/plugins/registry/storage/webui`)
- Multi-database storage abstraction contracts (default sqlite, mysql/postgres stubs)
- Metadata/resource stub APIs and health probes (`/api/healthz`, `/api/readyz`)
- Frontend shell baseline with MUI
- Theme preference (`system/light/dark`) with persistence
- Sidebar menu composition and grouped rendering (`system/user/dynamic`)
- Reusable page shell components (`ListPageShell`, `DetailPageShell`)
- Plugin templates and manifest validation tests
- Single-executable backend mode (embedded static + optional `--static-dir` override)

## 4. Next Priority (Phase 1 continuation)

- Multi-YAML create dialog with namespace defaulting rules
- Registry payload enrichment (typed resource/page/slot/menu schema alignment)
- Cluster switch lifecycle hardening (state/cache invalidation)
- User preferences API persistence (theme/language/cluster/namespace)

## 5. Canonical Documents

- Architecture design: `docs/plans/2026-02-25-kubedeck-architecture-design.md`
- Implementation plan: `docs/plans/2026-02-25-kubedeck-microkernel-baseline-implementation.md`

## 6. Execution Rule

- One task per feature branch/worktree.
- Each task must pass relevant tests before PR.
- Integration branch (`main`) must remain runnable.
