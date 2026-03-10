# KubeDeck V1 Microkernel And I18n Implementation Boundary

## 1. Purpose

This document defines what microkernel and i18n work must be completed before broader V1 feature expansion continues.

## 2. Boundary Principle

The goal is not to finish the full plugin platform or full multilingual product before V1.

The goal is to establish the minimum architecture shape and minimum i18n discipline that prevent structural rework later.

## 3. V1 Must Deliver

V1 microkernel and i18n preconditions must deliver:
- explicit kernel contribution contract families
- a clear kernel-versus-capability ownership line
- a mapping from the first built-in workflow into capability-style ownership
- minimum i18n runtime rules
- and a ban on further uncontrolled inline copy growth

## 4. V1 Must Not Require Yet

V1 does not need to require:
- runtime third-party plugin loading
- plugin marketplace UX
- full slot replacement semantics
- full locale switching
- full translation coverage
- backend locale persistence

## 5. Expansion Gate

Broader UI and workflow expansion should pause if:
- built-in workflows are still being added as shell exceptions
- backend capability boundaries are still implicit
- or new user-facing text is still being added without a localization boundary

## 6. Completion Check

This boundary is satisfied when all answers below are yes:
- Are page/menu/action/slot contribution families defined?
- Is backend capability registration and execution authority defined?
- Can built-in workflow areas be described as capability contributions?
- Is locale treated as a product-level concern?
- Is there a minimum rule preventing uncontrolled new hard-coded copy?
