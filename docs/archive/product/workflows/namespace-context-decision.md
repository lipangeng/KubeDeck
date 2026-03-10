# KubeDeck Namespace Context Decision Draft

## 1. Purpose

This document defines how namespace should behave in KubeDeck as a cross-page working context.

Namespace is not only a page-level filter. It affects:
- visible resource scope,
- default operation target,
- page-to-page continuity,
- and workflow safety.

## 2. Decision Summary

Namespace should be treated as a shared working context under the currently active cluster.

KubeDeck should support these namespace scopes:
- single namespace,
- multiple namespaces,
- all namespaces.

However, action-oriented flows such as `Create` and `Apply` must always resolve to an explicit execution target before submission.

## 3. Namespace Model

### 3.1 Context Hierarchy

The recommended context hierarchy is:

1. active cluster
2. namespace scope
3. current resource domain or page

This means namespace belongs to global working context, not only to one page.

### 3.2 Supported Namespace Scope Types

KubeDeck should support:
- `single`
- `multiple`
- `all`

Definitions:
- `single`: exactly one namespace is active
- `multiple`: a defined set of namespaces is active
- `all`: all visible namespaces in the current cluster are in scope

## 4. Page-Type Rules

### 4.1 List And Browse Pages

List-oriented pages such as `Workloads` should support:
- single namespace,
- multiple namespaces,
- all namespaces.

Reason:
- users often browse across wide scope before deciding what to act on,
- and this matches real Kubernetes operational behavior.

### 4.2 Detail Pages

Detail pages should inherit the namespace context from the source page.

If the viewed object is namespaced:
- the object namespace becomes explicit and must be shown.

If the viewed object is cluster-scoped:
- namespace context can remain as inherited browsing context,
- but the object itself should be marked as cluster-scoped.

### 4.3 Create / Apply Pages Or Surfaces

Create / Apply flows must not submit against ambiguous namespace scope.

Before final submission, execution must resolve into one of these:
- a single explicit namespace,
- or an object-defined namespace from the payload,
- or a cluster-scoped resource with no namespace target.

Multiple and all-namespace scope may help determine user context before create/apply starts, but they are not valid final execution targets by themselves.

## 5. Default Rules

### 5.1 Initial Default

When the user first enters a cluster:
- if a stored last-used namespace scope exists for that cluster, use it,
- otherwise default to `single: default`.

### 5.2 Cluster Switch Default

When the user switches cluster:
- first try to restore the last-used namespace scope for the target cluster,
- if not available, fall back to `single: default`.

KubeDeck should not blindly carry namespace names across clusters if the target cluster may not have that namespace.

### 5.3 No Context Case

If there is no prior namespace information:
- use `single: default` as the safe baseline for action-oriented entry,
- while still allowing the user to broaden scope in list-oriented pages.

## 6. Inheritance Rules

### 6.1 Homepage To Workloads

Homepage should show only the minimum namespace context summary.

When the user enters `Workloads`:
- inherit the current namespace scope from shared working context.

### 6.2 Workloads To Detail

When the user opens a resource from `Workloads`:
- inherit the current namespace scope,
- but also show the actual object namespace explicitly.

### 6.3 Workloads To Create / Apply

When the user starts `Create` or `Apply` from `Workloads`:
- pass the current namespace scope into the operation surface as context.

Then resolve as follows:
- if scope is `single`, use that namespace as the default target,
- if scope is `multiple` or `all`, require explicit namespace confirmation or rely on object-defined namespace before submit.

### 6.4 Return Flow

After create/apply completes:
- return to the originating page with the original namespace browsing scope preserved.

This means the execution target namespace and the browsing scope may be different values after the action.

## 7. Resolution Rules For Create / Apply

### 7.1 Single Namespace Scope

If the current scope is `single`:
- use that namespace as the default execution target,
- unless the payload explicitly defines another valid namespace.

### 7.2 Multiple Namespace Scope

If the current scope is `multiple`:
- do not auto-pick one namespace silently,
- require the user to confirm one target namespace,
- or require the payload to define it per object.

### 7.3 All Namespace Scope

If the current scope is `all`:
- do not treat `all` as a valid write target,
- require explicit target namespace resolution before submit,
- unless the object is cluster-scoped.

### 7.4 Multi-Document Apply

For multi-document apply:
- each document must resolve to a concrete target,
- and the result should indicate which namespace each document used.

If some documents are cluster-scoped and some are namespaced:
- this must be visible in both validation and result feedback.

## 8. Validation Rules

Before submit, KubeDeck should validate:
- whether the final namespace target is concrete when required,
- whether the target namespace is valid in the current cluster,
- whether any document conflicts with the chosen namespace resolution rule,
- and whether the resource is cluster-scoped or namespaced.

Validation must not silently rewrite ambiguous namespace scope into an unintended target.

## 9. UI Rules

### 9.1 Browsing UI

List pages should allow namespace scope controls for:
- single,
- multiple,
- and all.

The UI must clearly distinguish:
- current browsing scope,
- and actual object namespace where relevant.

### 9.2 Action UI

Create / Apply UI must show:
- current cluster,
- current browsing scope,
- resolved execution namespace target,
- and any unresolved namespace problem before submit.

### 9.3 Feedback UI

Result feedback should show:
- which namespace scope the user was browsing,
- which concrete namespace each action used,
- and which objects were cluster-scoped.

## 10. Safety Rules

- Do not treat `all namespaces` as a write target.
- Do not silently collapse `multiple` into one namespace.
- Do not drop namespace context when navigating between pages.
- Do not preserve a namespace across cluster switch without verifying or safely falling back.

## 11. Recommended First Implementation

For the first implementation:
- browsing supports `single` and `all`,
- `multiple` may be modeled in the decision now but delayed in UI delivery if needed,
- create/apply must resolve to a single namespace or a cluster-scoped object before submit,
- and namespace context must persist between Homepage, Workloads, and Create / Apply.

This keeps the model correct without forcing full multi-select UI complexity into the first delivery.
