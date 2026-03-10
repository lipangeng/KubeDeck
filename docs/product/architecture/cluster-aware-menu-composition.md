# Cluster-Aware Menu Composition Design

## Status

Draft approved through brainstorming discussion.

This document defines the target menu model for KubeDeck after the microkernel baseline.

## 1. Problem Statement

The current kernel contribution pipeline already supports `page`, `menu`, `action`, and `slot` contributions, but the left navigation still behaves like a rendered list of registered menu entries.

That is not sufficient for the intended product behavior.

KubeDeck needs a menu system where:

- the system provides a strong default menu,
- CRDs and plugin routes are treated as equal mountable capabilities,
- cluster availability changes the enabled or disabled state of configured entries,
- and users can customize the final menu with both global and cluster-local overrides.

The menu system must therefore move from "registered menu items" to "composed menu results".

## 2. Design Goals

The menu system must satisfy these goals:

1. Preserve a stable default information architecture.
2. Treat built-in workflows, CRDs, and plugin routes as uniform mountable inputs.
3. Keep the menu focused on high-value workflows instead of turning it into a raw Kubernetes resource tree.
4. Always provide a standard `CRDs` resource entry as a fallback resource path.
5. Show configured-but-unavailable capabilities as disabled instead of silently removing them.
6. Support user customization through global and cluster-local overrides.

## 3. Non-Goals

This design does not attempt to:

- auto-generate the full business menu from every discovered CRD,
- let users redefine capability truth or cluster availability,
- turn menu configuration into a scripting system,
- or replace backend capability contracts with UI-only rules.

## 4. Core Model

The menu system is composed from four layers.

### 4.1 `MenuBlueprint`

The system-owned default menu skeleton.

This layer defines:

- default groups,
- group order,
- default mounted entries,
- and baseline information architecture.

It does not directly encode cluster truth.

### 4.2 `MenuMount`

A mountable capability that can be placed into the blueprint.

Every mount must have a uniform shape regardless of source.

Valid sources:

- built-in workflow
- CRD-backed workflow entry
- plugin route

Each mount should at minimum provide:

- `id`
- `sourceType`
- `target`
- `defaultGroup`
- `order`
- `title`
- `icon`
- `visibility`
- `availability constraints`

### 4.3 `MenuCompositionResult`

The final menu produced for one user and one cluster context.

This is the only layer consumed by the frontend navigation renderer.

### 4.4 `MenuOverride`

User customization on top of the composed result.

Two override scopes are required:

- global user override
- cluster-local override

## 5. Default Menu Skeleton

The default first-level menu should use four groups.

### 5.1 `Core`

High-frequency, essential operating entry points.

Typical examples:

- `Workloads`
- `Services`
- `Config`
- `Secrets`

### 5.2 `Platform`

Platform-oriented or medium-frequency domains.

Typical examples:

- `Networking`
- `Storage`
- `Observability`
- built-in platform workflows

### 5.3 `Extensions`

Explicitly mounted CRD-backed entries and plugin routes.

This is the main location where CRD capabilities and plugin capabilities are treated equally.

### 5.4 `Resources`

Fallback resource entry points.

This group is not the primary task menu. It exists to preserve full resource access paths.

At minimum it must include:

- `CRDs`

## 6. CRD Handling Rules

CRDs are Kubernetes resources and must have one standard fallback entry:

- `Resources -> CRDs`

That entry opens:

1. the CRD definition list
2. then the selected CRD's instance list

Important rule:

- the menu system must not attempt to auto-place every discovered CRD into the primary navigation

Instead:

- only CRDs explicitly mounted into the blueprint or user override appear in other groups

If a configured CRD does not exist in the current cluster:

- it remains visible in the menu
- it is disabled
- it shows an availability indicator and an explanatory tooltip or hint

## 7. Plugin Route Handling Rules

Plugin routes follow the same mounting model as CRD entries.

A plugin route:

- does not own its final menu position by itself
- only exposes a mountable capability
- is placed by the blueprint and override system

If a plugin capability is unavailable in the current context:

- it remains visible if it was explicitly configured
- it becomes disabled
- it must not silently disappear unless policy explicitly hides it

## 8. Availability States

Each composed menu entry must resolve to one of these states:

- `enabled`
- `disabled-unavailable`
- `hidden`

### `enabled`

The capability exists and is accessible in the current context.

### `disabled-unavailable`

The entry is intentionally configured, but the backing capability is missing in the current cluster or context.

This is required to preserve menu stability.

### `hidden`

The entry is suppressed by policy or user override.

## 9. User Override Model

Two override levels are required.

### 9.1 Global Override

Applies to all clusters for one user.

Allowed operations:

- reorder groups
- reorder items
- move an item between groups
- hide an item
- pin an item
- optionally assign a display alias

### 9.2 Cluster Override

Applies only to one cluster for one user.

Used for:

- cluster-specific reorder
- cluster-specific hide
- cluster-specific move
- cluster-specific pin

The final composition order must be:

1. system blueprint
2. dynamic mount resolution
3. global override
4. cluster override

## 10. Guardrails

The override system must not allow users to:

- invent capabilities that do not exist
- turn unavailable capabilities into enabled ones
- redefine backend capability truth
- or bypass authorization and policy checks

The override system is for menu composition only, not capability definition.

## 11. Implications For Current Architecture

The current system already provides useful groundwork:

- backend manifest discovery
- frontend module discovery
- contribution contracts for `page/menu/action/slot`
- end-to-end microkernel composition

However, the current `menu contribution` abstraction is still too close to direct page registration.

To support this design, menu handling must evolve from:

- "registered menu items"

to:

- "menu blueprint plus mount composition"

This is a design-level evolution, not just a visual refactor.

## 12. Current Plan Alignment

The project is still aligned with the approved microkernel roadmap.

Recent work completed the extensibility pipeline:

- backend manifest loading
- frontend plugin module discovery
- sample plugin validation
- slot and remote page rendering

What is still missing is the higher-level menu composition system described in this document.

That means the project is not off-plan, but the next design layer has now become explicit.

## 13. Recommended Next Planning Step

Before implementation, create a dedicated plan for:

- `MenuBlueprint` schema
- `MenuMount` schema
- composition pipeline
- override persistence model
- and availability-resolution rules

No menu implementation work should start before that plan is reviewed.
