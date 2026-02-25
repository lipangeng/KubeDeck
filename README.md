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

## Docs Checklist

Required planning artifacts:

- `docs/PLAN.md`
- `docs/plans/2026-02-25-kubedeck-architecture-design.md`
