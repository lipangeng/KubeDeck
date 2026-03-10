# KubeDeck 规划总览

本文档是 KubeDeck 在产品方向、范围边界和交付优先级上的单一事实来源。

## 1. 产品定义

KubeDeck 是一个面向平台团队的、多集群、可插件扩展的 Kubernetes Web 控制平面。

它的目标不是优先去一比一替代 `kubectl`，而是提供：
- 一致的多集群操作上下文，
- 内置资源与 CRD 的统一控制界面，
- 以及一个可承载团队自定义工作流的插件化壳层。

## 2. 产品原则

1. 用户任务优先：UI 和 API 必须先服务真实操作流程，再暴露框架能力。
2. 多集群优先：cluster 上下文是用户的核心关注点，而不是附属筛选条件。
3. 统一资源模型：内置资源与 CRD 必须共享同一套心智模型和扩展路径。
4. 插件扩展是产品价值的一部分：插件应扩展真实工作流，而不仅仅是注册元数据。
5. 后端权威：鉴权与策略执行必须由后端负责。
6. 双语文档：产品文档与协作文档必须分别维护独立的中英文版本。

## 3. 目标用户

主要用户包括：
- 平台工程师，
- DevOps / SRE 运维人员，
- 以及需要跨多个 Kubernetes 集群交付业务的工程角色。

这些用户需要一个统一的 Web 控制平面，既能承载标准资源操作，也能承载团队特有流程。

## 4. 核心产品场景

KubeDeck 首先必须在以下场景中成立：
- 快速切换到正确的 cluster 上下文，
- 立即进入一个真实可用的资源域或工作流入口，
- 在一致上下文中浏览并操作 Kubernetes 资源，
- 并通过页面、面板、动作、菜单等方式扩展团队自定义流程。

## 5. MVP 闭环

第一个必须成立的用户闭环是：

1. 用户进入系统。
2. 用户确认或切换当前 cluster。
3. 用户进入一个具体资源域。
4. 用户看到真实的资源列表。
5. 用户执行一次标准操作，例如 create 或 apply。
6. 用户获得明确结果反馈。
7. 用户在同一 cluster 和 namespace 上下文中继续后续操作。

如果这条链路不能完整成立，产品仍然只是一个壳，而不是可用控制平面。

## 6. 首页职责

首页的首要职责应当是帮助用户：
- 确认当前操作上下文，
- 判断下一步最可能要做的任务，
- 并立即进入该工作流。

首页不应主要承担运行时诊断面板的职责。

## 7. 当前现实

截至当前基线：
- 后端 metadata 和 resource API 仍然是 stub，
- 前端导航结构已经存在，但尚未形成完整任务流，
- 插件合同与模板已经存在，
- 但产品尚未真正提供第一条端到端用户闭环。

因此当前最主要的问题不是视觉层面的不足，而是任务没有闭环。

AI 可以作为后续增强层在首条真实任务流成立后再引入，但不属于当前 MVP 范围。

## 8. 优先级顺序

近期优先级应按以下顺序推进：

1. 明确并固化产品需求，
2. 打通第一条用户任务闭环，
3. 让首页与导航围绕该任务闭环重组，
4. 加固 cluster 上下文与状态生命周期，
5. 然后再扩大插件驱动工作流。

## 9. 规范文档

