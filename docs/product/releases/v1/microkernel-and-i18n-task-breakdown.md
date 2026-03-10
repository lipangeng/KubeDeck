# KubeDeck V1 Microkernel And I18n Task Breakdown

## 1. Purpose

This document breaks the microkernel and i18n preconditions into an execution-ready sequence.

## 2. Task Order

### Task A: Kernel Contract Definition

Define:
- frontend page/menu/action/slot contract shapes
- backend capability/metadata/execution contract shapes

Output:
- contract draft accepted

### Task B: Kernel Ownership Mapping

Define:
- shell ownership
- built-in capability ownership
- shared context ownership
- backend kernel versus execution ownership

Output:
- implementation mapping accepted

### Task C: I18n Runtime Minimum Model

Define:
- locale preference ownership
- copy access boundary
- localizable contribution metadata rule
- no-inline-sprawl rule for new code

Output:
- i18n runtime model accepted

### Task D: V1 Expansion Gate

Define:
- what feature work must pause until kernel and i18n preconditions are respected
- what can continue in parallel

Output:
- V1 boundary accepted

### Task E: First Patch Planning

Plan:
- first kernel-oriented patch scope
- first i18n-discipline patch scope
- checks that prevent accidental overreach

Output:
- first patch scope and checklist accepted

## 3. Recommended Execution Rule

Do not start deeper feature expansion until Tasks A through D are accepted.

Task E should be the last pre-development step before code work.
