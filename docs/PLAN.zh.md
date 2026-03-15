# KubeDeck PLAN

本文档是当前产品方向、架构重点与开发入口的总索引。

## 1. 当前产品定义

KubeDeck 是一个面向平台团队的、多集群、可插件扩展的 Kubernetes Web 控制面。

当前产品方向由四个核心点定义：

1. 多集群操作上下文是一等公民，
2. 内置资源与 CRD 共享统一资源模型，
3. 前后端必须遵循微内核扩展架构，
4. 导航与资源页面必须通过组合生成，而不是硬编码堆叠。

KubeDeck 的目标不是再做一个通用 Kubernetes 仪表盘。

## 2. 当前开发阶段

项目当前处于“开发模式准备完成”的阶段。

这意味着：

- 产品与架构目标已经重新澄清，
- 历史规划草稿已经从主阅读路径移除，
- 后续实现必须遵循当前架构文档，而不是旧的壳层优先假设。

## 3. 当前优先级

开发优先级顺序为：

1. 保持微内核与插件运行时形状稳定，
2. 实现 cluster-aware 菜单组合系统，
3. 实现资源页外壳与 tab-first 扩展模型，
4. 在这些基础上恢复第一条真实用户工作流，
5. 在首条工作流稳定后再扩展插件能力。

## 4. 当前非优先项

以下内容当前明确不是优先项：

- AI 辅助工作流交付，
- 通用聊天入口，
- 完整 i18n 产品能力，
- 大型仪表盘式可观测界面，
- 在工作流未完成前的大量视觉打磨。

## 5. 开发入口

进入开发模式时，请先阅读：

- 开发模式说明：`docs/product/development-mode.zh.md`
- Development mode guide: `docs/product/development-mode.md`

## 6. 当前核心文档

### 产品

- 产品需求：`docs/product/product-requirements.zh.md`
- Product requirements: `docs/product/product-requirements.md`

### 架构

- 微内核与插件运行时：`docs/product/architecture/microkernel-plugin-runtime.zh.md`
- Microkernel and plugin runtime: `docs/product/architecture/microkernel-plugin-runtime.md`
- Cluster-aware 菜单组合：`docs/product/architecture/cluster-aware-menu-composition.zh.md`
- Cluster-aware menu composition: `docs/product/architecture/cluster-aware-menu-composition.md`
- 前端资源页扩展模型：`docs/product/architecture/frontend-resource-page-extension-model.zh.md`
- Frontend resource page extension model: `docs/product/architecture/frontend-resource-page-extension-model.md`

### V1 交付

- V1 实现边界：`docs/product/releases/v1/implementation-boundary.zh.md`
- V1 implementation boundary: `docs/product/releases/v1/implementation-boundary.md`
- V1 开发计划：`docs/product/releases/v1/v1-development-plan.zh.md`
- V1 development plan: `docs/product/releases/v1/v1-development-plan.md`
- Scoped menu settings implementation plan：`docs/product/releases/v1/scoped-menu-settings-implementation-plan.zh.md`
- Scoped menu settings implementation plan (EN): `docs/product/releases/v1/scoped-menu-settings-implementation-plan.md`
- 基础架构补全实施计划：`docs/product/releases/v1/foundation-architecture-remediation-plan.zh.md`
- Foundation architecture remediation plan: `docs/product/releases/v1/foundation-architecture-remediation-plan.md`

### 归档

- 归档索引：`docs/archive/README.zh.md`
- Archive index: `docs/archive/README.md`

## 7. 推荐阅读顺序

进入当前开发模式时，推荐按以下顺序阅读：

1. `docs/PLAN.zh.md`
2. `docs/product/product-requirements.zh.md`
3. `docs/product/development-mode.zh.md`
4. `docs/product/architecture/microkernel-plugin-runtime.zh.md`
5. `docs/product/architecture/cluster-aware-menu-composition.zh.md`
6. `docs/product/architecture/frontend-resource-page-extension-model.zh.md`
7. `docs/product/releases/v1/implementation-boundary.zh.md`
8. `docs/product/releases/v1/v1-development-plan.zh.md`

## 8. 文档生效规则

只有本 PLAN 中列出的文档属于当前开发主线。

`docs/archive/` 中保留的是历史规划与中间推导文档，仅供参考，不再作为当前实现决策来源。
