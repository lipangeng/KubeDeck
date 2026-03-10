# KubeDeck V1 首个微内核补丁任务清单与文件级职责

## 1. 目的

本文档把首个微内核补丁代码落点草案进一步压缩成可执行的开发前清单。

它定义两件事：
- 首补丁任务清单
- 每个计划文件的文件级职责边界

## 2. 首补丁任务清单

首个微内核补丁建议按以下顺序执行。

### Task 1：定义前端内核合同类型

目标：
- 建立规范的前端 contribution contract 家族

文件：
- `frontend/src/kernel/contracts/types.ts`
- `frontend/src/kernel/contracts/pageContribution.ts`
- `frontend/src/kernel/contracts/menuContribution.ts`
- `frontend/src/kernel/contracts/actionContribution.ts`
- `frontend/src/kernel/contracts/slotContribution.ts`

完成标准：
- page/menu/action/slot 贡献结构已经显式化
- transport DTO 不再是内核合同设计默认落点

### Task 2：定义后端能力合同类型

目标：
- 建立规范的后端 capability contract 家族

文件：
- `backend/pkg/sdk/capability.go`
- `backend/pkg/sdk/menu.go`
- `backend/pkg/sdk/page.go`
- `backend/pkg/sdk/action.go`

完成标准：
- 后端能力注册不再只停留在 identity 层
- capability metadata 类型已经独立于 API handler 存在

### Task 3：注册前端内置贡献

目标：
- 给内置工作流域建立 contribution-style 的注册路径

文件：
- `frontend/src/kernel/builtins/registerBuiltInPages.ts`
- `frontend/src/kernel/builtins/registerBuiltInMenus.ts`
- `frontend/src/kernel/builtins/registerBuiltInActions.ts`
- `frontend/src/kernel/builtins/registerBuiltInSlots.ts`

完成标准：
- `Homepage`、`Workloads`、`Create`、`Apply` 都可以被描述为 built-in contribution

### Task 4：注册后端内置能力

目标：
- 让内置能力元数据在后端侧显式化

文件：
- `backend/internal/core/builtins/homepage_capability.go`
- `backend/internal/core/builtins/workloads_capability.go`
- 可选：
  - `backend/internal/core/builtins/create_action.go`
  - `backend/internal/core/builtins/apply_action.go`

完成标准：
- 后端内置能力已经成为显式的内核输入

### Task 5：新增前端内核运行时组合入口

目标：
- 为壳层提供一个统一的 kernel-owned composition entry

文件：
- `frontend/src/kernel/runtime/kernelRegistry.ts`
- `frontend/src/kernel/runtime/composeKernelNavigation.ts`
- `frontend/src/kernel/runtime/resolveWorkflowActions.ts`
- `frontend/src/kernel/runtime/renderSlots.ts`

完成标准：
- 壳层已经有一个清晰位置可以消费内置与未来插件贡献

### Task 6：新增后端能力组合入口

目标：
- 让后端插件逻辑从 ID 存储演进到 capability composition

文件：
- `backend/internal/plugins/capability_registry.go`
- `backend/internal/plugins/menu_composer.go`
- `backend/internal/plugins/page_composer.go`
- `backend/internal/plugins/action_composer.go`

完成标准：
- 后端 capability metadata 已可独立于 API handler 组合

### Task 7：新增最低 i18n 文案边界

目标：
- 阻止新增产品 UI 文案继续放大内联硬编码扩散

文件：
- `frontend/src/i18n/copy.ts`
- `frontend/src/i18n/messages/en.ts`
- 可选：
  - `frontend/src/i18n/types.ts`

完成标准：
- 新增 UI 文案已有统一访问路径

### Task 8：新增窄兼容接线

目标：
- 让新内核基线可以在不进行全面迁移的前提下与当前代码共存

可做窄改动的文件：
- `frontend/src/core/pluginHost.ts`
- `frontend/src/sdk/types.ts`
- `frontend/src/App.test.tsx`
- `backend/internal/plugins/manager.go`
- `backend/pkg/sdk/plugin.go`

完成标准：
- 新增内核结构已经可以编译或被验证，而不需要立即引发全面 UI/API 迁移

## 3. 文件级职责

### 前端内核合同文件

`frontend/src/kernel/contracts/types.ts`
- 负责共享合同辅助类型
- 不负责产品页面细节

`frontend/src/kernel/contracts/pageContribution.ts`
- 负责页面贡献结构
- 不负责页面渲染逻辑

