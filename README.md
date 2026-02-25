# KubeDeck

KubeDeck is a plugin-extensible Kubernetes web control plane built with a microkernel architecture.

## Repositories and Docs

- SSOT plan: `docs/PLAN.md`
- Architecture baseline: `docs/plans/2026-02-25-kubedeck-architecture-design.md`
- Implementation task plan: `docs/plans/2026-02-25-kubedeck-microkernel-baseline-implementation.md`

## Build and Test

Backend:

- `cd backend && go test ./...`
- `cd backend && go build ./...`

Frontend:

- `cd frontend && npm install --registry=https://registry.npmmirror.com`
- `cd frontend && npm test -- --run`
- `cd frontend && npm run build`

## Runtime Configuration

### Development (recommended)

Use split runtime: backend API + frontend dev server. This is the default developer workflow.

- Backend: `cd backend && PORT=8080 go run ./cmd/kubedeck`
- Frontend: `cd frontend && npm run dev`

If backend is not on `:8080`, set:

- `VITE_BACKEND_TARGET=http://127.0.0.1:<port>`

Example:

- `cd frontend && VITE_BACKEND_TARGET=http://127.0.0.1:18080 npm run dev`

### Production (single executable)

Backend supports serving embedded UI static assets for single-binary deployment.

1. Build frontend assets:
   - `cd frontend && npm install --registry=https://registry.npmmirror.com`
   - `cd frontend && npm run build`
2. Package frontend dist into backend embed directory (release/build stage):
   - `rm -rf backend/internal/webui/dist`
   - `mkdir -p backend/internal/webui/dist`
   - `cp -R frontend/dist/. backend/internal/webui/dist/`
3. Build backend binary:
   - `cd backend && go build -o kubedeck ./cmd/kubedeck`
4. Run single executable:
   - `cd backend && ./kubedeck --port 8080`

Runtime options:

- `--port` or `PORT`: HTTP listen port (default `8080`)
- `--static-dir` or `STATIC_DIR`: optional local static directory override

Static override example:

- `cd backend && ./kubedeck --port 8080 --static-dir /opt/kubedeck/static`

If `--static-dir` is not provided, server uses embedded assets from `backend/internal/webui/dist`.

Notes:

- Development does not require embedding frontend build artifacts.
- The repository keeps a minimal placeholder `backend/internal/webui/dist/index.html` so backend can still start independently.

### Production (reverse proxy mode)

Alternative deployments are still supported:

1. Same-origin reverse proxy (preferred): expose one origin and route `/api` to backend.
2. Split origin: frontend and backend on different domains/ports with gateway-level routing and CORS policy.

## Docs Checklist

Required planning artifacts:

- `docs/PLAN.md`
- `docs/plans/2026-02-25-kubedeck-architecture-design.md`
