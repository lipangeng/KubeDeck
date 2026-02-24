# Repository Guidelines

## Project Structure & Module Organization
This repository is currently in planning-first stage.

- `docs/plans/`: approved architecture and implementation plans (source of truth for upcoming scaffold work).
- `.agents/`: local workflow skills and guardrails (commit format, feature-branch policy, requirements).
- `LICENSE`: project license.

Planned code layout (from approved plans):
- `frontend/shell/` for Vite + TypeScript + MUI shell.
- `backend/` for Go microkernel core and APIs.
- `plugins/templates/` for frontend/backend plugin templates.

## Build, Test, and Development Commands
At this stage, validation is documentation and Git hygiene:

- `rg --files` — list tracked project files quickly.
- `git status --short` — check working tree before/after edits.
- `test -f docs/plans/2026-02-25-kubedeck-architecture-design.md` — confirm baseline design doc exists.

When scaffold code is added, follow commands defined in plan docs (for example `go test ./...`, `npm run test`, `npm run build`).

## Coding Style & Naming Conventions
- Use ASCII by default.
- Keep changes small and scoped to one logical intent.
- Follow planned directory boundaries (`frontend/`, `backend/`, `plugins/`).
- Branch names must be feature-style: `codex/*`, `feature/*`, `feat/*`, or `chore/*`.

## Testing Guidelines
- Prefer TDD for new work: write failing test, implement minimal fix, re-run tests.
- Add tests next to affected modules (Go `_test.go`, frontend `*.test.ts[x]`).
- Do not claim completion without running relevant verification commands.

## Commit & Pull Request Guidelines
- Use Conventional Commits: `<type>(<scope>): <subject>`.
- Keep subject imperative, <= 72 chars.
- Commit body must include bilingual EN/ZH sections:
  - EN: What changed / Why / How to test
  - ZH: 变更内容 / 原因 / 测试方法
- Never commit to `main`/`master`; work only on feature branches and open PRs for human merge.

## Security & Configuration Tips
- Never commit secrets, tokens, private keys, or sensitive `.env` files.
- Treat backend authorization as authoritative; UI hints are not security controls.
