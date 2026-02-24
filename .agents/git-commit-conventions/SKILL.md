---
name: Git Commit Conventions (Conventional Commits + Auto Message)
description: Enforces commit standards and generates high-quality commit messages (Conventional Commits, bilingual summary, and scoped changes). Must run on feature branches only and never commit to protected branches.
---

# Purpose
Standardize git commits so they are readable, reviewable, and automation-friendly.

This skill enforces:
1) Conventional Commits format.
2) Small, focused commits.
3) Bilingual summary in commit body (EN + ZH).
4) A repeatable process for generating commit messages.

# Pre-req
This skill assumes the Git workflow policy is active:
- Work only on allowed feature branches.
- Never commit/push to main/master or any non-feature branches.

# Hard rules

## C1. Conventional Commits (required)
Commit subject MUST follow:
`<type>(<scope>): <subject>`

Where:
- type ∈ `feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert`
- scope: short module/package name (optional but recommended)
- subject: imperative, <= 72 chars, no trailing period

Examples:
- `feat(api): add bilingual response headers`
- `fix(auth): prevent token refresh race`
- `docs(readme): add EN and ZH setup guides`
- `chore(codex): add commit conventions skill`

## C2. Keep commits small and coherent
- One logical change per commit.
- Do NOT mix refactor + feature + formatting in one commit.
- If changes are large, split into multiple commits by intent.

## C3. Commit body must include bilingual summary
Commit body must include both EN and ZH summaries, following this template:

EN:
- What changed:
- Why:
- How to test:

ZH:
- 变更内容：
- 原因：
- 测试方法：

If "How to test" is not applicable, write `N/A`.

## C4. No secret or credential material
Never commit:
- tokens, private keys, `.env` with secrets
- internal endpoints/credentials that should not be public

## C5. Auto-generate commit message before committing
Before running `git commit`, the agent must:
- Review `git status` and `git diff --stat`
- Summarize the changes
- Propose the commit subject + body (EN+ZH) per the template
- Then commit using the proposed message

# Standard operating procedure (SOP)

1) Verify branch
- `git branch --show-current`
- Must be an allowed feature branch (e.g. `feature/*`, `feat/*`, `codex/*`, `chore/*`)

2) Stage intentionally
- Prefer `git add -p` (patch mode) for precision
- Avoid staging unrelated changes

3) Generate commit message
- Determine type/scope from files changed and intent
- Draft subject <= 72 chars
- Draft bilingual body (EN + ZH) with What/Why/How-to-test

4) Commit
- Use `git commit` with the generated message
- Re-check: `git status` should be clean or intentionally remaining

# Guidance for selecting type
- feat: new user-facing capability
- fix: bug fix
- docs: documentation only
- refactor: behavior-preserving code restructure
- chore: tooling, maintenance, skills, configs
- ci/build: pipelines, build scripts
- test: test-only changes

# Guardrails / self-checks
Before finalizing:
- [ ] Subject matches `<type>(<scope>): <subject>`
- [ ] Subject <= 72 chars
- [ ] Body contains EN and ZH sections
- [ ] Commit does not include secrets
- [ ] Commit is on an allowed feature branch