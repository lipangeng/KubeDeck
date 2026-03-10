# KubeDeck 共享工作上下文事件与更新规则

## 1. 目的

本文档定义哪些事件被允许更新共享工作上下文、每个事件可以修改哪些字段、不能修改哪些字段，以及事件完成后必须满足哪些状态约束。

它的存在目的是在进入代码映射之前，让共享上下文在行为层面保持安全和一致。

## 2. 事件设计规则

所有事件必须遵守以下规则：
- 一个事件只能修改为完成其目的所必需的最小字段集合
- 事件不能静默重置无关的连续性状态
- cluster、namespace、浏览范围与执行目标必须保持逻辑分离
- 操作类事件不能重写浏览上下文，除非用户明确选择了该变化

## 3. 事件清单

首条任务流应允许以下事件：

1. `enter_homepage`
2. `request_cluster_switch`
3. `complete_cluster_switch`
4. `fail_cluster_switch`
5. `enter_workloads`
6. `update_namespace_scope`
7. `update_list_context`
8. `start_action`
9. `validate_action`
10. `fail_action_validation`
11. `resolve_execution_target`
12. `submit_action`
13. `complete_action_success`
14. `complete_action_partial_failure`
15. `complete_action_failure`
16. `acknowledge_action_result`
17. `return_to_workloads`

## 4. 事件规则

### 4.1 `enter_homepage`

可修改：
- `currentWorkflowDomain`

不可修改：
- `activeCluster`
- `namespaceScope`
- `listContext`
- `actionContext`

事件后约束：
- 共享上下文仍然有效
- 不能因为进入 Homepage 就丢失任务连续性状态

### 4.2 `request_cluster_switch`

可修改：
- `activeCluster.status`

不可修改：
- `activeCluster.id`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- `actionContext`

事件后约束：
- `activeCluster.status` 进入 `switching`
- 旧的稳定上下文在切换成功或失败前仍然可用

### 4.3 `complete_cluster_switch`

可修改：
- `activeCluster.id`
- `activeCluster.status`
- `activeCluster.lastStableId`
- `namespaceScope`
- `actionContext`

不可修改：
- 对于仍然有效的同一工作流域，不能无故修改无关的 `listContext`

事件后约束：
- `activeCluster.status` 变为 `ready`
- `namespaceScope` 已恢复或安全重置
- 不兼容的 `actionContext` 被清空
- 不存在旧 cluster 与新 cluster 混杂的上下文

### 4.4 `fail_cluster_switch`

可修改：
- `activeCluster.status`

不可修改：
- `activeCluster.id`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- `actionContext`

事件后约束：
- `activeCluster.status` 变为 `failed`，或安全回退到上一个 `ready`
- 最近一次稳定上下文仍然可用
- 用户不会停留在未定义混合状态

### 4.5 `enter_workloads`

可修改：
- `currentWorkflowDomain`

不可修改：
- `activeCluster`
- `namespaceScope`
- `actionContext`

可保留：
- `listContext`

事件后约束：
- `currentWorkflowDomain.id` 为 `workloads`
- cluster 与 namespace 连续性保持不变

### 4.6 `update_namespace_scope`

可修改：
- `namespaceScope`

不可修改：
- `activeCluster`
- `currentWorkflowDomain`
- `actionContext.executionTarget`

可触发重新校验：
- `actionContext`

事件后约束：
- namespace 浏览范围变为新选择值
- 浏览范围与执行目标依然分离
- 若已有操作上下文，应标记为待重新校验，而不是静默重写

### 4.7 `update_list_context`

可修改：
- `listContext`

不可修改：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `actionContext`

事件后约束：
- `listContext` 只反映当前工作流域下的浏览上下文
- 临时视觉状态不能进入共享上下文

### 4.8 `start_action`

可修改：
- `actionContext.actionType`
- `actionContext.originDomain`
- `actionContext.status`

