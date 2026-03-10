# KubeDeck Task A 代码变更计划

## 1. 目的

本文档定义 V1 Task A：共享工作上下文模型 的第一步代码规划。

它不实现代码，只定义首批应创建哪些文件、文件职责边界是什么、哪些现有代码在第一步先不动，以及在什么条件下允许完全替换旧代码。

## 2. 规划规则

本计划服从新的产品模型，而不是服从当前历史代码形态。

如果现有代码与新模型冲突：
- 以新模型为准
- 历史代码可以被收缩、绕开，或在后续完全替换

兼容旧结构本身不是目标。

## 3. Task A 目标

Task A 应产出首套规范的共享工作上下文基础，覆盖：
- active cluster
- namespace scope
- current workflow domain
- list continuity
- action continuity

当这些概念拥有一套统一状态来源，而不再散落在无关页面局部所有权中时，Task A 才算完成。

## 4. 首批文件创建集合

首个代码变更集合应创建这些文件：

- `frontend/src/state/work-context/types.ts`
- `frontend/src/state/work-context/clusterContext.ts`
- `frontend/src/state/work-context/namespaceContext.ts`
- `frontend/src/state/work-context/workflowContext.ts`
- `frontend/src/state/work-context/actionContext.ts`
- `frontend/src/state/work-context/listContext.ts`
- `frontend/src/state/work-context/events.ts`
- `frontend/src/state/work-context/selectors.ts`

在更深的页面改造开始之前，这组文件应先存在。

## 5. 每个新文件的职责

### `types.ts`

应定义：
- 共享工作上下文类型
- 根状态结构
- V1 关键 union 类型

不应定义：
- 页面渲染逻辑
- fetch 逻辑

### `clusterContext.ts`

应定义：
- cluster 状态切片
- cluster 更新逻辑
- 切换生命周期规则

不应定义：
- 除显式协同契约外的 namespace 默认规则
- 列表状态

### `namespaceContext.ts`

应定义：
- namespace scope 状态切片
- 浏览范围更新逻辑
- cluster 切换后的恢复/重置规则

不应定义：
- create/apply 执行目标的所有权

### `workflowContext.ts`

应定义：
- 当前工作流域状态
- workflow 进入与返回规则

不应定义：
- 列表筛选
- 操作结果详情

### `actionContext.ts`

应定义：
- 操作生命周期状态
- 执行目标状态
- 结果摘要状态

不应定义：
- namespace 浏览范围的事实来源

### `listContext.ts`

应定义：
- 与连续性相关的列表状态

不应定义：
- 纯表格展示细节

### `events.ts`

应定义：
- 共享工作上下文的公共事件 / 更新入口

不应定义：
- 页面布局逻辑

### `selectors.ts`

应定义：
- Homepage、Workloads、Create/Apply、结果反馈所需的 selectors

不应定义：
- 修改逻辑

## 6. 现有代码处理策略

在第一步中，应以保守方式处理现有代码，但不能把它视为权威。

### 初始可暂时不动

以下类别在首批文件创建阶段可以先保持不动：
- theme 与 theme mode 文件
- SDK parsing 文件，除非上下文结构必须依赖它们
- plugin host 文件
- page-shell 纯展示辅助组件

### 不得作为新模型设计来源

以下类别不能被当作新模型事实来源：
- 当前顶层页面组合方式
- 当前主入口里的本地状态分布
- 如果与新模型冲突的历史 state helper 形态

### 必要时允许替换

如果旧文件阻碍新模型落地，则可以：
- 停止继续扩展它们
- 绕开它们
- 或在后续阶段完全替换它们

Task A 不应为了保住这些文件而扭曲新架构。

## 7. 推荐执行顺序

Task A 的首个实现顺序应为：

1. 创建 `types.ts`
2. 创建 `clusterContext.ts`
3. 创建 `namespaceContext.ts`
4. 创建 `workflowContext.ts`
5. 创建 `actionContext.ts`
6. 创建 `listContext.ts`
7. 创建 `events.ts`
8. 创建 `selectors.ts`

只有在这些文件存在后，页面级接入才应开始。

## 8. Task A 此时不应做的事

Task A 此时不应：
- 重做 Homepage 视觉
- 构建 Workloads 页面 UI
- 实现 Create / Apply UI
- 扩展插件集成
- 引入 AI 相关上下文
- 优化视觉打磨

Task A 的本质是共享状态所有权，不是可见功能完整度。

## 9. 允许清空旧代码的规则

如果当前某个文件从根本上阻碍共享工作上下文模型干净落地，则在后续实现步骤中允许清空或替换该代码。

但必须满足以下前提：
- 新的替代结构已经存在
- 替代职责已经明确
- 被移除的代码已明确落在接受模型之外

简而言之：
- 不要为了“看起来安全”去保留错误所有权
- 也不要在替代结构未准备好前就删除旧代码

## 10. Task A 进入实现前提

只有满足以下条件，Task A 才算可以进入真实代码工作：
- 新文件集合被接受
- 文件职责被接受
- 历史代码被明确视为可选而非强制绑定
- 团队同意：当旧结构与新模型冲突时，可以替换旧结构
