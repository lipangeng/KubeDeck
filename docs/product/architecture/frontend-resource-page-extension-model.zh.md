# 前端资源页面扩展模型

## 状态

当前有效的架构文档。

本文档定义 KubeDeck 的目标前端资源页面扩展模型。

它建立在现有微内核基线和集群感知菜单组合模型之上。

## 1. 问题定义

KubeDeck 需要一种资源页面模型，同时满足以下目标：

- CRD 资源始终有安全的默认页面与 YAML 编辑路径；
- `Pod` 等内置 Kubernetes 资源可以拥有更复杂、更专业的页面体验；
- 平台要支持未来页面扩展，而不要求所有资源都立刻走完全自定义页面；
- 系统既支持局部扩展、tab 替换，也支持后续整页接管。

因此，前端资源页面模型不能被简化成：

- “所有资源都共用一个通用页面”，或者
- “内置资源全部另起一套页面体系”。

更合理的目标是统一外壳下的分层扩展模型。

## 2. 设计目标

该模型必须满足以下目标：

1. 每个资源都能解析到一个默认可用页面；
2. YAML 始终是基础保底能力；
3. 内置资源与 CRD 共用同一套顶层资源页面模型；
4. 近期优先采用 tab-first 的扩展模式；
5. 少数资源类型在必要时允许整页接管；
6. 扩展模型在抽象层必须足够前置，以支持未来更多扩展点类型。

## 3. 非目标

本设计不打算：

- 在当前阶段实现块级扩展；
- 确定每种资源最终的视觉设计；
- 或一次性定义所有 Kubernetes 资源的完整 tab 清单。

当前目标是定义模型和优先级，而不是完成全部实现细节。

## 4. 推荐模型

KubeDeck 应采用：

- 统一的 `ResourcePageShell`
- 默认 tab 模型
- 分层扩展能力
- 少数资源的选择性整页接管

最终推荐方向为：

`ResourcePageShell + Default Tabs + Extension Capabilities + Final Resolution`

## 5. 核心对象

### 5.1 `ResourcePageShell`

这是所有资源页面共享的外壳。

它负责：

- 资源身份上下文
  - cluster
  - namespace
  - resource kind
  - resource name
- 共享页面布局
- 共享动作区
- 共享 tab 容器
- 默认 loading / empty / error 处理

外壳本身不应承载资源特定业务逻辑。

### 5.2 `ResourceCapability`

这一层描述某种资源类型允许暴露哪些页面能力。

例如：

- 是否支持 `overview`
- 是否支持 `yaml`
- 是否支持额外 runtime tabs
- 是否支持 tab 替换
- 是否支持整页接管

这一层描述的是“可做什么”，不是“页面怎么渲染”。

### 5.3 `ResourceExtensionCapability`

这是扩展类型的抽象层。

即使近期只实现其中一部分，也必须从一开始在模型中显性定义。

建议的扩展类型包括：

- `tab`
- `tab-replace`
- `page-takeover`
- `action`
- `slot`
- `section`

近期优先实现的只有：

- `tab`
- `tab-replace`
- `page-takeover`

其他类型先在模型层预留，后续再实现。

### 5.4 `ResourceExtension`

一个具体的资源页面扩展贡献。

它可以来自：

- 内置资源逻辑
- 插件扩展

它可以影响：

- tabs
- actions
- takeover 行为
- 以及未来的 slot 或 section 扩展

### 5.5 `ResourcePageResolution`

某个资源在某个上下文下最终解析出的页面结果。

它需要回答：

- 使用哪个 shell；
- 当前有哪些 tabs；
- 哪些 tabs 被替换；
- 哪些 actions 可用；
- 当前页面是 shell-based 还是 takeover 模式。

## 6. 默认 Tabs

已确认的基线模型是 tab-first。

页面外壳不应该只是“未来可能有 tabs 的容器”，而应该从一开始就将默认能力表达为 tabs。

每个资源页面至少应默认拥有以下 tabs：

- `Overview`
- `YAML`

这条规则同时适用于：

- 内置资源
- CRD 资源

