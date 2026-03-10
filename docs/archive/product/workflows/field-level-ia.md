# KubeDeck Homepage And Workloads Field-Level IA Draft

## 1. Purpose

This document defines the field-level information architecture for the Homepage and Workloads page in the first core workflow.

The goal is to make sure both pages prioritize task entry and task completion over shell diagnostics or framework visibility.

## 2. Field Priority Model

Field priority levels in this document:
- `P0`: must be visible on first screen without extra interaction
- `P1`: important for task completion, but may sit below the first visual focus
- `P2`: supportive or contextual, can be collapsed, secondary, or hidden by default

## 3. Homepage IA

### 3.1 Homepage Primary Goal

The homepage must help the user:
- confirm current operating context,
- identify the next likely task,
- and enter the first workflow immediately.

### 3.2 Homepage Blocks

The homepage should contain these blocks:
- global header
- current context block
- primary task entry block
- recent or default task block
- blocking notice block

It should not use diagnostics as a leading block.

### 3.3 Global Header

Fields:
- product name (`P1`)
- user identity or session entry (`P2`)
- global settings entry (`P2`)

Required first-screen visibility:
- product name

May be downgraded:
- settings
- non-critical account details

### 3.4 Current Context Block

Fields:
- active cluster (`P0`)
- namespace scope or last-used namespace (`P1`)
- current environment label if applicable (`P1`)
- cluster change action (`P0`)

Required first-screen visibility:
- active cluster
- cluster change action

May be downgraded:
- secondary environment labeling
- extended cluster metadata

### 3.5 Primary Task Entry Block

Fields:
- primary destination label, such as `Workloads` (`P0`)
- short explanation of the destination (`P1`)
- entry action button or clickable card (`P0`)
- optional default next action hint (`P1`)

Required first-screen visibility:
- primary destination label
- primary entry action

May be downgraded:
- explanatory helper copy if space is limited

### 3.6 Recent Or Default Task Block

Fields:
- most recent task or page (`P1`)
- resume action (`P1`)
- default fallback task (`P1`)

Required first-screen visibility:
- none, if screen space is constrained

May be downgraded:
- all recent activity details

### 3.7 Blocking Notice Block

Fields:
- blocking problem summary (`P1`)
- recommended next action (`P1`)
- dismissal or retry action (`P2`)

Required first-screen visibility:
- only if the issue prevents task entry

Must be downgraded or hidden by default:
- non-blocking shell diagnostics
- raw status probes
- backend endpoint details

### 3.8 Homepage Fields To Hide Or Demote

These should not be first-screen homepage fields:
- raw API target
- registry type count
- healthz / readyz as headline indicators
- failure summary for backend shell checks
- navigation grouped by technical source such as `system`, `user`, `dynamic`

## 4. Workloads Page IA

### 4.1 Workloads Page Primary Goal

The Workloads page must help the user:
- confirm current operating context,
- inspect the relevant resource set,
- and begin a standard operation quickly.

### 4.2 Workloads Page Blocks

The Workloads page should contain these blocks:
- page header
- context and scope bar
- filter and search bar
- workload list block
- primary action block
- result or refresh status block

### 4.3 Page Header

Fields:
- page title such as `Workloads` (`P0`)
- page purpose subtitle (`P2`)
- current cluster badge or label (`P0`)

Required first-screen visibility:
- page title
- current cluster

May be downgraded:
- descriptive subtitle

### 4.4 Context And Scope Bar

Fields:
- namespace selector or scope display (`P0`)
- current resource subtype or workload category (`P1`)
- active filters summary (`P1`)
- clear filters action (`P1`)

Required first-screen visibility:
- namespace selector or scope display

May be downgraded:
- verbose filter explanations

### 4.5 Filter And Search Bar

Fields:
- keyword search (`P1`)
- status filter (`P1`)
- resource type filter if needed (`P2`)
- sort control (`P2`)

Required first-screen visibility:
- search
- the one or two highest-value filters

May be downgraded:
- advanced sort controls
- low-frequency filters

### 4.6 Workload List Block

Fields per row or item:
- resource name (`P0`)
- kind or type (`P1`)
- namespace if not globally fixed (`P1`)
- status (`P0`)
- age or last updated time (`P2`)
- summary health or readiness indicator (`P1`)
- quick row action entry (`P2`)

Required first-screen visibility in the list:
- resource name
- status
- enough context to know what object is being viewed

May be downgraded:
- secondary timestamps
- rarely used quick actions

### 4.7 Primary Action Block

Fields:
- `Create` action (`P0`)
- `Apply` action (`P0`)
- refresh action (`P1`)

Required first-screen visibility:
- at least one primary operation entry

May be downgraded:
- refresh if there is already an automatic or inline refresh affordance

### 4.8 Result Or Refresh Status Block

Fields:
- last operation result summary (`P1`)
- last refresh status (`P2`)
- retry action if needed (`P2`)

Required first-screen visibility:
- only when the information is directly relevant to the user just after an action

Must be downgraded:
- persistent shell-level diagnostics unrelated to the current workload task

## 5. First-Screen Visibility Rules

### Homepage P0 Fields

These must be visible without scrolling or opening secondary panels:
- active cluster
- cluster change action
- primary task label
- primary task entry action

### Workloads Page P0 Fields

These must be visible without scrolling or opening secondary panels:
- page title
- current cluster
- namespace selector or namespace scope
- workload list entry area
- primary operation entry such as `Create` or `Apply`
- visible workload name and status

## 6. Demotion Rules

Information should be demoted or hidden when it:
- serves implementation validation more than user action,
- does not affect the next immediate step,
- duplicates context already shown elsewhere,
- or exposes technical source categories instead of task categories.

## 7. Scope Boundary

This IA draft does not yet define:
- exact visual layout,
- typography,
- spacing,
- responsive breakpoints,
- or component library decisions.

It defines only what information should appear, in what hierarchy, and with what task priority.
