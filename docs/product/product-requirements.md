# KubeDeck Product Requirements Clarification

## 1. Purpose

This document clarifies the original product requirement behind KubeDeck so that future UI, architecture, and implementation decisions can be evaluated against the same target.

## 2. Problem Statement

Teams operating multiple Kubernetes clusters often rely on a fragmented toolchain:
- `kubectl` for low-level operations,
- separate dashboards for cluster visibility,
- and ad hoc internal tools for team-specific workflows.

This fragmentation increases context switching, weakens consistency, and makes custom workflows expensive to maintain.

## 3. Product Goal

KubeDeck should provide one extensible web control plane where users can:
- operate in a clear multi-cluster context,
- work with both built-in resources and CRDs through one model,
- and access team-specific workflows through plugins.

## 4. Target Users

Primary target users:
- platform engineers,
- SRE / DevOps operators,
- delivery engineers working across clusters.

Secondary users may exist later, but the product should not initially optimize for casual or read-only audiences.

## 5. Core User Jobs

The first core jobs to support are:
- choose the correct cluster,
- enter a high-frequency resource or workflow area,
- inspect relevant resources,
- perform a standard action,
- and continue follow-up work without losing context.

## 6. Unique Value

KubeDeck should differentiate by combining:
- multi-cluster context as a first-class concept,
- unified handling of built-in resources and CRDs,
- plugin-based workflow extension,
- and backend-authoritative access control.

Its value is not simply being another Kubernetes dashboard.

## 7. AI Product Positioning

AI in KubeDeck should be treated as a later task enhancement layer, not as part of the MVP and not as an isolated chat feature.

AI should help users complete real control-plane work by using the current cluster, namespace, resource, page, and workflow context.

However, AI should only be introduced after the first non-AI workflow is complete end-to-end. Before that point, AI would add complexity without enough product leverage.

The first three AI-supported task flows should be:
- resource creation and change execution,
- incident investigation and state understanding,
- and plugin-powered team workflows.

In these flows, AI should act as:
- an intent translator,
- a YAML and action assistant,
- a risk and result explainer,
- a state summarizer,
- and a workflow-specific helper embedded into plugin experiences.

AI should not be introduced as:
- a generic homepage chatbot,
- a context-free Kubernetes Q&A box,
- or an execution path that bypasses permission and workflow controls.

## 8. Minimum Viable Workflow

The first workflow that must work end-to-end is:

1. Enter the system.
2. Select or confirm cluster.
3. Enter a concrete resource domain.
4. View real resources.
5. Perform a standard operation such as create or apply.
6. Receive an actionable result.
7. Continue in the same working context.

AI can strengthen this workflow later by helping users draft changes, validate intent, and understand results, but AI should not replace the workflow itself and is not required for MVP.

## 9. Homepage Requirements

The homepage should help users:
- understand current cluster context,
- find the next likely task,
- and enter the main workflow quickly.

The homepage should not primarily optimize for framework diagnostics, developer verification, or metadata inspection.

If AI appears on the homepage, it should only help users enter or continue a task in context. It should not dominate the homepage as a standalone feature.

## 10. Why The Current UI Deviates

The current UI deviates from product goals because:
- it emphasizes runtime status and shell readiness over task entry,
- it exposes navigation source categories instead of user task categories,
- and it does not yet carry users into a complete workflow.

As a result, the interface communicates implementation progress more clearly than product value.

This also means AI should not be added on top of the current homepage structure as cosmetic innovation. Without a real task flow, AI would mostly amplify the wrong surface and increase scope before core value exists.

## 11. Decision Filter

Future changes should be judged by these questions:
- Does this make the first user workflow more complete?
- Does this reduce multi-cluster context switching?
- Does this strengthen the unified resource model?
- Does this enable meaningful plugin extension?
- Does this improve user task completion more than framework visibility?