## 7. YAML 保底规则

YAML 不是可选增强能力，而是资源页面模型的基础保底能力。

这意味着：

- 每种资源默认都必须有 YAML 路径；
- 对 CRD 来说，YAML 是最小可用编辑能力；
- 后续 YAML 能力的演进，应以可版本化 tab 的形式出现，而不是变成页面特例。

未来可能的 YAML 能力变体例如：

- `yaml.v1`
- `yaml.v2`
- `yaml.diff`
- `yaml.assisted`

不同资源未来可以启用不同 YAML 变体，但 YAML 路径本身必须始终存在。

## 8. Tab-First 扩展策略

近期资源页面扩展应优先围绕 tab 进行。

原因是：

- tab 是稳定而显性的扩展面；
- 可以显著减少近期对块级扩展的依赖；
- 可以比较干净地承载大多数早期页面增强需求。

未来可能出现的 tabs 包括：

- `Events`
- `Logs`
- `Metrics`
- `Runtime`
- `Related Resources`
- 插件自定义 tab

也就是说，近期资源页面的差异化优先通过以下方式表达：

- 增加 tabs
- 替换 tabs
- 禁用某些默认 tabs

## 9. 内置资源与 CRD 的关系

内置资源与 CRD 不应走两套完全无关的页面体系。

它们应共享同一套 shell 和扩展模型。

差异不在于“有没有模型”，而在于“扩展程度不同”。

### CRD

CRD 默认应至少拥有：

- `Overview`
- `YAML`

只有在显式贡献时才增加更多 tabs。

### 内置资源

内置资源可以增加更丰富的 tabs 和资源特有能力。

例如：

- `Pod` 后续可以增加 runtime 导向 tabs；
- 其他内置资源可以增加 status 或 topology tabs。

但除非明确需要 takeover，否则它们也应从同一套 shell-based 模型开始。

## 10. 整页接管规则

某些资源类型最终可能需要完全自定义页面。

这是允许的，但不应成为默认路径。

推荐规则：

- 默认先使用 shell + tabs；
- 只有在 shell 模型明确不足时才升级为整页接管。

这意味着当前确认的工作模式是：

- 默认走 shell-based 扩展；
- page takeover 是显式能力；
- 不能把 takeover 当成第一反应。

对 `Pod` 这类复杂资源，已确认的方向是：

- 模型上同时支持 shell 增强与整页接管；
- 但近期默认实现顺序应先从 shell 增强开始。

## 11. 块级扩展优先级

块级或 section 级扩展被认为是有价值的，但不属于近期必须项。

后置的扩展类型包括：

- summary block 注入
- inline section 替换
- side-panel 片段
- page body 子区块覆盖

这些能力应在模型层被预留，但不要求在当前阶段落地实现。

当前确认的优先级顺序是：

1. shell
2. default tabs
3. tab extension
4. tab replacement
5. page takeover
6. block-level extension later

这样可以在近期保持系统简单，同时保留长期扩展能力。

## 12. 对当前前端架构的影响

当前微内核基线已经支持：

- page contributions
- menu contributions
- action contributions
- slot contributions

但资源页面扩展还需要 page 层内部更细的模型。

系统需要从：

- “只有泛化页面贡献”

升级为：

- resource-page shell resolution
- tab capability resolution
- 支持 takeover 的资源页面组合模型

这不是单纯的 UI 增强，而是前端扩展模型的架构升级。

## 13. 与当前计划的一致性

项目仍然处在既定主线之内。

已经完成的微内核与插件扩展链路，是资源页面扩展模型成立的前提工作。

本文档所识别出的缺口，是当前下一层尚未显式设计的部分：

- 结构化的资源页面扩展模型

这说明项目没有跑偏，只是页面扩展层现在需要先于实现被正式定义出来。

## 14. 推荐的下一步规划

在进入实现前，应单独补一份实现规划，覆盖：

- `ResourcePageShell` contract
- tab capability schema
- tab registration 与 replacement 规则
- YAML capability versioning strategy
- page takeover decision rules

在这份规划被确认前，不应直接扩大资源页面实现范围。
