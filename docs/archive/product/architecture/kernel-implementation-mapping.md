# KubeDeck Kernel Implementation Mapping Draft

## 1. Purpose

This document maps the kernel contribution contract into implementation-oriented ownership areas without depending on the current historical code layout.

## 2. Mapping Principle

The kernel should be small, stable, and composition-focused.

Capability modules should carry workflow-specific code.

This mapping should describe target ownership, not legacy file preservation.

## 3. Frontend Ownership Areas

### 3.1 Kernel Shell

The frontend kernel shell should own:
- contribution registration
- menu composition
- workflow route entry resolution
- shared working context access
- slot rendering orchestration

The shell should not own:
- detailed built-in workflow page rendering
- action form business logic

### 3.2 Built-In Capability Modules

Built-in capability modules should own:
- `Homepage` page contribution
- `Workloads` page contribution
- `Create` action contribution
- `Apply` action contribution

These modules are built-in for V1, but should already behave like capability contributors rather than shell-private code.

### 3.3 Shared Context Layer

The shared context layer should remain kernel-adjacent and own:
- cluster continuity
- namespace continuity
- workflow continuity
- list continuity
- action continuity

Capability modules should consume it but not redefine it.

### 3.4 Slot Layer

The slot layer should own:
- slot definitions
- slot lookup and rendering rules
- safe fallback behavior when no contribution exists

V1 can keep this narrow, but the ownership line should exist now.

## 4. Backend Ownership Areas

### 4.1 Backend Kernel

The backend kernel should own:
- capability registration
- capability metadata composition
- cluster-aware metadata resolution
- backend-authoritative execution entry points
- permission and policy enforcement

### 4.2 Built-In Capability Providers

Built-in providers should own:
- workload metadata contribution
- workload page capability metadata
- create/apply action capability metadata
- resource-operation execution details behind kernel entry points

### 4.3 Registry And Metadata Layer

The registry and metadata layer should own:
- capability discovery structures
- cluster-scoped menu/page/action visibility resolution
- transport-safe frontend metadata payloads

### 4.4 Execution Layer

The execution layer should own:
- action dispatch
- target validation
- permission check integration
- result summary generation

## 5. V1 Recommended Module Groups

Before wider implementation continues, the project should establish at least these conceptual module groups:

- frontend kernel shell
- frontend built-in capability modules
- frontend shared context layer
- frontend slot layer
- backend kernel registry/composition layer
- backend built-in capability providers
- backend execution layer

## 6. What Should Not Be Mapped Into The Kernel

The following should not be treated as kernel responsibilities:
- page-local form state
- workload table presentation details
- diagnostic panel layout
- AI assistant logic
- plugin marketplace behavior

These may later consume kernel contracts, but they are not kernel ownership.

## 7. V1 Mapping Readiness Check

Implementation may continue only when these conditions are true:
- built-in workflow code has a clear destination outside the shell core
- shared working context remains separate from capability modules
- backend execution remains separate from menu/page metadata composition
- slot ownership is defined even if the first slot set is small
