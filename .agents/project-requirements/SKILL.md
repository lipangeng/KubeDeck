---
name: Project Requirements (Bilingual + Commenting)
description: Enforces bilingual (ZH+EN) deliverables and detailed comments for key logic. Always outputs docs in both Chinese and English, and ensures key code is well commented.
---

# Purpose
This skill defines project-wide requirements:
1) Bilingual support: English + Chinese.
2) Documentation outputs must be in two versions: EN and ZH.
3) Key code/logic must have detailed comments.

# When to use
Use this skill for any task that:
- Adds/edits user-facing docs (README, design docs, runbooks, CHANGELOG, API docs).
- Adds/edits code that includes non-trivial logic, algorithms, tricky edge cases, security, concurrency, networking, IO, caching, or business rules.
- Introduces new modules or core flows where future maintainability matters.

# Non-negotiable rules

## R1. Bilingual deliverables (EN + ZH)
- Any user-facing documentation must be produced in **two separate documents**:
    - English version (suffix `.en.md` or placed under `docs/en/`)
    - Chinese version (suffix `.zh.md` or placed under `docs/zh/`)
- Keep both versions **content-equivalent** (same headings/sections), unless locale-specific notes are explicitly required.
- If the repo already has a docs convention, follow it. Otherwise, default to:
    - `docs/en/<doc-name>.md`
    - `docs/zh/<doc-name>.md`

## R2. Source code comments for key logic
- For any **key logic**, add detailed comments explaining:
    - What the code is doing
    - Why this approach was chosen
    - Assumptions & invariants
    - Edge cases and failure modes
    - Complexity/performance notes (if relevant)
    - Security considerations (if relevant)
- Comments must be **bilingual** for key logic:
    - Prefer a two-line style:
        - `// EN: ...`
        - `// ZH: ...`
    - Or block comment with EN then ZH sections.
- Do **not** spam trivial comments. Focus on code that a new maintainer would struggle to reason about.

## R3. Output checklist
Before finishing any task, ensure:
- [ ] If docs changed: both EN + ZH versions updated/created.
- [ ] If key logic changed: bilingual comments exist for key areas.
- [ ] Documentation references (paths/commands/config) match actual repo state.

# Recommended conventions

## Documentation naming
- If updating existing docs:
    - Keep original path and add parallel EN/ZH docs if missing.
- If creating new docs:
    - Use `docs/en/` and `docs/zh/` default structure.

## Commenting style examples
### Example 1 (inline)
```ts
// EN: Use a bounded retry with jitter to reduce thundering-herd during partial outages.
// ZH: 使用带抖动的有界重试，避免部分故障时产生“惊群”请求风暴。
for (let attempt = 0; attempt < MAX_RETRY; attempt++) { ... }