# KubeDeck V1 Microkernel And I18n First Patch Scope

## 1. Purpose

This document defines the minimum first patch scope for introducing kernel-oriented structure and i18n minimum discipline.

## 2. Patch Principle

The first patch should establish structure and guardrails, not full feature breadth.

## 3. First Patch Should Include

The first patch should include only:
- kernel contribution type definitions
- minimal built-in contribution registration structure
- minimal shell-to-capability composition entry
- minimum i18n copy-access boundary for new UI text
- tests or validation covering contract stability where practical

## 4. First Patch Must Not Include

The first patch must not include:
- full dynamic plugin loading
- plugin marketplace or management UI
- broad backend rewrite
- full runtime locale switcher
- full translation migration of all existing pages
- unrelated visual redesign

## 5. Existing-Code Rule

The first patch may coexist with historical code, but it must stop deepening shell-owned workflow behavior.

## 6. Success Criteria

The first patch is successful when:
- the kernel contract baseline exists
- built-in workflow registration starts moving behind contribution boundaries
- new UI text gains a consistent future-localizable access path
