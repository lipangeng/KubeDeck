# KubeDeck Microkernel Extensibility And I18n Minimum Constraints

## 1. Purpose

This document defines which parts of the microkernel architecture and i18n model must enter the product now, and which parts can be deferred after the first workflow becomes stable.

The goal is to prevent two failure modes:
- shipping the first workflow by hard-coding built-in behavior into the shell,
- or postponing all i18n thinking until a later rewrite.

## 2. Decision Summary

The two topics do not have the same urgency.

Microkernel extensibility is a product-defining architecture concern and must enter the implementation path now.

I18n is also important, but full i18n capability can be deferred. What must enter now is the minimum rule set that prevents future page growth from locking text and behavior together.

## 3. Why Microkernel Must Enter Now

KubeDeck is defined as a multi-cluster, plugin-extensible control plane rather than a fixed built-in dashboard.

That means the following are not secondary implementation details:
- how pages are registered,
- how menus are contributed,
- how actions are attached to workflows,
- how built-in resource areas are represented,
- and how backend capabilities are exposed to the shell.

If these concerns are postponed, the first workflow will likely be implemented as application-specific special cases. Once that happens, later "pluginization" becomes a structural rewrite instead of controlled evolution.

## 4. Why Full I18n Can Be Deferred

The first real workflow is still more important than complete language delivery.

KubeDeck does not need full i18n in V1 if doing so would delay the first usable control-plane workflow.

However, the project should not continue creating new UI layers with hard-coded page text and no locale boundary. That would turn later i18n into a large mechanical rewrite across the shell, pages, and plugin contracts.

## 5. Microkernel Constraints That Must Enter Now

The following must enter the architecture immediately.

### 5.1 Built-In Functionality Must Follow Kernel Rules

Built-in product areas such as `Homepage`, `Workloads`, and their actions must not be treated as permanent shell-only exceptions.

They may ship inside the repository, but they should be modeled as first-class kernel contributions rather than privileged ad hoc code paths.

### 5.2 Frontend Must Have Explicit Contribution Contracts

The frontend shell must define stable contribution contracts for at least:
- page registration,
- menu contribution,
- action contribution,
- and extension slot contribution.

The shell may still provide default built-in implementations, but built-in code should attach through the same contract family wherever possible.

### 5.3 Backend Must Have Explicit Capability Boundaries

The backend kernel must define stable capability boundaries for at least:
- plugin identity,
- resource or workflow capability registration,
- menu or page metadata exposure,
- and policy-authoritative execution entry points.

The backend should not only act as a stub router with a thin plugin registry.

### 5.4 First Workflow Must Be Compatible With Extension

The first workflow does not need full third-party plugin delivery yet, but its structure must not prevent:
- adding new workflow domains,
- contributing actions into an existing workflow,
- attaching side panels or contextual extensions,
- and expanding built-in resource handling without rewriting the shell.

### 5.5 Kernel Ownership Must Be Clear

The kernel owns:
- contribution registration,
- composition rules,
- route and menu assembly,
- workflow context boundaries,
- and permission-aware execution boundaries.

Plugins or built-in capability modules own:
- their pages,
- their actions,
- their domain-specific view logic,
- and their contribution metadata.

## 6. Microkernel Items That Can Be Deferred

The following can be deferred until after the first workflow is stable:
- dynamic third-party plugin loading at runtime,
- hot installation or removal of plugins,
- a rich marketplace or plugin management UI,
- full slot replacement semantics across all page regions,
- and broad plugin coverage across many resource domains.

These are valid later directions, but they are not required to establish the kernel shape now.

## 7. I18n Minimum Constraints That Must Enter Now

The following i18n rules should start now, even if complete i18n is deferred.

### 7.1 New UI Text Must Have A Locale Boundary

New product UI text should not keep spreading as arbitrary hard-coded strings inside page and feature components.

At minimum, the project should establish one consistent place or access pattern for UI copy, even if only one locale is initially populated.

### 7.2 Contracts Must Support Localizable Labels

Menu, page, action, and extension contribution contracts should be defined so that labels and descriptions can later be localized without changing the contract shape.

The project should avoid assuming that a contribution only has one fixed display string forever.

### 7.3 Locale Must Be Treated As A Product Preference

Even if language switching is not built in immediately, locale must already be considered a product-level preference rather than page-local state.

### 7.4 Text And Logic Must Stay Separate

Implementation should avoid mixing control logic with user-facing copy in a way that makes later extraction difficult.

This applies especially to:
- page titles,
- button labels,
- result messages,
- empty states,
- and workflow guidance text.

## 8. I18n Items That Can Be Deferred

The following can be postponed after the first workflow is stable:
- full locale switcher UI,
- complete EN/ZH runtime language selection,
- date, time, and number locale formatting everywhere,
- backend-driven locale preference persistence,
- and translated plugin packs from third parties.

These are important, but they should not block the first usable workflow.

## 9. Immediate Execution Rule

From this point onward:
- new implementation work must not deepen shell-only built-in special cases,
- new workflow surfaces should be evaluated against kernel contribution boundaries,
- and new UI text should not be added without a future-localizable access path.

In short:
- microkernel shape is a now decision,
- full i18n delivery is a later decision,
- minimum i18n discipline is a now decision.

## 10. Architecture Readiness Check

Before continuing major implementation, these questions should be answerable with "yes":
- Can built-in workflow pages be described as kernel contributions instead of shell exceptions?
- Are menu, page, action, and extension boundaries explicit?
- Does the first workflow still leave room for later plugin contribution?
- Can new user-facing text be localized later without rewriting component logic?

If any answer is "no", architecture work should resume before broader feature expansion.
