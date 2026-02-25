# KubeDeck

KubeDeck is a plugin-extensible Kubernetes web control plane built with a microkernel architecture.

## Repositories and Docs

- SSOT plan: `docs/PLAN.md`
- Architecture baseline: `docs/plans/2026-02-25-kubedeck-architecture-design.md`
- Implementation task plan: `docs/plans/2026-02-25-kubedeck-microkernel-baseline-implementation.md`

## Build and Test

Backend:

- `cd backend && GOCACHE=/tmp/go-build go test ./...`
- `cd backend && GOCACHE=/tmp/go-build go build ./...`

Frontend:

- `cd frontend && npm install --registry=https://registry.npmmirror.com`
- `cd frontend && npm test -- --run`
- `cd frontend && npm run build`

## Runtime Configuration

### Development (recommended)

Run backend on `:8080` and frontend on Vite dev server. Frontend uses `/api` proxy to backend.

- Backend: `cd backend && PORT=8080 go run ./cmd/kubedeck`
- Frontend: `cd frontend && npm run dev`

If backend is not on `:8080`, set:

- `VITE_BACKEND_TARGET=http://127.0.0.1:<port>`

Example:

- `cd frontend && VITE_BACKEND_TARGET=http://127.0.0.1:18080 npm run dev`

### Production

Use one of these deployment models:

1. Same-origin reverse proxy (preferred): expose frontend and proxy `/api` to backend service.
2. Split origin: frontend and backend on different domains/ports, and gateway/reverse proxy handles routing and CORS policy.

## Docs Checklist

Required planning artifacts:

- `docs/PLAN.md`
- `docs/plans/2026-02-25-kubedeck-architecture-design.md`
