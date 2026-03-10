# KubeDeck Homepage And Workloads Low-Fidelity Page Spec

## 1. Purpose

This document converts the current product requirements, workflow draft, state rules, and field-level IA into a low-fidelity page specification for the Homepage and Workloads page.

The goal is not visual design polish. The goal is to define what the user should see first, what each area must do, and how the page should support the first core workflow.

## 2. Scope

This low-fidelity specification covers:
- Homepage
- Workloads page

It does not yet define:
- final visual style,
- final component choices,
- responsive pixel-level layout,
- or high-fidelity interaction motion.

## 3. Homepage Low-Fidelity Spec

### 3.1 Page Goal

The Homepage must:
- establish current operating context,
- present the primary entry into the first workflow,
- and avoid competing with that task using diagnostics-heavy content.

### 3.2 Page Zones

The Homepage should contain these top-level zones in order:

1. Global header
2. Current context zone
3. Primary task entry zone
4. Resume / default task zone
5. Blocking notice zone

### 3.3 Wireframe-Level Structure

```text
+--------------------------------------------------------------+
| Header: Product / Session / Secondary settings               |
+--------------------------------------------------------------+
| Current Context                                               |
| Cluster: [prod v]   Namespace: default / last used           |
+--------------------------------------------------------------+
| Primary Task Entry                                            |
| Workloads                                                     |
| Browse and operate workloads in the current cluster           |
| [Enter Workloads]                                             |
+--------------------------------------------------------------+
| Resume Or Default Task                                        |
| Continue last context or use default task entry               |
+--------------------------------------------------------------+
| Blocking Notice (only when needed)                            |
| Problem summary / Next action / Retry                         |
+--------------------------------------------------------------+
```

### 3.4 Zone Responsibilities

#### Global Header

Must do:
- identify the product,
- expose minimal global access points,
- stay visually secondary to the task entry flow.

Must not do:
- carry task-specific diagnostics,
- carry cluster-specific operational details as the main message.

#### Current Context Zone

Must do:
- show current `activeCluster`,
- let the user switch cluster,
- show the minimum namespace context needed before entering work.

Must not do:
- overload the user with deep cluster metadata,
- display technical identifiers as the dominant content.

#### Primary Task Entry Zone

Must do:
- make `Workloads` the clearest next step,
- explain in one short line what happens after entry,
- provide a single clear action.

Must not do:
- compete with multiple equal-priority destinations,
- force the user to interpret technical source categories.

#### Resume / Default Task Zone

Must do:
- help returning users continue the likely next step,
- provide a fallback if no recent context exists.

Must not do:
- overshadow the primary task entry,
- become a history-heavy dashboard.

#### Blocking Notice Zone

Must do:
- appear only when the first workflow is genuinely blocked,
- summarize the blocker and the next action.

Must not do:
- surface non-blocking shell diagnostics,
- persist as the main content when no blocking issue exists.

### 3.5 Homepage Acceptance Conditions

The Homepage is acceptable when:
- a user can identify current cluster immediately,
- a user can enter `Workloads` without confusion,
- task entry is more prominent than diagnostics,
- and no source-based navigation grouping is required for basic use.

## 4. Workloads Page Low-Fidelity Spec

### 4.1 Page Goal

The Workloads page must:
- become the main operational workspace,
- make cluster and namespace context explicit,
- and let the user begin a standard action quickly.

### 4.2 Page Zones

The Workloads page should contain these top-level zones in order:

1. Page header
2. Context and scope bar
3. Filters and search bar
4. Primary action zone
5. Workload list zone
6. Result / status zone

### 4.3 Wireframe-Level Structure

```text
+--------------------------------------------------------------+
| Header: Workloads                         Cluster: prod      |
+--------------------------------------------------------------+
| Scope Bar                                                     |
| Namespace: [default v]   Filters: Running, Error   [Clear]   |
+--------------------------------------------------------------+
| Search / Filters                                               |
| [Search........] [Status v] [Type v]                           |
+--------------------------------------------------------------+
| Actions                                                        |
| [Create] [Apply] [Refresh]                                     |
+--------------------------------------------------------------+
| Workload List                                                  |
| Name        Kind        Namespace      Status      Health      |
| api         Deployment  default        Running     Ready       |
| web         Deployment  default        Pending     Warning     |
+--------------------------------------------------------------+
| Result / Status (contextual only)                              |
| Last action succeeded / partial failed / failed               |
+--------------------------------------------------------------+
```

### 4.4 Zone Responsibilities

#### Page Header

Must do:
- confirm the user is in `Workloads`,
- restate current cluster clearly.

Must not do:
- contain large volumes of secondary explanation,
- become a dashboard summary panel.

#### Context And Scope Bar

Must do:
- make namespace scope explicit,
- show current filter state,
- allow quick context correction.

Must not do:
- hide namespace until deep interaction,
- fragment context across multiple unrelated panels.

#### Search And Filters Bar

Must do:
- support fast narrowing of visible resources,
- surface the one or two filters most relevant to the first workflow.

Must not do:
- overload the user with advanced filtering before basic flow exists.

#### Primary Action Zone

Must do:
- expose `Create` and `Apply` as the first operation entries,
- make action availability obvious.

Must not do:
- bury the primary action inside row menus or secondary controls.

#### Workload List Zone

Must do:
- show a real resource list,
- give enough identifying information to act on the right object,
- show current status clearly.

Must not do:
- optimize first for metrics, charts, or overview cards,
- require drill-in before users can understand what objects exist.

#### Result / Status Zone

Must do:
- show recent action outcome when it matters,
- connect the outcome back to the current list context.

Must not do:
- permanently occupy prime page space when no action has occurred,
- show shell-level diagnostics unrelated to the workload task.

### 4.5 Workloads Acceptance Conditions

The Workloads page is acceptable when:
- the user can identify current cluster and namespace immediately,
- the user can see a real list of resources,
- the user can start `Create` or `Apply` without hunting,
- and the result of an action can be understood without leaving the task path.

## 5. Cross-Page Consistency Rules

Both pages must follow these rules:
- cluster context must stay visible at page entry,
- source-based navigation labels must not be primary UX language,
- diagnostics must be secondary unless they block the current workflow,
- and the page must answer “what do I do next?” before “how is the shell implemented?”

## 6. Open Decisions For The Next Spec Layer

The next document should decide:
- whether `Create / Apply` is a page, drawer, or dialog,
- how namespace persistence works between Homepage and Workloads,
- what exact columns or cards appear in the first workload list,
- and how contextual result feedback is visually attached to the task flow.
