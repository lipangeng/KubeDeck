# KubeDeck 共享工作上下文状态结构草案

## 1. 目的

本文档将共享工作上下文模型进一步收敛为结构化状态草案。

它暂不定义代码层类型，而是定义预期的状态结构、必需字段、可选字段，以及 V1 期望。

## 2. 结构原则

该状态结构必须满足以下规则：
- 表达用户操作上下文，而不是表达偶然 UI 结构
- 明确区分浏览范围与执行目标
- 支持 Homepage、Workloads 与 Create / Apply 之间的连续性
- 保持足够小，避免沦为所有 UI 状态的垃圾桶

## 3. 根状态

共享工作上下文根结构应包含：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- `actionContext`

## 4. 字段结构

### 4.1 `activeCluster`

含义：
- 当前实际操作的 cluster

推荐结构：

```text
activeCluster:
  id: string
  status: ready | switching | failed
  lastStableId?: string
```

V1 必需：
- `id`
- `status`

V1 可选：
- `lastStableId`

### 4.2 `namespaceScope`

含义：
- 当前 active cluster 下的浏览范围

推荐结构：

```text
namespaceScope:
  mode: single | multiple | all
  values: string[]
  source: default | restored | user_selected
```

规则：
- `single` 必须只有一个值
- `multiple` 应有多个值
- `all` 可使用空 `values` 或忽略 `values`

V1 必需：
- `mode`
- `source`

V1 必需行为：
- 支持 `single`
- 支持 `all`

V1 可选：
- UI 层完整支持 `multiple`

### 4.3 `currentWorkflowDomain`

含义：
- 当前顶层工作流区域

推荐结构：

```text
currentWorkflowDomain:
  id: workloads | other_future_domain
  source: homepage_entry | direct_navigation | return_flow
```

V1 必需：
- `id`

V1 可选：
- `source`

V1 期望：
- `workloads` 是唯一必需域

### 4.4 `listContext`

含义：
- 用户返回当前工作流域时，为保证连续性所需的浏览上下文

推荐结构：

```text
listContext:
  searchText?: string
  statusFilters?: string[]
  subtypeFilters?: string[]
  sortKey?: string
  sortDirection?: asc | desc
```

V1 必需：
- 在 schema 层无强制字段

V1 必需能力：
- 即使首版字段很少，也必须保留这个结构入口

V1 指导：
- 只保留与连续性有关的字段
- 不要把临时展示状态放进来

### 4.5 `actionContext`

含义：
- 操作进行中或刚完成后所需保留的上下文

推荐结构：

```text
actionContext:
  actionType?: create | apply
  originDomain?: workloads
  status?: idle | editing | validating | submitting | success | partial_failure | failure
  executionTarget?:
    kind: namespace | cluster_scoped
    namespace?: string
  resultSummary?:
    outcome: success | partial_failure | failure
    affectedObjects?: string[]
    failedObjects?: string[]
```

V1 必需：
- 操作开始后必须有 `status`
- 提交前必须有 `executionTarget`

V1 必需能力：
- 显式执行目标收敛
- 操作完成后的结果摘要

V1 可选：
- 结果摘要中的丰富对象元数据

## 5. V1 必需字段

对 V1 而言，最小必需字段集合是：

```text
activeCluster.id
activeCluster.status
namespaceScope.mode
namespaceScope.source
currentWorkflowDomain.id
actionContext.status
actionContext.executionTarget
```

其余字段可以先保持最小，只在工作流连续性真正需要时再扩展。

## 6. 字段归属

属于共享工作上下文的字段：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- 与连续性相关的 `listContext`
- 与连续性相关的 `actionContext`

不属于共享工作上下文的字段：
- modal 可见性
- drawer 可见性
- 临时 hover/focus 状态
- 非阻塞性诊断信息
- 纯本地草稿格式状态

## 7. 迁移预期

### 7.1 Cluster 切换

当 cluster 切换时：
- `activeCluster.status` 进入 `switching`
- `namespaceScope` 恢复或安全重置
- 不兼容的 `actionContext` 清空
- 若仍然有效，`currentWorkflowDomain` 可保留

### 7.2 进入工作流

当从 Homepage 进入 Workloads 时：
- `currentWorkflowDomain.id` 变为 `workloads`
- 共享 cluster 与 namespace 保持不变

### 7.3 操作开始

当打开 Create / Apply 时：
- 设置 `actionContext.actionType`
- `actionContext.status` 进入 `editing`

### 7.4 校验

提交前：
- `actionContext.status` 进入 `validating`
- 只有在 `executionTarget` 已明确后，才允许进入 `submitting`

### 7.5 提交

提交时：
- `actionContext.status` 进入 `submitting`

完成后：
- `actionContext.status` 变为 `success`、`partial_failure` 或 `failure`
- `resultSummary` 可用

### 7.6 返回链路

返回 Workloads 时：
- `currentWorkflowDomain.id` 保持 `workloads`
- `namespaceScope` 保持不变
- `listContext` 保持不变
- `actionContext` 可在确认后清空

## 8. 校验约束

该状态结构必须支持以下产品约束：
- 浏览范围可以是 `all`
- 执行目标不能是模糊值
- `all` 绝不能成为写入目标
- cluster 切换后不能留下未定义混合状态
- 操作失败不能破坏浏览连续性

## 9. 延后扩展能力

该结构应为后续能力预留空间，而不需要改动根模型：
- 更丰富的 `multiple` namespace 行为
- workload 详情上下文
- 更多工作流域
- 恢复状态与任务记忆
- 若 V1 后产品范围调整，再考虑 AI 相关上下文

## 10. 进入代码映射前提

只有满足以下条件，这个结构才算可以进入实现映射：
- V1 最小字段集合被接受
- namespace scope 语义被接受
- execution target 语义被接受
- 并且没有人试图把共享工作上下文变成所有本地 UI 状态的容器
