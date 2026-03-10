# KubeDeck V1 Implementation Boundary

## 1. Purpose

This document defines what V1 must include now that menu composition and resource-page extension have entered the product architecture.

V1 is complete only when one real workflow is usable on top of the new runtime shape.

## 2. V1 Core Principle

V1 must prove product value and architecture validity at the same time.

That means V1 is not complete merely because:

- the shell renders,
- plugin contracts exist,
- or menus and pages can technically register.

V1 is complete only when one real workflow is delivered through the current microkernel, menu, and resource-page models.

## 3. V1 Must Deliver

### 3.1 Runtime Foundation

V1 must preserve:

- the microkernel and plugin runtime,
- composed menu behavior,
- and the shared resource-page shell model.

### 3.2 First Real Workflow

V1 must deliver one complete workflow:

1. enter homepage,
2. confirm active cluster,
3. enter `Workloads` through the composed menu,
4. view real workload data,
5. enter a resource or action path,
6. complete one standard operation,
7. receive clear result feedback,
8. continue without losing core context.

### 3.3 Required Pages And Surfaces

V1 must include:

- homepage,
- left navigation driven by the composed menu model,
- `Workloads` as the first real domain,
- one shared resource-page shell,
- `Overview` and `YAML` as the default tabs,
- and one action path such as create, apply, or edit.

### 3.4 Required Context

V1 must preserve:

- active cluster,
- namespace scope,
- current menu and workflow domain,
- current resource identity when entering resource pages,
- and enough continuity to return to the prior working path.

### 3.5 Required Namespace Behavior

V1 must support:

- `single` namespace scope,
- `all` namespace scope for browsing,
- explicit target resolution for operations that require a concrete namespace,
- and continuity of namespace context between menu navigation, lists, and actions.

### 3.6 Required Menu Behavior

V1 must show the new menu model in product form:

- a system-owned default menu skeleton,
- explicit built-in entries,
- room for CRD and plugin mounts,
- disabled state for configured-but-unavailable entries,
- and a `CRDs` fallback resource entry.

User override storage may remain minimal in V1, but the runtime shape must not block it.

### 3.7 Required Resource Page Behavior

V1 must show the new resource-page model in product form:

- all first-wave resources resolve through the shared shell,
- `Overview` and `YAML` are real tabs,
- built-in specialization still fits the shared model,
- and the model leaves room for later tab replacement and takeover.

## 4. V1 Should Deliver If Low Cost

These items are useful but not required for V1 completion:

- lightweight global user menu preferences,
- cluster-local menu overrides with minimal persistence,
- more than one specialized resource page,
- more than one YAML variant,
- resume shortcuts on homepage,
- delayed `multiple` namespace selection in the UI.

## 5. V1 Must Not Include

V1 must not expand into:

- AI-assisted workflows,
- generalized chat surfaces,
- plugin marketplace or hot-install UI,
- broad observability dashboard work,
- broad block-level page-extension systems,
- large multi-domain expansion before the first domain is stable,
- or heavy visual polish before workflow validity is proven.

## 6. V1 Completion Check

V1 is in bounds only if all answers below are yes:

- Does the workflow run on top of the current microkernel runtime?
- Is the left menu composed rather than hard-coded?
- Can the user enter `Workloads` from that menu?
- Can the user reach a real resource list and resource page?
- Does the resource page expose `Overview` and `YAML` through the shared model?
- Can the user complete one real operation with valid target resolution?
- Can the user understand results and continue without losing context?

If any answer is no, V1 scope must not expand.
