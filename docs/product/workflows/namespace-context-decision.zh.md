# KubeDeck Namespace 上下文决策草案

## 1. 目的

本文档定义 KubeDeck 中 namespace 作为跨页面工作上下文时，应如何工作。

namespace 不是单纯的页面筛选条件，它会影响：
- 资源可见范围，
- 操作默认落点，
- 页面间连续性，
- 以及工作流安全性。

## 2. 决策摘要

namespace 应被视为当前 active cluster 之下的共享工作上下文。

KubeDeck 应支持以下 namespace 范围：
- 单个 namespace，
- 多个 namespace，
- 全部 namespace。

但像 `Create`、`Apply` 这样的操作型流程，在提交前必须收敛为明确的执行目标。

## 3. Namespace 模型

### 3.1 上下文层级

推荐的上下文层级为：

1. active cluster
2. namespace scope
3. 当前资源域或页面

这意味着 namespace 属于全局工作上下文，而不是只属于单个页面。

### 3.2 支持的 Namespace 范围类型

KubeDeck 应支持：
- `single`
- `multiple`
- `all`

定义如下：
- `single`：当前仅有一个 namespace 生效
- `multiple`：当前有一组已定义 namespace 生效
- `all`：当前 cluster 下所有可见 namespace 都在范围内

## 4. 页面类型规则

### 4.1 列表与浏览类页面

像 `Workloads` 这样的列表页应支持：
- 单个 namespace，
- 多个 namespace，
- 全部 namespace。

原因是：
- 用户在决定具体操作前，往往需要先跨更大范围浏览，
- 这也更符合真实的 Kubernetes 使用方式。

### 4.2 详情页

详情页应继承来源页面的 namespace 上下文。

如果对象是 namespaced 资源：
- 必须明确展示对象实际所在的 namespace。

如果对象是 cluster-scoped 资源：
- namespace 上下文可以保留为继承来的浏览上下文，
- 但对象本身必须标明为 cluster-scoped。

### 4.3 Create / Apply 页面或操作界面

Create / Apply 流程不能在 namespace 范围模糊的情况下提交。

在最终提交前，执行目标必须收敛为以下之一：
- 单个明确的 namespace，
- 或 payload 中对象自行定义的 namespace，
- 或 cluster-scoped 资源，无需 namespace 目标。

`multiple` 和 `all` 可以作为 create/apply 启动前的用户上下文，但它们本身不能直接作为最终写入目标。

## 5. 默认规则

### 5.1 初始默认值

当用户第一次进入某个 cluster 时：
- 如果存在该 cluster 的上次 namespace 范围记录，则优先使用，
- 否则默认使用 `single: default`。

### 5.2 Cluster 切换默认值

当用户切换 cluster 时：
- 优先恢复目标 cluster 的上次 namespace 范围，
- 若不存在，则回退到 `single: default`。

KubeDeck 不应在未验证的情况下，把 namespace 名称直接从一个 cluster 生搬到另一个 cluster。

### 5.3 无上下文场景

如果完全没有历史 namespace 信息：
- 对操作型入口，应使用 `single: default` 作为安全基线，
- 同时仍允许用户在列表型页面中扩大浏览范围。

## 6. 继承规则

### 6.1 首页到 Workloads

首页只应展示最小必要的 namespace 上下文摘要。

当用户进入 `Workloads` 时：
- 继承当前共享工作上下文中的 namespace 范围。

### 6.2 Workloads 到详情页

当用户从 `Workloads` 打开某个资源时：
- 继承当前 namespace 范围，
- 同时明确展示对象自身的实际 namespace。

### 6.3 Workloads 到 Create / Apply

当用户从 `Workloads` 发起 `Create` 或 `Apply` 时：
- 将当前 namespace 范围作为上下文传入操作界面。

随后按以下方式收敛：
- 如果当前范围是 `single`，则将其作为默认目标 namespace，
- 如果当前范围是 `multiple` 或 `all`，则提交前必须显式确认目标 namespace，或依赖对象自身定义的 namespace。

### 6.4 返回流程

当 create/apply 完成后：
- 应返回到来源页面，并保留原始 namespace 浏览范围。

这意味着执行目标 namespace 与返回后的浏览范围，可能是两个不同值。

## 7. Create / Apply 收敛规则

### 7.1 单个 Namespace 范围

如果当前范围是 `single`：
- 默认使用该 namespace 作为执行目标，
- 除非 payload 中显式定义了另一个有效 namespace。

### 7.2 多个 Namespace 范围

如果当前范围是 `multiple`：
- 不能静默自动选择其中一个 namespace，
- 必须要求用户确认单个目标 namespace，
- 或要求 payload 对每个对象自行定义。

### 7.3 全部 Namespace 范围

如果当前范围是 `all`：
- 不能把 `all` 当作合法写入目标，
- 提交前必须收敛为明确目标 namespace，
- 除非该对象本身是 cluster-scoped。

### 7.4 多文档 Apply

对于多文档 apply：
- 每个文档都必须收敛到具体目标，
- 结果反馈中应明确标出每个文档最终使用了哪个 namespace。

如果部分文档是 cluster-scoped，而部分是 namespaced：
- 则在校验和结果反馈中都必须清晰可见。

## 8. 校验规则

提交前，KubeDeck 应校验：
- 当必须要求明确 namespace 时，最终目标是否已经具体化，
- 目标 namespace 在当前 cluster 中是否合法，
- 是否有文档与当前 namespace 收敛规则冲突，
- 以及资源本身是 cluster-scoped 还是 namespaced。

校验不能把模糊的 namespace 范围静默改写为非预期目标。

## 9. UI 规则

### 9.1 浏览类 UI

列表页应允许对 namespace 范围进行控制，支持：
- 单个，
- 多个，
- 全部。

UI 必须清楚区分：
- 当前浏览范围，
- 以及对象实际所在的 namespace（如相关）。

### 9.2 操作类 UI

Create / Apply UI 必须展示：
- 当前 cluster，
- 当前浏览范围，
- 已收敛出的执行目标 namespace，
- 以及提交前仍未解决的 namespace 问题。

### 9.3 反馈类 UI

结果反馈应展示：
- 用户当时的 namespace 浏览范围，
- 每次操作最终使用的具体 namespace，
- 以及哪些对象是 cluster-scoped。

## 10. 安全规则

- 不能把 `all namespaces` 当作写入目标。
- 不能把 `multiple` 静默收敛成某一个 namespace。
- 不能在页面跳转时丢失 namespace 上下文。
- 不能在 cluster 切换后未经验证就保留 namespace，必须验证或安全回退。

## 11. 推荐的首版实现

对首版实现，建议：
- 浏览支持 `single` 和 `all`，
- `multiple` 先在决策模型中保留，但若有必要可推迟到后续 UI 版本，
- create/apply 在提交前必须收敛为单个 namespace 或 cluster-scoped 对象，
- 并且 namespace 上下文必须在 Homepage、Workloads 与 Create / Apply 之间保持连续。

这样既能保证模型正确，又不会把完整多选 UI 复杂度一次性压进首版交付。
