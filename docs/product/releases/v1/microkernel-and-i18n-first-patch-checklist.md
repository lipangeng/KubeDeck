# KubeDeck V1 Microkernel And I18n First Patch Checklist

## 1. Purpose

This checklist prevents the first kernel/i18n patch from expanding into a full platform rewrite.

## 2. Before Start

- [ ] The microkernel and i18n decision document is accepted.
- [ ] The kernel contribution contract draft is accepted.
- [ ] The kernel implementation mapping draft is accepted.
- [ ] The i18n runtime model draft is accepted.
- [ ] The V1 boundary document is accepted.

## 3. In Progress Guardrails

- [ ] Do not add full runtime plugin loading.
- [ ] Do not add marketplace or plugin management UX.
- [ ] Do not widen backend work beyond minimum contract support.
- [ ] Do not attempt full translation migration.
- [ ] Do not keep adding new shell-only workflow logic.

## 4. Required Deliverables

- [ ] kernel contract baseline
- [ ] initial contribution-oriented registration path for built-in workflow areas
- [ ] minimum copy-access boundary for new UI text

## 5. Drift Indicators

The patch is drifting if:
- [ ] shell logic keeps owning more built-in workflow behavior
- [ ] contribution contract design is still implicit in page components
- [ ] new user-facing text is still added as arbitrary inline strings
- [ ] full plugin platform scope enters the patch
- [ ] full i18n rollout enters the patch

## 6. Done Criteria

- [ ] The first patch establishes structure, not full expansion.
- [ ] Built-in workflows have a visible path toward capability-style ownership.
- [ ] New UI text no longer deepens future i18n rewrite cost.