- 产品需求澄清：`docs/product/product-requirements.zh.md`
- Product requirements clarification: `docs/product/product-requirements.md`
- 微内核扩展架构与 i18n 最低约束：`docs/product/architecture/microkernel-and-i18n-decision.zh.md`
- Microkernel extensibility and i18n minimum constraints: `docs/product/architecture/microkernel-and-i18n-decision.md`
- 内核贡献合同草案：`docs/product/architecture/kernel-contribution-contract.zh.md`
- Kernel contribution contract draft: `docs/product/architecture/kernel-contribution-contract.md`
- 内核实现映射草案：`docs/product/architecture/kernel-implementation-mapping.zh.md`
- Kernel implementation mapping draft: `docs/product/architecture/kernel-implementation-mapping.md`
- 集群感知菜单组合设计：`docs/product/architecture/cluster-aware-menu-composition.zh.md`
- Cluster-aware menu composition design: `docs/product/architecture/cluster-aware-menu-composition.md`
- i18n 运行时模型草案：`docs/product/architecture/i18n-runtime-model.zh.md`
- I18n runtime model draft: `docs/product/architecture/i18n-runtime-model.md`
- 首个基础任务流草案：`docs/product/workflows/first-workflow.zh.md`
- First core workflow draft: `docs/product/workflows/first-workflow.md`
- 共享工作上下文模型规范：`docs/product/workflows/shared-working-context-model.zh.md`
- Shared working context model spec: `docs/product/workflows/shared-working-context-model.md`
- 共享工作上下文状态结构草案：`docs/product/workflows/shared-working-context-state-schema.zh.md`
- Shared working context state schema draft: `docs/product/workflows/shared-working-context-state-schema.md`
- 共享工作上下文事件与更新规则：`docs/product/workflows/shared-working-context-events-and-update-rules.zh.md`
- Shared working context events and update rules: `docs/product/workflows/shared-working-context-events-and-update-rules.md`
- 共享工作上下文实现映射草案：`docs/product/workflows/shared-working-context-implementation-mapping.zh.md`
- Shared working context implementation mapping draft: `docs/product/workflows/shared-working-context-implementation-mapping.md`
- 共享工作上下文文件职责草案：`docs/product/workflows/shared-working-context-file-responsibility-draft.zh.md`
- Shared working context file responsibility draft: `docs/product/workflows/shared-working-context-file-responsibility-draft.md`
- 首页与 Workloads 页字段级信息架构草案：`docs/product/workflows/field-level-ia.zh.md`
- Homepage and Workloads field-level IA draft: `docs/product/workflows/field-level-ia.md`
- Namespace 上下文决策草案：`docs/product/workflows/namespace-context-decision.zh.md`
- Namespace context decision draft: `docs/product/workflows/namespace-context-decision.md`
- 首版实现边界清单：`docs/product/releases/v1/implementation-boundary.zh.md`
- V1 implementation boundary: `docs/product/releases/v1/implementation-boundary.md`
- V1 微内核与 i18n 实现边界：`docs/product/releases/v1/microkernel-and-i18n-implementation-boundary.zh.md`
- V1 microkernel and i18n implementation boundary: `docs/product/releases/v1/microkernel-and-i18n-implementation-boundary.md`
- V1 实施任务拆分：`docs/product/releases/v1/implementation-task-breakdown.zh.md`
- V1 implementation task breakdown: `docs/product/releases/v1/implementation-task-breakdown.md`
- V1 微内核与 i18n 任务拆分：`docs/product/releases/v1/microkernel-and-i18n-task-breakdown.zh.md`
- V1 microkernel and i18n task breakdown: `docs/product/releases/v1/microkernel-and-i18n-task-breakdown.md`
- V1 首批开发任务：`docs/product/releases/v1/initial-development-tasks.zh.md`
- V1 initial development tasks: `docs/product/releases/v1/initial-development-tasks.md`
- Task A 代码变更计划：`docs/product/releases/v1/task-a-code-change-plan.zh.md`
- Task A code change plan: `docs/product/releases/v1/task-a-code-change-plan.md`
- Task A 首个补丁范围：`docs/product/releases/v1/task-a-first-patch-scope.zh.md`
- Task A first patch scope: `docs/product/releases/v1/task-a-first-patch-scope.md`
- Task A 首个补丁检查清单：`docs/product/releases/v1/task-a-first-patch-checklist.zh.md`
- Task A first patch checklist: `docs/product/releases/v1/task-a-first-patch-checklist.md`
- V1 微内核与 i18n 首补丁范围：`docs/product/releases/v1/microkernel-and-i18n-first-patch-scope.zh.md`
- V1 microkernel and i18n first patch scope: `docs/product/releases/v1/microkernel-and-i18n-first-patch-scope.md`
- V1 微内核与 i18n 首补丁检查清单：`docs/product/releases/v1/microkernel-and-i18n-first-patch-checklist.zh.md`
- V1 microkernel and i18n first patch checklist: `docs/product/releases/v1/microkernel-and-i18n-first-patch-checklist.md`
- V1 首个微内核补丁代码落点草案：`docs/product/releases/v1/microkernel-first-patch-code-mapping.zh.md`
- V1 first microkernel patch code mapping draft: `docs/product/releases/v1/microkernel-first-patch-code-mapping.md`
- V1 首个微内核补丁任务清单与文件级职责：`docs/product/releases/v1/microkernel-first-patch-task-list-and-file-responsibility.zh.md`
- V1 first microkernel patch task list and file responsibility: `docs/product/releases/v1/microkernel-first-patch-task-list-and-file-responsibility.md`
- 架构设计：`docs/plans/2026-02-25-kubedeck-architecture-design.md`
- 实施计划：`docs/plans/2026-02-25-kubedeck-microkernel-baseline-implementation.md`
- PLAN (EN)：`docs/PLAN.md`

## 10. 执行规则

- 每个任务使用独立 feature branch 或 worktree。
- 每次合并后的仓库都必须保持可运行。
- PR 前需要验证受影响的前后端测试。
- 当需求或计划变化时，必须同步更新中英文文档。
