# KubeDeck Architecture Design (Updated Baseline)

Date: 2026-02-25
Status: Active baseline (updated to match merged code)

## 1. Architecture Overview

KubeDeck adopts microkernel + plugin architecture on both backend and frontend.

- Backend core provides: plugin runtime contracts, registry models, auth/storage abstractions, metadata/resource API stubs, and web UI serving.
- Frontend shell provides: plugin host contracts, route/menu composition, global UI state, and reusable page shells.
- Business capabilities are designed to be delivered as plugins (including built-in functionality).

## 2. Implemented Directory Reality

```txt
backend/
  cmd/kubedeck/main.go
  internal/
    api/ auth/ core/ plugins/ registry/ storage/ webui/
  pkg/sdk/
frontend/
  src/
    components/page-shell/
    core/
    sdk/
    state/
    theme.ts themeMode.ts
plugins/templates/
  frontend-plugin-template/
  backend-plugin-template/
```

Note: historical `frontend/shell/` path in earlier drafts has been flattened to `frontend/`.

## 3. Contract Baseline

Implemented contract surfaces:
- Backend API routes:
  - `GET /api/meta/registry?cluster=<id>`
  - `GET /api/meta/clusters`
  - `GET /api/meta/menus?cluster=<id>`
  - `POST /api/resources/apply`
  - `GET /api/healthz`, `GET /api/readyz`
- Frontend menu contract includes `source`, `order`, `visible`, enabling grouped rendering.
- Storage factory supports `sqlite|mysql|postgres` driver selection (sqlite default path in design).

## 4. UI Kernel Baseline

Implemented:
- MUI `ThemeProvider` and `CssBaseline`
- Theme preference mode: `system/light/dark` + local persistence
- Sidebar with grouped sections: `System Menus`, `User Menus`, `Dynamic Menus`
- Reusable page containers:
  - `ListPageShell`
  - `DetailPageShell`

## 5. Deployment Baseline

- Development: split runtime (frontend dev server + backend API).
- Production: single backend executable can serve embedded static assets.
- Optional static override: `--static-dir` / `STATIC_DIR`.

## 6. Gaps to Close Next

- Replace metadata/resource stubs with typed registry-driven responses.
- Implement multi-YAML apply with namespace defaulting and per-document results.
- Implement real cluster manager/discovery loop and CRD dynamic menus per cluster.
- Add persistent user preference backend APIs.
