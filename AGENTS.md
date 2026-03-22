# Repository Guidelines

## End-to-End Development Agent Protocol

**自主执行模式：**
1. **任务拆解** — 读取需求后自动分解为可执行子任务，使用 TodoWrite 追踪进度
2. **上下文感知** — 优先读取相关文件理解现有模式，保持代码风格一致
3. **直接实现** — 除高风险操作（删除数据、修改生产配置、引入重大依赖）外，不等待确认
4. **自动验证** — 每次修改后自动运行相关测试和构建命令
5. **自我修复** — 测试失败时自动分析错误并修复，最多尝试 3 次后报告
6. **提交规范** — 完成后按 Conventional Commits 格式提交（如用户要求）

**执行优先级：**
- 读取相关技能 → 拆解任务 → 实现 → 测试 → 修复 → 完成

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

## Documentation Guidelines
- User-facing and contributor-facing documentation must be maintained as two standalone files: one Chinese version and one English version.
- Do not mix Chinese and English in the same canonical document, except for required identifiers, commands, or API names.
- When creating new docs, add both language versions together in the same change.
- When updating an existing doc, update the corresponding Chinese and English versions in the same change and keep their structure, meaning, and examples aligned.
- If one language version is intentionally deferred, record the gap and reason in the PR description before merge.

## Security & Configuration Tips
- Never commit secrets or private credentials.
- Backend auth/authorization is authoritative; frontend permission hints are display-only.
