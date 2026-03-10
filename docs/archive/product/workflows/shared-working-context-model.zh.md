# KubeDeck 共享工作上下文模型规范

## 1. 目的

本文档定义 KubeDeck 的共享工作上下文模型。

这是一份产品级规范，不是实现描述。

它的目的是说明：在首条任务流中，哪些状态必须跨页面保留，哪些状态属于页面局部，以及允许的状态迁移边界是什么。

## 2. 设计原则

共享工作上下文应表示用户当前的操作上下文，而不是表示当前 UI 的实现方式。

它应回答：
- 用户当前在哪个操作环境中，
- 用户当前在什么范围内工作，
- 用户当前位于哪个工作流域，
- 以及哪些上下文必须在导航和操作后保留下来。

它不应吸收所有临时 UI 细节。

## 3. 上下文层级

共享工作上下文应遵循以下层级：

1. active cluster
2. namespace scope
3. 当前工作流域
4. 工作流连续性状态

这意味着：
- cluster 是顶层锚点，
- namespace 从属于 cluster，
- 页面或资源域从属于 cluster 与 namespace，
- 连续性状态只服务于工作流保留。

## 4. 共享上下文字段

共享工作上下文应包含以下字段。

### 4.1 Active Cluster

字段：
- `activeCluster`

含义：
- 用户当前实际操作的 cluster

为什么共享：
- 所有页面和操作都依赖它

### 4.2 Namespace Scope

字段：
- `namespaceScope`

含义：
- 当前 active cluster 下的 namespace 浏览范围

预期结构：
- `single`
- `multiple`
- `all`

V1 要求：
- 产品实际行为支持 `single` 与 `all`
- 模型层保留对 `multiple` 的兼容性

为什么共享：
- 它必须在 Homepage、Workloads、详情页与操作界面之间持续存在

### 4.3 当前工作流域

字段：
- `currentWorkflowDomain`

含义：
- 用户当前所在的主要工作区域，例如 `workloads`

为什么共享：
- 它定义了用户当前位于首条任务流的哪个区域
- 它帮助操作完成后返回到正确位置

### 4.4 列表上下文

字段：
- `listContext`

含义：
- 为了在导航或操作后继续当前任务而保留的最小浏览状态

可包含：
- 搜索词
- 已选状态筛选
- 已选子类型筛选
- 排序方式

为什么共享：
- 用户在 create/apply 后应能回到原有工作列表上下文

约束：
- 只有与连续性相关的列表状态才应进入共享上下文
- 纯展示型或临时 UI 状态不应进入

### 4.5 操作上下文

字段：
- `actionContext`

含义：
- 对于正在进行或刚完成的任务操作，所需保留的最小持续上下文

可包含：
- 操作类型，例如 `create` 或 `apply`
- 来源工作流域
- 已收敛出的执行目标摘要
- 最近一次操作结果摘要

为什么共享：
- 成功/失败反馈与返回链路都依赖它

约束：
- 除非明确要求用于恢复，否则草稿内容本身不必默认进入共享上下文

## 5. 什么不属于共享上下文

以下内容默认不应属于共享工作上下文：
- 弹窗开关状态
- 抽屉宽度或纯视觉布局状态
- hover 状态
- 行展开状态
- 纯本地表单格式状态
- 诊断探针时间戳
- 与工作流连续性无关的壳层状态组件

除非它们成为跨页面恢复所必需，否则这些状态应留在页面局部或组件局部。

## 6. 页面对共享上下文的职责

### 6.1 Homepage

Homepage 可读取：
- `activeCluster`
- `namespaceScope` 摘要
- `currentWorkflowDomain`，但仅在需要支持恢复语义时

Homepage 可更新：
- `activeCluster`
- 进入首个工作流域

Homepage 不应拥有：
- 详细列表状态
- 操作草稿状态

### 6.2 Workloads 页面

Workloads 可读取：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `listContext`
- 相关操作结果摘要

Workloads 可更新：
- `currentWorkflowDomain`
- `listContext`
- namespace scope
- 操作上下文的起点

### 6.3 Create / Apply 操作界面

Create / Apply 可读取：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `actionContext`

Create / Apply 可更新：
- 已收敛出的执行目标
- 操作状态
- 操作结果摘要

Create / Apply 不应覆盖：
- 用户的浏览上下文，除非用户主动修改

### 6.4 结果反馈

结果反馈可读取：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- `actionContext`

结果反馈可更新：
- 仅限返回链路所需的确认或完成标记

## 7. 保留规则

以下字段必须在首条任务流中持续保留：
- `activeCluster`
- `namespaceScope`
- `currentWorkflowDomain`
- 与连续性相关的 `listContext`

以下字段只需在一次操作周期内保留：
- `actionContext`
- 操作结果摘要

以下字段可以自由重置：
- 纯本地 UI 展示状态

## 8. 迁移规则

### 8.1 Cluster 切换

当 `activeCluster` 变化时：
- namespace scope 必须重新验证或安全重置
- 如果目标 cluster 支持同一工作流路径，则当前工作流域可保留
- 不兼容的操作上下文必须清空

### 8.2 Namespace Scope 切换

当 `namespaceScope` 变化时：
- 如果页面支持新范围，则浏览上下文继续有效
- 提交前操作上下文可能需要重新校验

### 8.3 工作流域切换

当 `currentWorkflowDomain` 变化时：
- cluster 与 namespace 保持不变
- 如果新工作流域与旧域不同，则页面专属的列表上下文可以重置

### 8.4 操作开始

当 create/apply 开始时：
- 浏览上下文必须保持
- 创建操作上下文
- namespace 收敛必须朝着明确执行目标推进

### 8.5 操作完成

当 create/apply 完成时：
- 操作结果摘要可用
- 用户返回到原工作流域
- 浏览上下文保持不变，除非用户主动修改

## 9. 模型中的 Namespace 要求

模型必须明确区分：
- 浏览范围
- 执行目标

这非常关键，因为：
- 浏览可使用 `all`
- 执行不能使用模糊的 `all`

因此：
- `namespaceScope` 表示浏览上下文
- 执行目标应属于 `actionContext`

这两者不能被折叠成同一个字段。

## 10. V1 模型要求

对 V1 而言，这个共享模型至少必须足以支持：
- Homepage -> Workloads 进入链路
- Workloads 浏览连续性
- 带明确目标收敛的 create/apply
- 返回 Workloads 时不丢失 cluster 和 namespace 上下文

如果这个模型无法支持这四个结果，它就不能作为 V1 可接受模型。

## 11. 实现映射规则

只有在这个模型被接受后，才应创建实现映射。

实现应当适配这个模型。
模型不能反过来适配当前代码结构中的偶然形态。
