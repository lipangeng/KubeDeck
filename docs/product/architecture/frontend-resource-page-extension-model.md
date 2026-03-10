# Frontend Resource Page Extension Model

## Status

Current architecture document.

This document defines the target frontend resource-page extension model for KubeDeck.

It is intended to sit on top of the existing microkernel baseline and the cluster-aware menu composition model.

## 1. Problem Statement

KubeDeck needs a resource-page model that satisfies several competing goals at the same time:

- CRD resources must always have a safe default page and YAML editing path.
- Built-in Kubernetes resources such as `Pod` may need richer and more specialized UI.
- The platform must support future page extensions without forcing every resource into a fully custom page.
- The system must allow partial extension, tab replacement, and eventual full-page takeover.

This means the frontend resource-page model cannot be reduced to:

- one generic page for everything, or
- a disconnected set of special pages for built-in resources.

Instead, KubeDeck needs a unified shell with layered extension capability.

## 2. Design Goals

The model must satisfy these goals:

1. Every resource can resolve to a default usable page.
2. YAML is always available as a baseline capability.
3. Built-in resources and CRDs share one top-level resource-page model.
4. The near-term system should prioritize tab-based extensibility.
5. Full-page takeover remains possible for a small number of resource types.
6. The extension model must be abstract enough to support future extension-point types.

## 3. Non-Goals

This design does not attempt to:

- implement block-level extension in the current phase,
- define final visual design for each resource page,
- or decide every tab for every Kubernetes resource type.

The goal is to define the model and priority order, not the complete final implementation.

## 4. Recommended Model

KubeDeck should use:

- a unified `ResourcePageShell`,
- a default tab model,
- layered extension capabilities,
- and selective page takeover for exceptional resource types.

This is the approved direction:

`ResourcePageShell + Default Tabs + Extension Capabilities + Final Resolution`

## 5. Core Objects

### 5.1 `ResourcePageShell`

The shell is the shared outer structure for all resource pages.

It should own:

- resource identity context
  - cluster
  - namespace
  - resource kind
  - resource name
- shared page layout
- shared action area
- shared tab container
- default loading / empty / error handling

The shell must not encode resource-specific business logic.

### 5.2 `ResourceCapability`

This describes what a resource type is allowed to expose through the page model.

Examples:

- supports `overview`
- supports `yaml`
- supports extra runtime tabs
- supports tab replacement
- supports page takeover

This layer describes capability, not concrete UI composition.

### 5.3 `ResourceExtensionCapability`

This is the abstraction layer for extension types.

Even if only a subset is implemented in the near term, the model must be explicit from the start.

Recommended extension types:

- `tab`
- `tab-replace`
- `page-takeover`
- `action`
- `slot`
- `section`

Near-term implementation priority should cover only:

- `tab`
- `tab-replace`
- `page-takeover`

The other types are deliberately modeled early but implemented later.

### 5.4 `ResourceExtension`

This is one concrete contribution for a resource page.

It may come from:

- built-in resource logic
- plugin contribution

It may add or change:

- tabs
- actions
- takeover behavior
- future slot or section extensions

### 5.5 `ResourcePageResolution`

The final resolved result for one resource in one context.

It answers:

- which shell is used,
- which tabs are present,
- which tabs are replaced,
- which actions are enabled,
- and whether the page is still shell-based or fully taken over.

## 6. Default Tabs

The approved baseline model is tab-first.

The shell should not merely contain tabs as a later extension point. Instead, default capabilities should already be expressed as tabs.

At minimum, every resource page should resolve to these default tabs:

- `Overview`
- `YAML`

This rule applies to both:

- built-in resources
- CRD resources

## 7. YAML Baseline Rule

YAML is not an optional extra feature.

It is a mandatory baseline capability of the resource-page model.

That means:

- every resource type must have a YAML path by default,
- CRDs must always be operable through YAML,
- and future YAML variants must be modeled as evolvable tab capabilities rather than ad-hoc page exceptions.

Examples of future YAML-capability variants:

- `yaml.v1`
- `yaml.v2`
- `yaml.diff`
- `yaml.assisted`

Different resource types may enable different YAML variants over time, but the existence of a YAML path must remain guaranteed.

## 8. Tab-First Extensibility Strategy

Near-term extension work should focus on tabs.

Why:

- tabs provide a stable and visible extension surface,
- they reduce the need for immediate block-level extension,
- and they can represent most early resource-page enhancements cleanly.

Examples of future tabs:

- `Events`
- `Logs`
- `Metrics`
- `Runtime`
- `Related Resources`
- custom plugin tabs

This means most short-term resource-page specialization should happen by:

- adding tabs,
- replacing tabs,
- or selectively disabling tabs.

## 9. Built-In Resources Versus CRDs

Built-in resources and CRDs should not use two unrelated page models.

They should share the same shell and extension model.

The difference is not in the existence of the model, but in how much they extend it.

### CRDs

CRDs should default to:

- `Overview`
- `YAML`

Additional tabs may be added only when explicitly contributed.

### Built-In Resources

Built-in resources may add richer tabs and resource-specific capabilities.

Examples:

- `Pod` may later add runtime-oriented tabs
- other built-in resources may add status or topology tabs

But built-in resources should still begin from the same shell-based model unless takeover is explicitly justified.

## 10. Page Takeover Rule

Some resource types may eventually need a full custom page.

This is allowed, but must not be the default path.

Recommended rule:

- default to shell plus tabs,
- escalate to full-page takeover only when the shell model is clearly insufficient.

This means the approved operating mode is:

- support shell-based enhancement by default,
- permit takeover as an explicit capability,
- do not normalize takeover as the first solution.

For complex resources such as `Pod`, the approved direction is:

- model supports both shell-based extension and takeover,
- but default implementation priority should start with shell-based extension first.

## 11. Block-Level Extension Priority

Block-level or section-level extension is recognized as useful but not required in the near term.

Examples of deferred extension types:

- summary block injection
- inline section replacement
- side-panel fragments
- page-body subsection overrides

These should remain modeled as future capability types, but they are not required for the current design phase.

The current priority order is:

1. shell
2. default tabs
3. tab extension
4. tab replacement
5. page takeover
6. block-level extension later

This keeps the near-term system simpler while preserving long-term extensibility.

## 12. Implications For Current Frontend Architecture

The current microkernel baseline already supports:

- page contributions
- menu contributions
- action contributions
- slot contributions

However, resource-page extensibility requires a more specific model inside the page layer.

The system must evolve from:

- generic page contributions only

to:

- resource-page shell resolution
- tab capability resolution
- and takeover-capable resource-page composition

This is an architectural evolution of the frontend extension model, not merely a UI enhancement.

## 13. Current Plan Alignment

The project remains aligned with the broader plan.

The microkernel and plugin pipeline work completed so far was necessary groundwork.

This document identifies the next missing layer:

- a structured resource-page extension model

That means the project is still on the approved path, but the page-extension layer must now be explicitly designed before implementation.

## 14. Recommended Next Planning Step

Before implementation, create a dedicated plan for:

- `ResourcePageShell` contract
- tab capability schema
- tab registration and replacement rules
- YAML capability versioning strategy
- page-takeover decision rules

No resource-page implementation expansion should begin before that plan is reviewed.
