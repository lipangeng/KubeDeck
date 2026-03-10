# KubeDeck V1 开发计划

## 1. 目的

本文档是当前 V1 的执行计划。

它用于替代此前分散的 task breakdown、first patch、checklist 和 code mapping 草稿，作为一份真正面向开发的计划文档。

## 2. 计划原则

V1 开发必须遵循“架构优先”的顺序：

1. 运行时基础，
2. 菜单组合，
3. 资源页外壳，
4. 首条真实工作流，
5. 受控特化。

不要反向操作，不要在当前架构之外重新堆页面工作流。

## 3. 阶段 1：稳定运行时基础

目标：

- 将微内核与插件运行时稳定为所有后续工作的共同基座。

交付物：

- 干净的前后端 kernel contracts，
- built-in 与 plugin capability 通过同一运行时家族注册，
- 符合 capability composition 的 metadata 与 execution 入口，
- 新增 UI 文案具备最低 i18n boundary。

退出条件：

- built-in 功能不再依赖 shell-only 假设，
- plugin capability 流保持完整，
- 新增 UI 文案能通过 i18n boundary 进入系统。

## 4. 阶段 2：实现菜单组合系统

目标：

- 将左侧导航从原始注册列表升级为组合结果。

交付物：

- 默认菜单 blueprint，
- 面向 built-in workflows、CRD、plugin route 的统一 mount 模型，
- cluster-aware availability 解析，
- 对“已配置但不可用”入口的禁用态，
- `CRDs` 兜底入口，
- 为未来 global / cluster-local overrides 预留空间。

退出条件：

- 实际渲染的导航来自组合结果，
- 显式配置的 CRD 与 plugin mounts 落在同一运行时家族内，
- 不可用但已配置的入口依然稳定可见。

## 5. 阶段 3：实现资源页外壳

目标：

- 为内置资源与 CRD 建立统一资源页模型。

交付物：

- `ResourcePageShell`，
- 默认 `Overview` 与 `YAML` tabs，
- 资源身份与上下文的基础处理，
- `tab`、`tab-replace`、`page-takeover` 三类扩展能力，
- 通过该模型交付首批资源页面。

退出条件：

- 第一批资源都通过 shell 解析，
- YAML 成为真实的基础 tab，
- 模型允许后续特化而不破坏共享 shell。

## 6. 阶段 4：接回首条真实工作流

目标：

- 在新的菜单与资源页系统之上，交付首条可用工作流。

交付物：

- 以任务入口为中心的 Homepage，
- 通过组合式菜单进入 `Workloads`，
- 真实 workload 列表，
- 一条真实动作路径，例如 create、apply 或 edit，
- 清晰结果反馈，
- 跨 cluster、namespace 与上一工作路径的连续性。

退出条件：

- 用户能在当前运行时形状之上完整走通首条工作流，
- 关键路径不再依赖已废弃的旧壳层模式。

## 7. 阶段 5：受控特化

目标：

- 证明架构能够支持更丰富的能力特化，而不会重新退化成大量特例。

交付物：

- 至少一个 built-in 资源的特化 tabs，
- 至少一个通过共享模型交付的 CRD 路径，
- 至少一个落在同一体系中的 plugin 入口或扩展。

退出条件：

- 所有特化都通过批准的扩展能力类型完成，
- 没有引入新的平行页面模型。

## 8. 明确后置内容

以下内容不在本计划范围内：

- AI 工作流支持，
- 通用聊天表面，
- 插件市场流程，
- 完整菜单自定义产品化，
- 广泛的块级资源页扩展，
- 在首条工作流稳定前扩展大量资源域。

## 9. 验证规则

每个阶段收尾时都应完成：

- 受影响的前端测试通过，
- 受影响的后端测试通过，
- 构建验证通过，
- 若架构或边界发生变化则同步更新文档。

## 10. 可直接进入实现的检查

只有在以下条件仍然成立时，才能直接按本计划推进：

- 当前实现仍然以 `microkernel-plugin-runtime` 为目标，
- 菜单开发仍然以 `cluster-aware-menu-composition` 为目标，
- 资源页开发仍然以 `frontend-resource-page-extension-model` 为目标，
- V1 范围仍然与 `implementation-boundary` 保持一致。

如果接下来要先完成基础架构补全，再进入更广泛的 V1 实施，请使用：

- `docs/product/releases/v1/foundation-architecture-remediation-plan.zh.md`
