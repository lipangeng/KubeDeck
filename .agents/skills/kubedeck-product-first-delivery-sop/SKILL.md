---
name: KubeDeck Product-First Delivery SOP
description: Use when working on KubeDeck features from requirement clarification through implementation. Enforces a product-first workflow: clarify user task, define workflow and IA, set V1 boundaries, then implement in small batches. Also recommends which supporting skills to use at each stage, especially notion-spec-to-implementation, playwright, openai-docs, gh-address-comments, and gh-fix-ci, with room to substitute equivalent obra/superpowers skills when your local bundle differs.
---

# Purpose

This skill standardizes how feature work should be done in KubeDeck.

It exists to prevent a common failure mode:
- jumping from a vague idea directly into UI or code,
- expanding scope before the first workflow is complete,
- and adding advanced capabilities before the core task path is usable.

# Skill Location

This is a project-local skill and should follow the Codex repository convention:
- store project skills under `.agents/skills/<skill-name>/SKILL.md`
- do not move this project SOP into the global home-directory skills unless it is intentionally promoted to a cross-project skill

For KubeDeck, this SOP should stay project-local because it encodes repository-specific product and delivery discipline.

# When To Use

Use this skill for:
- new feature exploration,
- product or UI rework,
- workflow design,
- V1 scope definition,
- implementation planning,
- and feature delivery work that should follow the product-first method already used in this repository.

Do not use this skill for:
- isolated bug fixes with clear scope,
- pure dependency or build maintenance,
- or one-off mechanical edits that do not affect product behavior.

# Core Method

Always move in this order:

1. clarify the original requirement
2. define the first real user workflow
3. define information architecture and context rules
4. define V1 boundary
5. break V1 into implementation tasks
6. implement only the first batch
7. validate the workflow before expanding scope

Do not skip ahead to implementation unless the earlier stage already exists in repository docs or is already clearly agreed.

# Current KubeDeck Working Style

For KubeDeck, the current expected working style is:
- product-first, not component-first
- workflow-first, not dashboard-first
- context continuity first, especially cluster and namespace
- V1 boundary first, before AI, wide plugin expansion, or polish
- documentation-first for unclear features

This means:
- Homepage exists to route users into work
- `Workloads` is the first real workspace
- namespace is a cross-page working context, not a trivial local filter
- AI is deferred until the non-AI workflow is complete

# Process Document Structure

Do not keep growing product-process documents flat under `docs/` forever.

When the current document set stabilizes, reorganize them into a clearer structure like:

- `docs/product/`
  Product intent and requirement-level documents
- `docs/product/workflows/`
  Workflow, state, namespace, IA, and page-spec documents
- `docs/product/releases/v1/`
  V1 boundary, breakdown, and initial delivery documents
- `docs/plans/`
  Architecture and implementation planning documents

Recommended mapping for the current document set:
- `product-requirements*.md` -> `docs/product/`
- `first-workflow*.md`
- `field-level-ia*.md`
- `low-fidelity-page-spec*.md`
- `namespace-context-decision*.md`
  These should move under `docs/product/workflows/`
- `v1-implementation-boundary*.md`
- `v1-implementation-task-breakdown*.md`
- `v1-initial-development-tasks*.md`
  These should move under `docs/product/releases/v1/`

Directory-structure rule:
- do not reorganize documents repeatedly during active clarification
- do reorganize once the workflow model, V1 boundary, and first task batch are stable enough
- keep EN and ZH documents mirrored in the same directory structure

# Standard SOP

## Stage 1: Requirement Clarification

Goal:
- determine who the user is, what job they are trying to do, and why KubeDeck should exist for that job

Actions:
- inspect current docs and code reality
- separate product goal from current implementation
- rewrite requirement in user-task language

Primary supporting skill:
- `notion-spec-to-implementation`

OpenAI/skills note:
- use this when you need structured requirement extraction, clarifications, and a plan from a spec

Superpowers note:
- if your obra/superpowers bundle has a product-spec or PRD-analysis equivalent, it may be used here as a substitute

Output:
- requirement clarification doc

## Stage 2: First Workflow Definition

Goal:
- define the first task that must work end-to-end

Actions:
- identify the first user path
- define entry, workspace, action, result, and return
- define success/failure expectations

Primary supporting skill:
- `notion-spec-to-implementation`

