# Repository Guidelines

## Project Structure & Module Organization
KubeDeck is now a runnable monorepo with backend, frontend, and plugin templates:
- `backend/`: Go service (`cmd/kubedeck`) plus `internal/{api,auth,core,plugins,registry,storage,webui}`.
- `frontend/`: Vite + React + TypeScript + MUI shell (`src/{core,state,sdk,components}`).
- `plugins/templates/`: starter templates for frontend/backend plugins.
- `docs/`: SSOT planning and architecture documents.

## Build, Test, and Development Commands
- `cd backend && go test ./...` — run backend unit tests.
- `cd backend && go build ./...` — verify backend compiles.
- `cd frontend && npm install --registry=https://registry.npmmirror.com` — install frontend deps.
- `cd frontend && npm test -- --run` — run frontend tests.
- `cd frontend && npm run build` — production frontend build.
- Dev run: backend `PORT=8080 go run ./cmd/kubedeck`, frontend `npm run dev`.

## Coding Style & Naming Conventions
- Use ASCII by default.
- Go: keep packages small and cohesive under `backend/internal/*`; tests in `_test.go`.
- Frontend: TypeScript strict mode, colocated tests `*.test.ts[x]`.
- Prefer explicit contract types for API payloads and plugin SDK boundaries.
- Branch names: `feat/*`, `fix/*`, `chore/*` (feature-branch workflow only).

## Testing Guidelines
- Follow TDD for behavior changes (fail first, then implement).
- Always run affected test suites before commit.
- Minimum verification before PR:
  - `cd backend && go test ./...`
  - `cd frontend && npm test -- --run && npm run build`

## Commit & Pull Request Guidelines
- Conventional Commits: `<type>(<scope>): <subject>`.
- Keep subject imperative and concise.
- Commit body must include bilingual EN/ZH sections:
  - EN: What changed / Why / How to test
  - ZH: 变更内容 / 原因 / 测试方法
- PRs should include: scope summary, validation commands run, and UI screenshots for frontend changes.

## Security & Configuration Tips
- Never commit secrets or private credentials.
- Backend auth/authorization is authoritative; frontend permission hints are display-only.