`frontend/src/kernel/contracts/menuContribution.ts`
- 负责菜单贡献结构
- 不负责菜单分组实现策略

`frontend/src/kernel/contracts/actionContribution.ts`
- 负责动作贡献结构
- 不负责动作表单行为

`frontend/src/kernel/contracts/slotContribution.ts`
- 负责槽位贡献结构
- 不负责槽位 UI 内容

### 前端内置注册文件

`frontend/src/kernel/builtins/registerBuiltInPages.ts`
- 负责注册内置页面贡献
- 不负责渲染页面

`frontend/src/kernel/builtins/registerBuiltInMenus.ts`
- 负责注册内置菜单贡献
- 不负责组合最终导航

`frontend/src/kernel/builtins/registerBuiltInActions.ts`
- 负责注册内置工作流动作
- 不负责执行动作

`frontend/src/kernel/builtins/registerBuiltInSlots.ts`
- 负责注册内置槽位贡献
- 不负责直接渲染槽位内容

### 前端运行时文件

`frontend/src/kernel/runtime/kernelRegistry.ts`
- 负责贡献聚合与查找
- 不负责页面局部状态

`frontend/src/kernel/runtime/composeKernelNavigation.ts`
- 负责根据已注册贡献组合导航
- 不负责侧栏渲染

`frontend/src/kernel/runtime/resolveWorkflowActions.ts`
- 负责按工作流域查找动作
- 不负责动作执行结果

`frontend/src/kernel/runtime/renderSlots.ts`
- 负责槽位贡献解析规则
- 不负责槽位内容实现

### 前端 i18n 文件

`frontend/src/i18n/copy.ts`
- 负责最低文案查找/访问方式
- 不负责组件行为

`frontend/src/i18n/messages/en.ts`
- 负责初始消息目录
- 不负责除目录结构之外的翻译逻辑

`frontend/src/i18n/types.ts`
- 如果引入，则负责 locale 相关类型定义
- 不负责偏好持久化

### 后端能力合同文件

`backend/pkg/sdk/capability.go`
- 负责后端能力注册结构
- 不负责 HTTP transport

`backend/pkg/sdk/menu.go`
- 负责后端菜单能力元数据结构
- 不负责菜单可见性执行

`backend/pkg/sdk/page.go`
- 负责后端页面能力元数据结构
- 不负责页面渲染

`backend/pkg/sdk/action.go`
- 负责后端动作能力元数据与执行合同结构
- 不负责业务动作实现

### 后端组合文件

`backend/internal/plugins/capability_registry.go`
- 负责能力聚合与查找
- 不负责 API 输出渲染

`backend/internal/plugins/menu_composer.go`
- 负责后端菜单组合
- 不负责 HTTP handler

`backend/internal/plugins/page_composer.go`
- 负责后端页面元数据组合
- 不负责前端 routes

`backend/internal/plugins/action_composer.go`
- 负责后端动作元数据组合
- 不负责动作执行内部实现

### 后端内置能力文件

`backend/internal/core/builtins/homepage_capability.go`
- 负责 homepage capability registration
- 不负责 homepage UI

`backend/internal/core/builtins/workloads_capability.go`
- 负责 workloads capability registration
- 不负责 workload API 响应实现

`backend/internal/core/builtins/create_action.go`
- 如果引入，则负责 built-in create action capability registration
- 不负责通用内核组合逻辑

`backend/internal/core/builtins/apply_action.go`
- 如果引入，则负责 built-in apply action capability registration
- 不负责通用内核组合逻辑

## 4. 首补丁中应保持在范围外的现有文件

以下文件不应成为首个微内核补丁的实现中心：

- `frontend/src/App.tsx`
- `frontend/src/pages/homepage/HomepageView.tsx`
- `frontend/src/pages/workloads/WorkloadsPage.tsx`
- `frontend/src/features/actions/ActionDrawer.tsx`
- `backend/internal/api/meta_handler.go`
- `backend/internal/api/resource_handler.go`

这些文件后续可以消费新的结构，但不应由它们来定义结构。

## 5. 首补丁完成检查

当以下条件成立时，首补丁才算进入可实施状态：
- 任务顺序已经接受
- 每个文件都有清晰且单一的职责描述
- 现有页面与 API 文件已被明确排除出首补丁中心
- 补丁仍然避免完整插件平台与完整 i18n 铺开范围
