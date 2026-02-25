# KubeDeck

KubeDeck is a plugin-extensible Kubernetes web control plane built on a microkernel + plugin architecture.

## Current Status

Implemented baseline (runnable):
- Go backend with metadata/resource stub APIs and health probes
- Vite + TypeScript + MUI frontend shell
- Sidebar menu composition (`system + user + dynamic`) and grouped rendering
- Theme preference (`system/light/dark`) with local persistence
- Single-binary backend mode with embedded static UI and optional static override

## Repository Layout

- `backend/`: Go service entrypoint, API handlers, registry/auth/storage/plugin contracts
- `frontend/`: Vite React shell, MUI UI kernel, state/core/sdk contracts
- `plugins/templates/`: frontend/backend plugin templates
- `docs/`: SSOT plan and architecture/implementation docs

## Build and Test

Backend:
- `cd backend && go test ./...`
- `cd backend && go build ./...`

Frontend:
- `cd frontend && npm install --registry=https://registry.npmmirror.com`
- `cd frontend && npm test -- --run`
- `cd frontend && npm run build`

## Run (Development)

Recommended split runtime:
- Backend: `cd backend && PORT=8080 go run ./cmd/kubedeck`
- Frontend: `cd frontend && npm run dev`

Optional dev proxy target override:
- `cd frontend && VITE_BACKEND_TARGET=http://127.0.0.1:18080 npm run dev`

## Run (Production)

### Single Executable

1. Build frontend assets:
   - `cd frontend && npm install --registry=https://registry.npmmirror.com`
   - `cd frontend && npm run build`
2. Package assets for embed:
   - `rm -rf backend/internal/webui/dist`
   - `mkdir -p backend/internal/webui/dist`
   - `cp -R frontend/dist/. backend/internal/webui/dist/`
3. Build and run backend:
   - `cd backend && go build -o kubedeck ./cmd/kubedeck`
   - `cd backend && ./kubedeck --port 8080`

Runtime options:
- `--port` / `PORT`
- `--static-dir` / `STATIC_DIR` (override embedded static directory)

### Reverse Proxy

Same-origin reverse proxy is recommended for production traffic routing.

## API Baseline

- `GET /api/meta/registry?cluster=<id>`
- `GET /api/meta/clusters`
- `GET /api/meta/menus?cluster=<id>`
- `POST /api/resources/apply`
- `GET /api/healthz`, `GET /api/readyz`
