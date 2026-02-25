---
name: Git Workflow (Feature Branch Only, No Other Branch Touch)
description: Enforces feature-branch-only development. Never modifies main or other branches. All work must happen on a dedicated feature branch; user merges manually.
---

# Purpose
Enforce a strict Git workflow:
1) Use feature branches for all changes.
2) The agent must not operate on main/master or any non-feature branches.
3) The agent must not merge into main/master; merging is done manually by a human.
4) Each implementation task must use its own dedicated feature branch.

# When to use
Use for any task that changes code or docs under version control.

# Hard rules (must follow)

## G1. Never operate on protected branches
- Do not commit, push, rebase, or reset on:
    - `main`, `master`, `release/*`, `hotfix/*`, or any branch not clearly a feature branch.
- Only operate on a branch matching one of:
    - `feature/*`
    - `feat/*`
    - `codex/*`
    - `chore/*` (if your team uses it for non-feature maintenance)
      If the repo has its own naming convention, follow it.

## G2. Always create or switch to a feature branch before editing
If current branch is not allowed:
1) Fetch latest default branch
2) Create a new feature branch from default branch HEAD
3) Switch to it
   Only then start changes.

Suggested naming:
- `codex/<short-topic>-<YYYYMMDD>`
- `feature/<short-topic>-<ticket-id>`

## G3. No merges into default branch
- Do not merge/rebase into `main/master`.
- Do not fast-forward default branch.
- Do not squash-merge locally into default branch.
- End state should be: feature branch has commits ready for PR.

## G4. PR-first delivery
- Prefer delivering changes via PR.
- If PR tooling is unavailable, still push the feature branch and provide:
    - branch name
    - commit list
    - summary
    - test status
    - review notes

## G5. One task = one branch (required)
- Never implement multiple independent tasks on the same branch.
- For each task:
  1) Create a dedicated feature branch from default branch (or agreed base branch).
  2) Implement only that task's scoped changes.
  3) Push that branch and open/prepare a separate PR.
- Branch naming should include task identity, e.g.:
  - `codex/task-08-shell-plugin-host-20260225`
  - `feature/task-12-verify-baseline`
- Recommended: use `git worktree` so each task has isolated workspace and no branch switching side effects.

# Standard operating procedure (SOP)

## Step 0: Inspect
- `git status`
- `git branch --show-current`
- `git remote -v`
- `git fetch --all --prune`

## Step 1: Ensure feature branch
If NOT on allowed branch:
- Determine default branch: `main` or `master` (use `git remote show origin`).
- Create branch:
    - `git checkout -b codex/<topic>-<YYYYMMDD> origin/<default>`
      (Alternative: `git switch -c ...`)

## Step 1.5: Ensure task-isolated branch
- Before editing, confirm whether current work is a new task.
- If yes, create/switch to a task-specific branch even if already on a feature branch.
- Do not reuse an existing task branch for a different task.

## Step 2: Work only on this branch
- Make edits
- Run formatting/tests if present
- Commit with clear messages

## Step 3: Push feature branch only
- `git push -u origin <feature-branch>`
  Never push to default branch.

## Step 4: Prepare for human merge
- Provide a PR description (What/Why/How to test/Risk).
- Confirm no default-branch modifications occurred.
- Confirm branch maps to exactly one task scope.

# Guardrails / self-checks
Before finishing:
- [ ] `git branch --show-current` is an allowed feature branch.
- [ ] Current branch is dedicated to one task only.
- [ ] `git log --decorate -n 5` shows commits only on feature branch.
- [ ] `git status` clean or intentional.
- [ ] No operations were performed on `main/master` or other branches.

# If the user asks to commit to main/master
Refuse that specific action and instead:
- create/switch to feature branch
- commit there
- push feature branch
- request manual merge via PR
