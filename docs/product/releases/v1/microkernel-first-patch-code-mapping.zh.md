# KubeDeck V1 首个微内核补丁代码落点草案

## 1. 目的

本文档用于定义微内核架构工作的第一批代码落点计划。

它不实现代码，只是把已经接受的内核贡献合同与实现映射，进一步落到具体代码职责区域与补丁范围上。

## 2. 首补丁原则

首个微内核补丁应建立面向 contribution 的结构，而不是试图一次完成整个平台重写。

它应做到：
- 引入规范的内核合同类型，
- 引入 built-in contribution 注册路径，
- 引入壳层组合入口，
- 并阻止后续功能继续加深“壳层专属归属”。

它暂时不应做到：
- 完整动态插件加载，
- 全面重构后端 API，
- 或把所有现有页面一次性全部迁移为插件式模块。

## 3. 当前现实与应对方式

当前前端仍把过多工作流归属放在 app shell 中，而插件合同还很薄，主要停留在收集层。

当前后端插件层也仍然只有 identity，尚未形成 capability registration 或 execution-oriented contract。

因此首补丁不应试图“做完插件系统”，而应先建立后续补丁可依赖的架构基线。

## 4. 前端首补丁代码落点

### 4.1 新增内核合同文件

新增前端内核合同区域：

- `frontend/src/kernel/contracts/types.ts`
- `frontend/src/kernel/contracts/pageContribution.ts`
- `frontend/src/kernel/contracts/menuContribution.ts`
- `frontend/src/kernel/contracts/actionContribution.ts`
- `frontend/src/kernel/contracts/slotContribution.ts`

目的：
- 定义规范的 contribution contract 家族
- 停止继续把 transport DTO 与 kernel concern 混在 `sdk/types.ts` 中

### 4.2 内置贡献注册

新增内置能力注册区域：

- `frontend/src/kernel/builtins/registerBuiltInPages.ts`
- `frontend/src/kernel/builtins/registerBuiltInMenus.ts`
- `frontend/src/kernel/builtins/registerBuiltInActions.ts`
- `frontend/src/kernel/builtins/registerBuiltInSlots.ts`

目的：
- 把 `Homepage`、`Workloads`、`Create`、`Apply` 建模成 built-in contribution
- 在不要求马上引入第三方插件的前提下，先把产品推向 capability-style ownership

### 4.3 壳层组合入口

新增壳层组合区域：

- `frontend/src/kernel/runtime/kernelRegistry.ts`
- `frontend/src/kernel/runtime/composeKernelNavigation.ts`
- `frontend/src/kernel/runtime/resolveWorkflowActions.ts`
- `frontend/src/kernel/runtime/renderSlots.ts`

目的：
- 提供统一的壳层组合入口
- 把注册、渲染和工作流状态分离开

### 4.4 i18n 最低边界

新增最低文案访问边界：

- `frontend/src/i18n/copy.ts`
- `frontend/src/i18n/messages/en.ts`

如果团队希望立即显式引入 locale 类型，也可新增：

- `frontend/src/i18n/types.ts`

目的：
- 阻止新增用户文案继续以内联字符串方式散落
- 在不要求完整 i18n 落地的前提下，先建立未来可本地化访问路径

## 5. 前端首补丁中应尽量不动的现有文件

首补丁应避免大范围改写这些文件：

- `frontend/src/App.tsx`
- `frontend/src/pages/homepage/HomepageView.tsx`
- `frontend/src/pages/workloads/WorkloadsPage.tsx`
- `frontend/src/features/actions/ActionDrawer.tsx`
- `frontend/src/components/page-shell/*`

这些文件后续可以做最小接线，但不应成为内核合同设计的起点。

## 6. 前端可做窄改动的现有文件

首补丁可以对以下文件做窄范围调整：

- `frontend/src/core/pluginHost.ts`
  目的：
  让当前 host 形状向新合同家族对齐，或被新的 kernel registry 包起来

- `frontend/src/sdk/types.ts`
  目的：
  降低 transport DTO 与 kernel contribution type 的直接重叠

- `frontend/src/App.test.tsx`
  目的：
  如果最小组合接线需要测试适配，可以做窄调整

这些改动必须保持很窄，不能顺势开始 UI 大迁移。

## 7. 后端首补丁代码落点

### 7.1 后端能力合同文件

新增后端内核合同区域：

- `backend/pkg/sdk/capability.go`
- `backend/pkg/sdk/menu.go`
- `backend/pkg/sdk/page.go`
- `backend/pkg/sdk/action.go`

目的：
- 让能力注册不再停留在插件身份层
- 建立后端侧规范的合同家族

### 7.2 后端内核组合区域

新增后端组合区域：

- `backend/internal/plugins/capability_registry.go`
- `backend/internal/plugins/menu_composer.go`
- `backend/internal/plugins/page_composer.go`
- `backend/internal/plugins/action_composer.go`

目的：
- 让后端插件逻辑从“只存 ID”向 capability composition 演进
- 为 cluster-aware 的能力元数据组合打基线

### 7.3 后端内置能力注册

新增 built-in capability 注册区域：

- `backend/internal/core/builtins/workloads_capability.go`
- `backend/internal/core/builtins/homepage_capability.go`

如果团队更倾向把动作单独拆开，也可以新增：

- `backend/internal/core/builtins/create_action.go`
- `backend/internal/core/builtins/apply_action.go`

目的：
- 让内置能力成为一等内核输入

## 8. 后端首补丁中应尽量不动的现有文件

首补丁应避免大范围修改这些区域：

- `backend/internal/api/meta_handler.go`
- `backend/internal/api/resource_handler.go`
- `backend/internal/registry/*`
- `backend/internal/auth/*`

这些区域后续可以消费新的 capability composition，但首补丁不应把 capability contract 与完整 API 行为重设计混在一起。

## 9. 不能作为设计来源的现有文件

首补丁不能把这些文件当成架构真相：

- `frontend/src/App.tsx`
- `frontend/src/core/pluginHost.ts`
- `backend/internal/plugins/manager.go`
- `backend/pkg/sdk/plugin.go`

它们可以被适配、包裹或替换，但新的合同家族必须由当前接受的文档驱动，而不是为了保住这些现有形状。

## 10. 推荐实现顺序

首个微内核补丁建议按以下顺序推进：

1. 前端内核合同类型
2. 后端能力合同类型
3. 前端内置贡献注册
4. 后端内置能力注册
5. 前端壳层组合入口
6. 后端能力组合入口
7. i18n 最低文案访问边界
8. 如有必要，再对旧 host 或 shell 文件做窄兼容接线

## 11. 首补丁测试落点

首补丁应增加针对以下内容的窄测试：

- contribution 注册稳定性
- 组合顺序
- 内置贡献暴露
- 如果引入了 lookup helper，则验证最低文案访问行为

建议新增测试区域：

- `frontend/src/kernel/runtime/kernelRegistry.test.ts`
- `frontend/src/kernel/runtime/composeKernelNavigation.test.ts`
- `backend/internal/plugins/capability_registry_test.go`
- `backend/internal/plugins/menu_composer_test.go`

## 12. 首补丁成功标准

当以下条件成立时，首个微内核补丁算成功：

- 内置工作流域已经有了壳层之外可见的注册路径
- 前后端都具备规范的内核贡献合同
- 壳层已有一个清晰的组合入口
- 新增 UI 文案不再只能通过直接内联硬编码加入
- 补丁中没有混入完整插件平台或完整 i18n 铺开范围
