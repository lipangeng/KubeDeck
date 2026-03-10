# KubeDeck I18n Runtime Model Draft

## 1. Purpose

This document defines the minimum runtime model that allows KubeDeck to grow without blocking later localization work.

It does not require full i18n delivery in V1.

## 2. Runtime Principle

V1 may ship with one dominant runtime locale, but the runtime model must already separate:
- locale preference,
- copy access,
- and localizable contribution metadata

from page logic.

## 3. Minimum Runtime Elements

### 3.1 Locale Preference

Locale should be treated as product-level preference state.

It should not be treated as page-local UI state.

### 3.2 Copy Access Layer

User-facing text should be accessed through one consistent copy-access pattern instead of arbitrary inline strings spread across pages and feature components.

V1 may keep this simple, but the access path must exist.

### 3.3 Localizable Contribution Metadata

Kernel contributions should already tolerate localizable labels and descriptions.

V1 may still resolve them from one locale source, but the contract shape must allow later expansion.

### 3.4 Formatting Boundary

Date, time, number, and other locale-sensitive presentation concerns should be treated as formatting boundaries rather than plain string concatenation.

Full formatting coverage may be deferred, but the boundary should be recognized.

## 4. What V1 Must Do

V1 should:
- treat locale as a product preference concept
- avoid expanding inline copy sprawl
- keep contribution metadata future-localizable
- keep text and control logic separate in new code

## 5. What V1 Does Not Need To Do

V1 does not need to ship:
- complete runtime language switching
- complete translation coverage
- complete locale-aware formatting everywhere
- backend locale persistence

## 6. New-Code Rule

From now on, new user-facing text in product UI should be added in a way that can later be moved behind a locale lookup without page rewrites.

This applies to:
- page titles
- button labels
- action labels
- result messages
- empty states
- workflow hints

## 7. Runtime Readiness Check

Before broader UI growth continues, these questions should be answerable with "yes":
- Is locale already treated as a product preference concept?
- Can new UI copy be moved behind locale lookup without redesigning page logic?
- Can contribution labels be localized later without changing contract shape?
