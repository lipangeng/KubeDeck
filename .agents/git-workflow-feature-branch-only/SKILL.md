---
name: Git Workflow (Feature Branch Only, No Other Branch Touch)
description: Enforces feature-branch-only development. Never modifies main or other branches. All work must happen on a dedicated feature branch; user merges manually.
---

# Purpose
Enforce a strict Git workflow:
1) Use feature branches for all changes.
2) The agent must not operate on main/master or any non-feature branches.
3) The agent must not merge into main/master; merging is done manually by a human.
4) Default to a single integration feature branch for continuous development; task branches are optional for high-risk or parallel work.

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

## G5. Integration branch first (default)
- Use one long-lived integration feature branch as the main development line (for example `codex/merge-validation-*`).
- Keep this integration branch deployable/runnable after each merged change.
- Use per-task branches only when needed:
  1) high-risk refactors,
  2) large parallel work,
  3) experiments likely to be discarded.
- If task branches are used, merge/cherry-pick back into integration branch quickly and delete task branches after integration.

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

## Step 1.5: Decide branch strategy
- Default: stay on the integration feature branch.
- If change is risky/parallel, create a short-lived task branch from integration branch.
- After validation, merge/cherry-pick task branch back to integration branch and clean it up.

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
- Confirm integration branch is green (tests/build pass) before PR.

# Guardrails / self-checks
Before finishing:
- [ ] `git branch --show-current` is an allowed feature branch.
- [ ] If using integration branch: it is runnable and green.
- [ ] If using task branch: it has been integrated back and is ready for cleanup.
- [ ] `git log --decorate -n 5` shows commits only on feature branch.
- [ ] `git status` clean or intentional.
- [ ] No operations were performed on `main/master` or other branches.

# If the user asks to commit to main/master
Refuse that specific action and instead:
- create/switch to feature branch
- commit there
- push feature branch
- request manual merge via PR
