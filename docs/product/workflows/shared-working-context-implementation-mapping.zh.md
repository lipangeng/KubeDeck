# KubeDeck 共享工作上下文实现映射草案

## 1. 目的

本文档将共享工作上下文模型、状态结构和事件规则映射为面向实现的结构草案。

它仍然是代码前阶段文档。它的作用是在实现开始前，先定义应该存在什么样的状态模块、更新入口以及页面消费关系。

## 2. 映射原则

实现必须服从产品模型，而不是服从当前 UI 的偶然结构。

这份映射必须：
- 保持 cluster、namespace、workflow、列表连续性和操作连续性的分离
- 避免一个过大的全局状态桶
- 避免在无关页面中复制同一份上下文
- 在更大范围扩张前先支持 V1 首条工作流

## 3. 推荐的状态模块拆分

共享工作上下文应实现为少量相互协作的状态模块，而不是一个没有边界的大 store。

推荐的逻辑模块：
- cluster 上下文模块
- namespace 上下文模块
- workflow 上下文模块
- 列表连续性模块
- 操作连续性模块

## 4. 模块职责

### 4.1 Cluster 上下文模块

负责：
- `activeCluster.id`
- `activeCluster.status`
- `activeCluster.lastStableId`

处理事件：
- `request_cluster_switch`
- `complete_cluster_switch`
- `fail_cluster_switch`

不应负责：
- namespace 浏览规则
- 列表筛选
- create/apply 结果状态

### 4.2 Namespace 上下文模块

负责：
- `namespaceScope.mode`
- `namespaceScope.values`
- `namespaceScope.source`

处理事件：
- `update_namespace_scope`
- cluster 切换后的 namespace 恢复或重置

不应负责：
- create/apply 的执行目标
- 页面本地 namespace UI 展示细节

### 4.3 Workflow 上下文模块

负责：
- `currentWorkflowDomain.id`
- `currentWorkflowDomain.source`

处理事件：
- `enter_homepage`
- `enter_workloads`
- `return_to_workloads`

不应负责：
- 列表筛选本身
- 操作结果详情

### 4.4 列表连续性模块

负责：
- `listContext.searchText`
- `listContext.statusFilters`
- `listContext.subtypeFilters`
- `listContext.sortKey`
- `listContext.sortDirection`

处理事件：
- `update_list_context`
- 返回后在同一工作流域中的恢复

不应负责：
- 纯视觉表格状态
- 行 hover 或行展开细节

### 4.5 操作连续性模块

负责：
- `actionContext.actionType`
- `actionContext.originDomain`
- `actionContext.status`
- `actionContext.executionTarget`
- `actionContext.resultSummary`

处理事件：
- `start_action`
- `validate_action`
- `fail_action_validation`
- `resolve_execution_target`
- `submit_action`
- `complete_action_success`
- `complete_action_partial_failure`
- `complete_action_failure`
- `acknowledge_action_result`

不应负责：
- namespace 浏览范围
- 无关的列表浏览状态

## 5. 推荐的更新入口

实现应暴露按意图分组的显式更新入口，而不是允许任意直接修改。

推荐的更新入口分组：
- cluster 更新
- namespace 范围更新
- workflow 导航更新
- 列表连续性更新
- 操作生命周期更新

即使最终实现使用 reducer、store 或 controller function，这些分组也应与事件模型对应。

## 6. 推荐的页面消费映射

### 6.1 Homepage

Homepage 应消费：
- cluster 上下文
- namespace 上下文摘要
- workflow 上下文中的主入口语义

Homepage 应触发：
- cluster 更新
- workflow 进入更新

Homepage 不应消费：
- 除非明确加入恢复能力，否则不应消费详细列表连续性状态
- 正在进行中的操作草稿内部状态

### 6.2 Workloads 页面

Workloads 应消费：
- cluster 上下文
- namespace 上下文
- workflow 上下文
- 列表连续性上下文
- 从操作返回后的结果摘要

Workloads 应触发：
- namespace 范围更新
- 列表连续性更新
- 操作开始
- 必要时的 workflow 重进入

### 6.3 Create / Apply 操作界面

Create / Apply 应消费：
- cluster 上下文
- namespace 浏览上下文
- 操作连续性上下文

Create / Apply 应触发：
- 操作生命周期更新
- 执行目标收敛

Create / Apply 不应触发：
- 未经用户明确操作就直接改写浏览范围

### 6.4 结果反馈界面

结果反馈应消费：
- 操作连续性上下文
- cluster 上下文
- namespace 浏览上下文

结果反馈应触发：
- 结果确认
- 返回 Workloads

## 7. 跨模块协同规则

这些模块必须遵守以下协同规则：

- cluster 模块可要求 namespace 模块恢复或重置
- namespace 模块可要求 action 模块重新校验
- workflow 模块可根据域兼容性保留或清空列表连续性
- action 模块不能改写 namespace 浏览范围
- 列表连续性模块必须在操作完成和返回链路中存活

## 8. V1 实现建议

对 V1，以下模块应被视为必需：
- cluster 上下文模块
- namespace 上下文模块
- workflow 上下文模块
- 操作连续性模块

列表连续性模块即使首版字段很少，也必须先保留其结构位置。

## 9. 当前不应映射的内容

以下内容不应进入首版实现映射：
- 插件扩展状态
- AI 交互状态
- 恢复记忆系统
- 完整多域导航协同
- 高级分析或诊断状态

## 10. 进入代码级设计前提

只有满足以下条件，这份映射才算可以进入代码级设计：
- 模块拆分被接受
- 事件归属被接受
- 页面消费映射被接受
- 且 V1 范围仍然限制在首条工作流内
