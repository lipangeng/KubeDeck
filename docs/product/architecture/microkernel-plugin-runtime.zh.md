# 微内核与插件运行时

## 1. 目的

本文档定义当前前后端运行时形状，后续所有开发都必须遵循它。

它将此前分散的微内核、合同、映射与最低 i18n 约束草稿收束为一份当前有效的架构主文档。

## 2. 架构定位

微内核不是后续增强项。

它是保护 KubeDeck 不退回“硬编码壳层 + 内置特例页面”模式的运行时边界。

产品当前依赖于：

- 贡献注册，
- 运行时组合，
- 后端能力权威，
- 可扩展的 UI 表面。

## 3. 运行时层次

KubeDeck 使用四层运行时结构。

### 3.1 Kernel Core

Kernel Core 负责：

- contribution contracts，
- composition rules，
- route 与 menu 的组装，
- context boundaries，
- execution boundaries。

前端与后端都必须存在这一层。

### 3.2 Built-In Capabilities

像 `Homepage`、`Workloads`、资源页和核心 actions 这样的内置产品域，不应继续作为永久性的壳层特例存在。

它们是仓库内置能力，但必须通过 kernel contracts 接入。

### 3.3 Plugin Capabilities

插件通过与 built-in 相同的运行时家族贡献能力。

当前主要贡献族为：

- `page`
- `menu`
- `action`
- `slot`

同时，资源页扩展现在还依赖以下扩展能力类型：

- `tab`
- `tab-replace`
- `page-takeover`

`section`、更细粒度 `slot` 等类型当前只保留模型，不进入近期实现。

### 3.4 Runtime Resolution

最终 UI 应消费解析后的结果，而不是直接消费原始注册项。

这条原则适用于：

- 菜单，
- 页面，
- 动作，
- 资源页 tabs，
- 各类扩展表面。

## 4. 后端职责

后端对以下内容拥有权威：

- capability registration，
- plugin identity 与 manifest loading，
- capability composition，
- resource 与 workflow execution entry points，
- 基于权限与上下文的 availability 判断。

如果产品能力依赖 capability truth，后端就不能退化成单纯 metadata stub。

## 5. 前端职责

前端负责：

- 消费解析后的 capability metadata，
- 基于 kernel 输入组合 UI runtime，
- 渲染组合后的菜单与页面，
- 承载产品信息架构，
- 在 cluster、menu 与资源导航之间保持工作上下文连续性。

前端不能绕过 kernel 再建一套不相关的页面体系。

## 6. 内置能力规则

内置功能必须被建模为一等 kernel contribution。

这意味着：

- 内置页面通过 kernel contracts 注册，
- 内置菜单参与 menu composition，
- 内置 actions 通过 action contracts 挂接，
- 内置资源体验也必须遵循将来插件可用的同一套页面扩展模型。

## 7. 菜单运行时规则

导航不是页面注册结果的直接渲染。

它必须按以下顺序解析：

1. default blueprint，
2. mountable capabilities，
3. cluster-aware availability，
4. user overrides，
5. final composed menu result。

这已经是当前运行时形状的一部分。

## 8. 资源页运行时规则

资源页不是随意堆叠的独立页面。

它必须按以下顺序解析：

1. 共享 `ResourcePageShell`，
2. 默认 tabs，
3. extension capabilities，
4. 可选 tab replacement，
5. 对少数特殊资源类型允许 page takeover。

这条规则同时适用于内置资源与 CRD。

## 9. i18n 最低运行时规则

完整 i18n 交付仍然可以后置，但运行时必须保持 locale boundary。

当前最低规则为：

- 新增用户文案必须通过 i18n access layer，
- capability labels 必须可本地化，
- locale 必须被视为产品级状态，
- 文案不能与控制逻辑深度耦合。

## 10. 后置项

以下方向是后续能力，不是当前运行时必需项：

- 第三方插件的动态热装载与热卸载，
- 插件市场或插件管理 UI，
- 对所有页面区域的完整 slot replacement，
- 完整的运行时语言切换与持久化，
- 广泛的块级页面扩展能力。

## 11. 架构检查

在继续增加大功能前，必须持续回答“是”的问题包括：

- 新功能是否能通过当前 kernel contracts 接入？
- 它是否保持组合式菜单，而不是硬编码导航？
- 它是否保持统一资源页外壳？
- 它是否让 built-in 与 plugin 继续处在同一运行时家族内？
- 它是否避免继续加深不可本地化的 UI 文案？

只要其中任一问题答案为“否”，就应先回到架构调整，而不是继续扩功能。
