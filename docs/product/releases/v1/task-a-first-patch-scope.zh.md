# KubeDeck Task A 首个补丁范围

## 1. 目的

本文档定义 Task A：共享工作上下文模型 的最小首个补丁范围。

目标是在建立新所有权模型的同时，让第一批代码改动保持足够小、足够结构化、风险足够低。

## 2. 补丁范围原则

首个补丁应先建立新的共享工作上下文基础，而不是试图一次性完成工作流 UI。

这个补丁应做到：
- 创建新的状态结构
- 避免大范围页面重写
- 避免过早迁移旧代码
- 为后续集成型补丁铺路

## 3. 首个补丁应包含什么

首个补丁应只包含：

1. 新的共享工作上下文文件集合
2. 根共享状态类型
3. 最小事件 / 更新接口
4. 最小 selectors
5. 状态规则的基础单元测试

## 4. 首个补丁应创建的文件

创建以下文件：

- `frontend/src/state/work-context/types.ts`
- `frontend/src/state/work-context/clusterContext.ts`
- `frontend/src/state/work-context/namespaceContext.ts`
- `frontend/src/state/work-context/workflowContext.ts`
- `frontend/src/state/work-context/actionContext.ts`
- `frontend/src/state/work-context/listContext.ts`
- `frontend/src/state/work-context/events.ts`
- `frontend/src/state/work-context/selectors.ts`

并为新状态模块补充对应测试（在有意义时）：

- `frontend/src/state/work-context/clusterContext.test.ts`
- `frontend/src/state/work-context/namespaceContext.test.ts`
- `frontend/src/state/work-context/workflowContext.test.ts`
- `frontend/src/state/work-context/actionContext.test.ts`
- `frontend/src/state/work-context/listContext.test.ts`
- `frontend/src/state/work-context/events.test.ts`

并非每个文件都必须独立测试，只要状态规则的覆盖足够即可。

## 5. 首个补丁明确不包含什么

首个补丁不应包含：
- Homepage UI 重写
- Workloads 页面重写
- Create / Apply 操作界面实现
- 路由整体重构
- 插件集成改动
- 后端契约改动
- 视觉重设计
- AI 相关代码

## 6. 首个补丁中应保持不动的现有文件

除非只是最小类型导入接线，否则首个补丁不应改动以下类别：
- `frontend/src/App.tsx`
- page-shell 组件
- theme 文件
- plugin host 文件
- 现有 SDK transport/parsing 文件
- backend 文件

原因：
- 首个补丁的任务是建立新的所有权结构，而不是把迁移和接线全部缠在一起

## 7. 首个补丁允许的最小现有文件改动

首个补丁只允许对以下事项做最小改动：
- 必要时的 barrel 或 import 接线
- 若新状态模块测试需要，则做测试配置改动

这些改动必须保持狭窄，不能借机开始 UI 迁移。

## 8. 首个补丁的行为范围

首个补丁只需要证明以下行为：
- cluster 状态可以被干净表达
- namespace 浏览范围可以被干净表达
- workflow domain 可以被干净表达
- action execution target 在结构上与 namespace 浏览范围分离
- 事件 / 更新入口能保持模型约束

首个补丁不需要证明：
- 端到端 UI 连续性
- 真实界面导航
- 可见的任务入口行为

## 9. 首个补丁的测试范围

首个补丁应测试：
- cluster 切换状态迁移
- namespace 范围有效性规则
- action execution target 的结构规则
- 事件归属边界
- 事件处理过程中对无关状态的保留

首个补丁暂不测试：
- 页面渲染结果
- 视觉布局行为
- 浏览器导航流程

## 10. 首个补丁的破坏性改动规则

首个补丁此时不应清空旧代码。

虽然之后允许替换旧代码，但首个补丁应先专注于把新结构建出来。

首个补丁允许：
- 与旧代码并存

首个补丁不允许：
- 在替代结构尚未落地前，大规模删除历史页面或状态代码

## 11. 成功标准

当满足以下条件时，首个补丁就是成功的：
- 新的 `state/work-context` 结构已存在
- V1 核心上下文概念拥有统一结构归属
- 测试验证了核心状态规则
- 且没有把不必要的 UI 迁移或旧代码清理混进补丁

## 12. 首个补丁之后的下一补丁

首个补丁之后的下一个补丁应聚焦：
- 让 Homepage 与 Workloads 开始消费新的上下文基础

真正受控的旧所有权迁移应从第二个补丁开始，而不是在首个补丁中完成。
