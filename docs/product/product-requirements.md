# KubeDeck Product Requirements

## 1. Purpose

This document is the current product requirement baseline for KubeDeck.

It defines what the product is for, who it serves, what the first version must prove, and which directions are intentionally deferred.

## 2. Product Problem

Platform teams operating multiple Kubernetes clusters usually work across fragmented tools:

- `kubectl` for low-level changes,
- separate dashboards for browsing state,
- internal tools for team-specific workflows,
- and ad hoc documentation for CRDs and custom operations.

This fragmentation causes:

- context switching,
- inconsistent operating paths,
- expensive workflow customization,
- and poor continuity across clusters and resource types.

## 3. Product Goal

KubeDeck should provide one extensible web control plane where users can:

- operate in a clear cluster context,
- work with built-in resources and CRDs through one model,
- navigate task-oriented menus instead of raw resource sprawl,
- and extend workflows through plugin-powered capabilities.

## 4. Target Users

Primary users are:

- platform engineers,
- SRE and DevOps operators,
- delivery engineers who work across clusters and environments.

The product is not initially optimized for casual viewers or dashboard-only audiences.

## 5. Core User Jobs

The first product jobs to support are:

1. identify the correct cluster and working scope,
2. enter a real resource or workflow domain quickly,
3. inspect current resources with stable context,
4. perform one standard action such as create, apply, or edit,
5. continue follow-up work without losing cluster or namespace continuity.

## 6. Product Value

KubeDeck differentiates through:

- multi-cluster context as a first-class operating concept,
- one mental model for built-in resources and CRDs,
- a microkernel and plugin runtime for capability extension,
- declarative menu composition instead of hard-coded navigation,
- and a resource-page shell that supports both default behavior and progressive specialization.

Its value is not “another Kubernetes dashboard”.

## 7. Navigation Requirement

Navigation must follow a composed menu model rather than a static list of registered entries.

The menu system must provide:

- a system-owned default menu blueprint,
- explicit mounting of built-in workflows, CRD-backed entries, and plugin routes,
- cluster-aware availability resolution,
- a standard `CRDs` fallback resource entry,
- and user customization through global and cluster-local overrides.

Only explicitly configured CRDs should appear in business-facing menu groups.

If a configured CRD is unavailable in the current cluster, the entry should remain visible but disabled with a clear hint.

## 8. Resource Page Requirement

Resource pages must follow one shared model.

The product should use:

- a unified `ResourcePageShell`,
- default tabs for `Overview` and `YAML`,
- tab-first extensibility for near-term growth,
- and explicit support for tab replacement or page takeover when necessary.

CRDs must always have a safe default path, including YAML access.

Built-in resources such as `Pod` may evolve richer resource-specific experiences, but they should still begin from the same top-level page model.

## 9. MVP Definition

V1 must prove one real workflow on top of the new architecture.

The minimal product loop is:

1. user enters the system,
2. confirms cluster context,
3. enters a real domain such as `Workloads`,
4. views real resources,
5. performs one standard operation,
6. receives clear result feedback,
7. stays in the same working context for the next step.

If this loop is not complete, the product is still an extensible shell rather than a usable control plane.

## 10. Homepage Requirement

The homepage must help the user:

- confirm current context,
- see the most likely next task,
- and enter that task immediately.

The homepage must not primarily behave like:

- a runtime diagnostics board,
- a framework verification screen,
- or a raw metadata view.

## 11. Current Non-Goals

The following are intentionally deferred:

- AI-assisted workflows,
- generic chat surfaces,
- full i18n product delivery,
- large observability dashboards,
- broad multi-domain expansion before the first workflow is stable,
- fine-grained block-level resource-page extension.

## 12. Decision Filter

Future product and implementation decisions should be judged by these questions:

- Does this strengthen the first real workflow?
- Does this preserve the microkernel and plugin runtime shape?
- Does this fit the composed menu model?
- Does this fit the shared resource-page shell model?
- Does this reduce context switching across clusters and resource types?
- Does this increase product value more than it increases framework complexity?
