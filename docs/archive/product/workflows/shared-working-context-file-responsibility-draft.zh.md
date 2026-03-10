# KubeDeck 共享工作上下文文件职责草案

## 1. 目的

本文档定义实现共享工作上下文时，文件级职责应如何拆分。

它刻意从产品文档和实现映射倒推，而不是从当前历史代码布局出发。

## 2. 设计规则

文件职责必须服从工作流边界和上下文边界。

它们不应围绕以下因素组织：
- 临时开发便利，
- 一个过大的 `App` 容器，
- 或布局、数据获取、状态逻辑的历史性混杂。

## 3. 推荐的前端职责区域

前端应拆分为这些职责区域：
- app shell 组合
- 共享工作上下文状态
- 工作流页面
- 操作界面
- SDK / API 访问
- 页面局部展示状态

## 4. 推荐的文件分组

### 4.1 共享工作上下文状态文件

推荐分组：
- `frontend/src/state/work-context/`

推荐文件：
- `clusterContext.ts`
- `namespaceContext.ts`
- `workflowContext.ts`
- `listContext.ts`
- `actionContext.ts`
- `events.ts`
- `selectors.ts`
- `types.ts`

职责：
- 定义共享状态模型
- 定义更新入口
- 定义页面消费用的 selector

不应包含：
- 页面布局代码
- 直接组件渲染
- 与上下文归属无关的任意 fetch 编排

### 4.2 工作流页面文件

推荐分组：
- `frontend/src/pages/`

推荐文件：
- `Homepage.tsx`
- `WorkloadsPage.tsx`

职责：
- 组合页面区域
- 消费共享上下文
- 触发与工作流相关的事件

不应包含：
- 共享状态的最终事实来源
- 跨域状态所有权

### 4.3 操作界面文件

推荐分组：
- `frontend/src/features/actions/`

推荐文件：
- `CreateApplySurface.tsx`
- `CreateApplyValidation.ts`
- `CreateApplyResult.tsx`

职责：
- 处理 create/apply 任务 UI
- 消费操作连续性状态
- 触发操作生命周期事件

不应包含：
- cluster 所有权
- namespace 浏览范围所有权

### 4.4 上下文感知组件

推荐分组：
- `frontend/src/components/context/`

推荐文件：
- `ClusterSelector.tsx`
- `NamespaceScopeSelector.tsx`
- `ContextSummary.tsx`

职责：
- 渲染可复用的上下文控件与摘要

不应包含：
- 全局工作流业务规则

### 4.5 工作流导向页面组件

推荐分组：
- `frontend/src/components/workflows/`

推荐文件：
- `HomepagePrimaryTaskCard.tsx`
- `WorkloadsList.tsx`
- `WorkloadsToolbar.tsx`
- `ActionResultBanner.tsx`

职责：
- 渲染工作流特定的 UI 区块
- 通过 props 或 selectors 被驱动

不应包含：
- 共享上下文事实来源的所有权

### 4.6 SDK / API 访问文件

推荐分组：
- `frontend/src/sdk/`

推荐文件：
- 保留 API transport 与 parsing 在这里
- 必要时补充工作流专用 client，例如：
  - `clustersApi.ts`
  - `workloadsApi.ts`
  - `applyApi.ts`

职责：
- 数据获取、解析、传输契约

不应包含：
- 共享上下文状态所有权
- 页面导航规则

## 5. 按域划分的文件职责规则

### 5.1 Cluster 职责

应放在：
- work-context 状态文件
- 上下文感知 selector 与控件组件

不应放在：
- 页面本地组件状态作为事实来源

### 5.2 Namespace 职责

应放在：
- work-context 状态文件

可被渲染在：
- Homepage 摘要
- Workloads 范围控件
- Create/Apply 目标确认

不应放在：
- 没有连续性契约的孤立页面本地状态

### 5.3 Workflow Domain 职责

应放在：
- workflow context 状态文件

可被消费在：
- Homepage 入口逻辑
- Workloads 页面
- 返回链路处理

不应放在：
- 任意布局容器

### 5.4 列表连续性职责

应放在：
- list context 状态文件

可被消费在：
- Workloads 页面及其直接子组件

不应放在：
- 全局 shell 布局
- 无关页面

### 5.5 操作连续性职责

应放在：
- action context 状态文件

可被消费在：
- Create/Apply 操作界面
- 结果反馈界面
- Workloads 返回状态

不应放在：
- 无关页面 UI

## 6. 推荐的 V1 文件创建顺序

对 V1，建议按以下顺序引入文件：

1. `frontend/src/state/work-context/types.ts`
2. `frontend/src/state/work-context/clusterContext.ts`
3. `frontend/src/state/work-context/namespaceContext.ts`
4. `frontend/src/state/work-context/workflowContext.ts`
5. `frontend/src/state/work-context/actionContext.ts`
6. `frontend/src/state/work-context/listContext.ts`
7. `frontend/src/state/work-context/events.ts`
8. `frontend/src/state/work-context/selectors.ts`
9. `frontend/src/pages/Homepage.tsx`
10. `frontend/src/pages/WorkloadsPage.tsx`

Create/Apply 操作界面应在共享上下文文件被接受后再跟进。

## 7. 本草案明确避免的事情

本草案不预设：
- 当前 `App.tsx` 应继续作为主编排面
- 当前 state 文件就是最终形态
- 现有文件拆分就是正确职责模型

如果当前文件与这份职责模型冲突，应以模型为准。

## 8. 进入代码规划前提

只有满足以下条件，这份文件职责草案才算可以进入实现规划：
- 状态模块拆分被接受
- 页面职责拆分被接受
- 团队同意让工作流边界主导文件所有权
