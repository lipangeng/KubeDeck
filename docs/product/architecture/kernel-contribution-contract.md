# KubeDeck Kernel Contribution Contract Draft

## 1. Purpose

This document defines the minimum contribution contracts that the KubeDeck kernel must expose before broader feature implementation continues.

The goal is to prevent built-in workflow code from becoming permanent shell-only logic.

## 2. Contract Principle

The kernel should not own business workflow implementation.

The kernel should own:
- contribution registration,
- composition rules,
- runtime context boundaries,
- and permission-aware execution boundaries.

Built-in capabilities and future plugins should attach through the same contract families whenever practical.

## 3. Frontend Contract Families

The frontend kernel must expose these minimum contribution families.

### 3.1 Page Contribution

Purpose:
- register a navigable workflow page or domain entry

Must support:
- contribution identity
- workflow domain identity
- route or entry key
- display label metadata
- ownership of the rendered page component

Must not assume:
- shell-only page hard-coding
- one fixed locale string forever

### 3.2 Menu Contribution

Purpose:
- contribute one or more visible navigation entries into kernel-composed navigation

Must support:
- contribution identity
- menu placement metadata
- order and visibility metadata
- localizable label metadata
- mapping to a page or workflow entry

Must not assume:
- only built-in menu sources exist
- menu grouping is permanently tied to technical source categories

### 3.3 Action Contribution

Purpose:
- contribute executable actions into a workflow domain such as `Workloads`

Must support:
- action identity
- owning workflow domain
- display label metadata
- execution surface type metadata
- permission visibility hinting from backend-provided state

Must not assume:
- action execution is always defined directly inside a page component

### 3.4 Extension Slot Contribution

Purpose:
- allow a capability module to extend an existing workflow page with additional panels, summaries, or context blocks

Must support:
- slot identity
- placement metadata
- owning contribution identity
- visibility conditions

Must not require:
- full replacement semantics in V1

## 4. Backend Contract Families

The backend kernel must expose these minimum capability families.

### 4.1 Plugin Identity Contract

Purpose:
- identify a backend capability provider uniquely and stably

Must support:
- stable plugin or capability ID
- versionable ownership boundary

This is the thinnest required contract, but it is not sufficient by itself.

### 4.2 Capability Registration Contract

Purpose:
- register what a built-in capability module or future plugin provides

Must support:
- page or workflow capability metadata
- resource-domain capability metadata
- action capability metadata
- menu exposure metadata

### 4.3 Metadata Exposure Contract

Purpose:
- expose the kernel-composed menu, page, and action metadata to the frontend shell

Must support:
- cluster-aware metadata resolution
- permission-aware visibility filtering
- localizable display key or label metadata

### 4.4 Execution Entry Contract

Purpose:
- execute user-visible actions through backend-authoritative policy boundaries

Must support:
- action identity
- target context
- permission enforcement
- result summary shape

Must not assume:
- frontend-only authority

## 5. Built-In Capability Rule

Built-in workflow areas such as `Homepage`, `Workloads`, `Create`, and `Apply` may ship from the main repository, but they should still be modeled as capability contributions rather than permanent shell exceptions.

In practice, this means:
- the shell may bootstrap them,
- but the shell should not define their workflow semantics inline forever.

## 6. Localizable Metadata Rule

Every contract family that carries user-facing labels should allow a future-localizable representation.

At minimum, contract design should already tolerate:
- a label key,
- a description key,
- or a structured label field that can later resolve by locale.

V1 does not need a complete runtime translation system for all contracts, but the contract shape must not block one.

## 7. V1 Minimum Contract Set

Before broader implementation continues, the project should at least define:
- one page contribution shape
- one menu contribution shape
- one action contribution shape
- one slot contribution shape
- one backend capability registration shape
- one backend metadata exposure shape
- one backend action execution shape

This is the minimum kernel contract baseline.

## 8. Deferred Contract Depth

The following may be deferred after the first workflow becomes stable:
- plugin dependency resolution rules
- dynamic install and uninstall flows
- rich version negotiation
- slot replacement precedence across many contributors
- plugin sandboxing policy details

These are real future concerns, but they are not needed to establish the kernel contract family now.

## 9. Readiness Check

The contract draft is usable when these questions can be answered with "yes":
- Can `Homepage` and `Workloads` be described as page contributions?
- Can `Create` and `Apply` be described as action contributions?
- Can navigation be described as menu contributions instead of shell-owned lists?
- Can future side panels or summaries be described as slot contributions?
- Can backend execution remain authoritative through explicit action contracts?