不可修改：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`

事件后约束：
- `actionContext.status` 变为 `editing`
- 浏览上下文不变

### 4.9 `validate_action`

可修改：
- `actionContext.status`

不可修改：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- `actionContext.executionTarget`

事件后约束：
- `actionContext.status` 变为 `validating`
- 执行目标在显式收敛前不能被隐式改写

### 4.10 `fail_action_validation`

可修改：
- `actionContext.status`
- 若后续引入，也可修改校验相关元数据

不可修改：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- 已有效的浏览范围

事件后约束：
- `actionContext.status` 回到 `editing`
- 用户输入仍然可恢复
- 浏览连续性不丢失

### 4.11 `resolve_execution_target`

可修改：
- `actionContext.executionTarget`

不可修改：
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- `activeCluster`

事件后约束：
- 执行目标必须具体化
- 若为 namespaced 资源，则目标 namespace 明确
- `all` 和未收敛的 `multiple` 绝不能作为执行目标存入

### 4.12 `submit_action`

可修改：
- `actionContext.status`

不可修改：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- `actionContext.executionTarget`

事件后约束：
- `actionContext.status` 变为 `submitting`
- 提交前执行目标已经是明确值

### 4.13 `complete_action_success`

可修改：
- `actionContext.status`
- `actionContext.resultSummary`

不可修改：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`

事件后约束：
- `actionContext.status` 变为 `success`
- 结果摘要可用
- 浏览连续性保持不变

### 4.14 `complete_action_partial_failure`

可修改：
- `actionContext.status`
- `actionContext.resultSummary`

不可修改：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`

事件后约束：
- `actionContext.status` 变为 `partial_failure`
- 成功与失败结果必须可区分
- 浏览连续性保持不变

### 4.15 `complete_action_failure`

可修改：
- `actionContext.status`
- `actionContext.resultSummary`

不可修改：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`

事件后约束：
- `actionContext.status` 变为 `failure`
- 结果摘要可用
- 失败不能破坏浏览连续性

### 4.16 `acknowledge_action_result`

可修改：
- `actionContext`

不可修改：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`

事件后约束：
- 确认后允许清空已完成的操作结果状态
- 浏览连续性保持不变

### 4.17 `return_to_workloads`

可修改：
- `currentWorkflowDomain`
- `actionContext`

不可修改：
- `activeCluster`
- `namespaceScope`
- `listContext`

事件后约束：
- 用户返回 `workloads`
- cluster 与 namespace 保持不变
- `listContext` 仍然可用
- `actionContext` 要么为了反馈展示保留，要么在确认后安全清空

## 5. 禁止的更新模式

以下行为被禁止：
- cluster 切换时无理由静默修改列表筛选
- namespaceScope 更新时静默修改 executionTarget
- action submit 修改浏览范围
- action failure 清空 cluster 或 namespace 上下文
- 回到 Homepage 时重置工作流连续性
- 把壳层诊断信息当作共享工作上下文更新

## 6. 全局状态约束

任意事件完成后，以下约束必须始终成立：
- `activeCluster` 始终有定义
- `namespaceScope` 始终从属于当前 active cluster
- `currentWorkflowDomain` 不能与当前任务路径冲突
- 允许提交时，`executionTarget` 绝不能是模糊值
- 浏览范围与执行目标绝不能由同一个字段表示

## 7. V1 最小事件支持

对于 V1，最小必需实现的事件集合是：
- `request_cluster_switch`
- `complete_cluster_switch`
- `fail_cluster_switch`
- `enter_workloads`
- `update_namespace_scope`
- `update_list_context`
- `start_action`
- `resolve_execution_target`
- `submit_action`
- `complete_action_success`
- `complete_action_failure`
- `return_to_workloads`

其余事件可以先停留在概念层，或在不破坏 V1 连续性的前提下做简化。

## 8. 进入代码映射前提

只有满足以下条件，这个事件模型才算可以进入实现映射：
- 事件列表被接受
- 禁止的更新模式被接受
- V1 最小事件支持集合被接受