Optional supporting skill:
- `playwright`
  Use only to verify current UI behavior, not to invent the workflow

Output:
- first workflow doc

## Stage 3: Information Architecture And Context Rules

Goal:
- define what the user must see and what context must persist

Actions:
- define page modules
- define field-level information hierarchy
- define state transitions
- define namespace and cluster inheritance rules

Primary supporting skill:
- `notion-spec-to-implementation`

Optional supporting skills:
- `playwright`
- `playwright-interactive`
  Use when validating real browser flow or checking whether the current UI matches the intended workflow

Output:
- IA doc
- state transition doc
- namespace decision doc

## Stage 4: V1 Boundary Definition

Goal:
- state clearly what V1 must include and what V1 must not include

Actions:
- list must-have workflow outcomes
- list deferred capabilities
- explicitly exclude AI and broad expansion when core workflow is incomplete

Primary supporting skill:
- `notion-spec-to-implementation`

Output:
- V1 implementation boundary doc

## Stage 5: Implementation Breakdown

Goal:
- translate V1 into execution phases and then into the first development batch

Actions:
- break V1 into dependent phases
- identify first batch only
- avoid parallelization until shared context rules are stable

Primary supporting skill:
- `notion-spec-to-implementation`

Output:
- implementation task breakdown
- initial development task list

## Stage 6: Implementation

Goal:
- implement only the active batch, starting from shared context and task continuity

Actions:
- inspect code structure first
- map tasks to exact files
- implement smallest viable slice
- verify affected tests

Primary supporting skill:
- no external planning skill is required here by default

Optional supporting skills:
- `openai-docs`
  Use only when the feature explicitly requires OpenAI product decisions or official API guidance
- `playwright` or `playwright-interactive`
  Use for UI flow verification after implementation

Important rule:
- do not bring AI into implementation scope unless AI is already in the accepted product boundary

## Stage 7: Review, PR, And Follow-Up

Goal:
- keep delivery aligned with the agreed scope

Actions:
- review against the V1 boundary checklist
- address review comments without reopening product scope
- fix CI only after understanding the failing check

Supporting skills:
- `gh-address-comments`
- `gh-fix-ci`

# Supporting Skill Map

Use this quick map:

- Requirement clarification:
  `notion-spec-to-implementation`
- Workflow and IA design:
  `notion-spec-to-implementation`
- Browser flow validation:
  `playwright` or `playwright-interactive`
- OpenAI product or API decisions:
  `openai-docs`
- Review comment handling:
  `gh-address-comments`
- CI debugging:
  `gh-fix-ci`

If your obra/superpowers bundle has equivalent skills, use them only when they fit the same stage. Do not switch stages just because a skill exists.

# Decision Rules

Before writing code, check these in order:

1. Is the user task clear?
2. Is the first workflow clear?
3. Is page context continuity clear?
4. Is V1 boundary clear?
5. Is this task inside the active batch?

If any answer is no, stop implementation and return to the missing design step.

# Non-Negotiable Guardrails

- Do not treat Homepage as a diagnostics dashboard.
- Do not let source-based navigation (`system`, `user`, `dynamic`) drive primary UX.
- Do not treat namespace as only a local page filter.
- Do not add AI to MVP just because it is attractive.
- Do not broaden V1 while the first workflow is still incomplete.
- Do not skip documentation for ambiguous product work.

# Expected Repository Outputs

When this skill is used correctly, work should usually produce or update some of these docs before implementation:
- `docs/product/product-requirements*.md`
- `docs/product/workflows/first-workflow*.md`
- `docs/product/workflows/field-level-ia*.md`
- `docs/product/workflows/low-fidelity-page-spec*.md`
- `docs/product/workflows/namespace-context-decision*.md`
- `docs/product/releases/v1/implementation-boundary*.md`
- `docs/product/releases/v1/implementation-task-breakdown*.md`
- `docs/product/releases/v1/initial-development-tasks*.md`

When these stabilize, they should be reorganized according to the process document structure above instead of continuing to accumulate flat under `docs/`.

# Exit Criteria

This SOP has been followed correctly when:
- product scope is documented before implementation,
- the first workflow is explicit,
- V1 boundary is written down,
- the active implementation batch is small and concrete,
- auxiliary skills are used to support the current stage instead of replacing discipline,
- and the process documents have a maintainable structure instead of indefinite flat growth.
