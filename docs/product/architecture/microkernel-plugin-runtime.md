# Microkernel And Plugin Runtime

## 1. Purpose

This document defines the current frontend and backend runtime shape that all new development must follow.

It consolidates the earlier microkernel, contract, mapping, and minimum-i18n drafts into one current architecture document.

## 2. Architecture Position

The microkernel is not a later enhancement.

It is the runtime boundary that protects KubeDeck from turning back into a hard-coded shell with special-case built-in pages.

The product depends on:

- contribution registration,
- runtime composition,
- backend capability authority,
- and extensible UI surfaces.

## 3. Runtime Layers

KubeDeck uses four runtime layers.

### 3.1 Kernel Core

The kernel core owns:

- contribution contracts,
- composition rules,
- route and menu assembly,
- context boundaries,
- and execution boundaries.

The kernel must exist on both frontend and backend.

### 3.2 Built-In Capabilities

Built-in product areas such as `Homepage`, `Workloads`, resource pages, and core actions are not permanent shell exceptions.

They are in-repository capabilities that must attach through kernel contracts.

### 3.3 Plugin Capabilities

Plugins contribute capabilities through the same runtime family used by built-ins.

Current contribution families are:

- `page`
- `menu`
- `action`
- `slot`

Resource-page expansion now also depends on extension capability types such as:

- `tab`
- `tab-replace`
- `page-takeover`

Additional types such as `section` and `slot` remain modeled for later use.

### 3.4 Runtime Resolution

The final UI should render resolved results, not raw registrations.

This applies to:

- menus,
- pages,
- actions,
- resource-page tabs,
- and extension surfaces.

## 4. Backend Responsibilities

The backend is authoritative for:

- capability registration,
- plugin identity and manifest loading,
- capability composition,
- resource and workflow execution entry points,
- and permission-aware availability.

The backend must not degrade into a thin metadata stub if the product depends on capability truth.

## 5. Frontend Responsibilities

The frontend is responsible for:

- consuming resolved capability metadata,
- composing the UI runtime from kernel inputs,
- rendering menus and pages from composed results,
- enforcing product information architecture,
- and preserving working context across cluster, menu, and resource navigation.

The frontend must not bypass the kernel by building unrelated page systems.

## 6. Built-In Rule

Built-in functionality must be modeled as first-class kernel contributions.

That means:

- built-in pages should be registered through kernel contracts,
- built-in menu entries should participate in menu composition,
- built-in actions should attach through action contracts,
- built-in resource experiences should follow the same page-extension model available to later contributors.

## 7. Menu Runtime Rule

Navigation is not a direct render of page registration.

It must resolve through:

1. default blueprint,
2. mountable capabilities,
3. cluster-aware availability,
4. user overrides,
5. final composed menu result.

This is now a required part of the runtime shape.

## 8. Resource Page Runtime Rule

Resource pages are not arbitrary standalone screens.

They must resolve through:

1. shared `ResourcePageShell`,
2. default tabs,
3. extension capabilities,
4. optional tab replacement,
5. optional page takeover for exceptional resource types.

This rule applies to both built-in resources and CRDs.

## 9. I18n Minimum Runtime Rule

Full i18n delivery is still deferred, but the runtime must preserve a locale boundary.

Current minimum rules are:

- new user-facing copy must flow through an i18n access layer,
- capability labels must remain localizable,
- locale must be treated as product-level state,
- and text must not be tightly mixed with control logic.

## 10. What Is Deferred

The following are valid later directions, not current runtime requirements:

- dynamic hot-loading and unloading of third-party plugins,
- marketplace or plugin-management UI,
- full slot replacement across all page regions,
- full runtime locale switching and persistence,
- and broad block-level page extension.

## 11. Architecture Check

Before adding major features, these answers must remain yes:

- Can the feature attach through current kernel contracts?
- Does it preserve composed menus rather than hard-coded navigation?
- Does it preserve the shared resource-page shell?
- Does it keep built-in functionality inside the same runtime family as plugins?
- Does it avoid deepening non-localizable UI text?

If the answer to any question is no, architecture adjustment must happen before feature expansion.
