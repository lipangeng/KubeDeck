# KubeDeck Phase-1 Continuation Implementation Plan

Date: 2026-02-25
Status: Active

## 1. Purpose

This plan continues from the merged baseline and focuses on turning current stubs into usable Phase-1 workflows.

## 2. Current Baseline (Already Implemented)

- Backend skeleton and contracts: `api/auth/core/plugins/registry/storage/webui`
- Frontend shell with MUI, theme preference, grouped menu rendering
- Plugin templates and manifest checks
- Single-binary backend static serving path

## 3. Phase-1 Remaining Tasks

### Task A: Multi-YAML Apply Flow (Frontend + Backend)

Scope:
- Frontend create dialog: namespace selector + YAML input
- Namespace defaulting rule:
  - list namespace selected -> use selected namespace
  - list namespace all -> use last used namespace, else `default`
- Backend `POST /api/resources/apply`: multi-document split and per-document result contract

Acceptance:
- Supports `---` separated documents
- Returns per-document success/failure/reason
- Explicit partial-failure response shape

### Task B: Registry and Menu Contract Hardening

Scope:
- Replace `resourceTypes` string stub with typed registry payload
- Align backend/TS menu and page metadata schema
- Keep frontend menu renderer fully typed

Acceptance:
- `GET /api/meta/registry` returns stable typed structure
- Frontend compiles without `any` fallback for registry/menu payload

### Task C: Cluster Context Lifecycle

Scope:
- Frontend cluster switch state model with cache invalidation hooks
- Backend cluster list endpoint remains source of selectable clusters
- Menu and registry reload on cluster change

Acceptance:
- No stale data bleed across cluster switch
- Cluster change triggers metadata refresh deterministically

### Task D: User Preference Backend Persistence

Scope:
- Add `/api/user/preferences` read/write stubs backed by storage abstractions
- Persist theme/language/default cluster/last namespace

Acceptance:
- Frontend can hydrate theme and defaults from backend response
- Storage driver abstraction remains unchanged

## 4. Engineering Rules

- One task per feature branch/worktree.
- TDD for behavior changes.
- Required verification before PR:
  - `cd backend && go test ./...`
  - `cd frontend && npm test -- --run && npm run build`
- Use npm mirror for installs:
  - `npm install --registry=https://registry.npmmirror.com`

## 5. Suggested Execution Order

1. Task B (contracts first)
2. Task C (cluster lifecycle)
3. Task A (multi-YAML create/apply)
4. Task D (preference persistence)
