# KubeDeck 开发模式

## 1. 目的

本文档定义如何基于当前产品与架构文档进入实际开发。

它替代了之前那种“大量草稿、补丁范围、探索记录同时处于主路径”的工作方式。

## 2. 当前工作规则

后续开发必须从当前有效的架构集合出发，而不是从旧的壳层优先代码路径出发，也不是从归档草稿出发。

当前主流程是：

1. 先阅读当前核心文档，
2. 基于微内核运行时进行实现，
3. 所有 UI 变化都对齐菜单组合与资源页扩展规则，
4. 在 V1 边界内分批交付，
5. 不让旧草稿重新回到实现决策主路径。

## 3. 必读文档

开始功能开发前，按以下顺序阅读：

1. `docs/PLAN.zh.md`
2. `docs/product/product-requirements.zh.md`
3. `docs/product/architecture/microkernel-plugin-runtime.zh.md`
4. `docs/product/architecture/cluster-aware-menu-composition.zh.md`
5. `docs/product/architecture/frontend-resource-page-extension-model.zh.md`
6. `docs/product/releases/v1/implementation-boundary.zh.md`
7. `docs/product/releases/v1/v1-development-plan.zh.md`

## 4. 开发护栏

所有新增开发都必须遵循以下规则：

### 4.1 不再重建旧壳层

不要通过在微内核之外继续堆硬编码页面、菜单和工作流分支来扩展应用。

### 4.2 菜单必须通过组合生成

不要把左侧导航继续当成“注册了什么就渲染什么”的列表。

菜单开发必须符合 blueprint、mount、availability、override 这套组合模型。

### 4.3 资源页必须遵循统一模型

不要再为内置资源和 CRD 分裂出两套不相关的页面体系。

资源页面必须遵循共享 shell 与 tab-first 扩展模型，除非有明确理由升级到 takeover。

### 4.4 V1 必须保持收敛

在首条工作流稳定之前，不要扩展到 AI、大型可观测性、插件市场或泛化定制能力。

### 4.5 i18n 需要保持可演进

完整 i18n 交付可以后置，但新增面向用户的文案必须通过 locale boundary 管理，而不是继续无边界硬编码。

## 5. 当前开发顺序

当前实施顺序为：

1. 稳定微内核与插件运行时，
2. 实现 cluster-aware 菜单组合，
3. 实现资源页外壳与 tab 能力，
4. 在这两层之上接回首条真实工作流，
5. 在 V1 成立后再扩展插件能力与更复杂的资源特化。

## 6. 当前有效事实来源

当前实现事实只来自：

- `PLAN`
- 当前产品需求文档
- 当前架构文档
- 当前 V1 边界与开发计划

归档文档不定义当前事实。

## 7. 归档文档的用途

归档文档仍可用于：

- 查看历史推导过程，
- 回溯旧方案对比，
- 理解旧术语，
- 或辅助迁移。

但除非其内容被重新吸收进当前核心文档，否则不能再作为新的实现决策来源。
