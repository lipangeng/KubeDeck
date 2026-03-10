# KubeDeck 基础架构补全实施计划

> **For agentic workers:** REQUIRED: 实施前必须使用 `test-driven-development`，并且所有变更都要遵循 `docs/product/development-mode.zh.md`。步骤统一使用 checkbox（`- [ ]`）跟踪。

**Goal:** 补齐当前“微内核骨架”与“产品级运行时模型”之间的结构缺口，让 V1 能基于正确基础继续推进。

**Architecture:** 本计划从已经可工作的 kernel contracts 与 plugin discovery 出发，按顺序补上缺失的产品运行时层：组合式菜单、共享工作上下文、共享资源页外壳，最后再把首条真实工作流接回这些系统之上。所有任务都必须坚持架构优先，禁止重新堆回 shell-only 路径。

**Tech Stack:** Go backend, React + TypeScript frontend, Vite, MUI, Vitest, Go test

---

## 1. 范围

本计划只覆盖进入直接开发前必须完成的基础架构补全工作。

不包含：

- AI 功能，
- 大范围 UI 打磨，
- 插件市场或插件管理，
- 广泛块级页面扩展，
- 大范围资源域扩展。

## 2. 当前缺口摘要

当前代码已经具备：

- 前后端 kernel contracts，
- plugin discovery，
- backend kernel snapshot 输出，
- frontend runtime hydrate，
- 最低限度的 i18n copy 访问。

但仍缺少以下产品级运行时层：

1. 组合式菜单运行时，
2. 共享工作上下文运行时，
3. 资源页外壳运行时，
4. 基于这些系统重建的首条真实工作流。

## 3. 文件结构

### Backend 区域

- Modify: `backend/pkg/sdk/menu.go`
  - 扩展 menu descriptor，使其能表示由 blueprint 或 mount 驱动的最终菜单结果
- Modify: `backend/internal/plugins/menu_composer.go`
  - 不再把菜单当成平面排序列表，而是从 blueprint 与 mount 输入组合生成
- Create: `backend/internal/plugins/menu_blueprint.go`
  - 定义系统拥有的默认菜单骨架
- Create: `backend/internal/plugins/menu_mounts.go`
  - 将 built-in、CRD 与 plugin capabilities 转成统一 menu mounts
- Create: `backend/internal/plugins/menu_resolution.go`
  - 解析 availability 状态与最终菜单项
- Modify: `backend/internal/plugins/kernel_snapshot.go`
  - 暴露组合后的菜单结果
- Modify: `backend/internal/api/kernel_handler.go`
  - 通过 kernel snapshot 输出新的菜单结构

### Frontend 区域

- Modify: `frontend/src/kernel/contracts/menuContribution.ts`
  - 让前端菜单类型对齐“组合结果”，而不是平面注册项
- Create: `frontend/src/kernel/runtime/menu/types.ts`
  - 前端菜单运行时结果类型
- Create: `frontend/src/kernel/runtime/menu/resolveMenuState.ts`
  - 将后端菜单组合结果解析为 UI 可消费的导航组与状态
- Modify: `frontend/src/kernel/runtime/composeKernelNavigation.ts`
  - 不再对平面列表排序，而是渲染分组后的组合导航
- Modify: `frontend/src/App.tsx`
  - 消费分组菜单结果，而不是平面按钮列表

### Shared context 区域

- Create: `frontend/src/kernel/runtime/context/types.ts`
  - active cluster、namespace scope、workflow domain、resource identity 与 continuity 类型
- Create: `frontend/src/kernel/runtime/context/reducer.ts`
  - 处理导航、资源进入、动作开始与结束时的上下文迁移
- Create: `frontend/src/kernel/runtime/context/selectors.ts`
  - active cluster、namespace scope、current domain、current resource 等 UI selector
- Modify: `frontend/src/kernel/runtime/KernelRuntimeContext.tsx`
  - 在 kernel snapshot 状态之外承载共享工作上下文

### Resource page 区域

- Create: `frontend/src/kernel/resource-pages/types.ts`
  - resource capability、tab capability、tab replacement、page takeover 类型
- Create: `frontend/src/kernel/resource-pages/resolveResourcePage.ts`
  - 对单个资源身份做最终页面解析
- Create: `frontend/src/kernel/resource-pages/ResourcePageShell.tsx`
  - 产品级的共享资源页外壳
- Create: `frontend/src/kernel/resource-pages/DefaultOverviewTab.tsx`
  - 默认 overview tab
- Create: `frontend/src/kernel/resource-pages/DefaultYamlTab.tsx`
  - 默认 YAML tab
- Create: `frontend/src/kernel/resource-pages/tabs.ts`
  - 默认 tabs 注册与排序规则
- Modify: `frontend/src/components/page-shell/ResourcePageShell.tsx`
  - 让它退回纯布局用途，或者在产品级 shell 成立后移除

### I18n 最低运行时区域

- Modify: `frontend/src/i18n/types.ts`
  - 定义产品级 locale state 结构
- Modify: `frontend/src/i18n/copy.ts`
  - 让 runtime 不再只假定直接使用 `'en'`
- Create: `frontend/src/i18n/localeContext.tsx`
  - 最小产品级 locale boundary
- Modify: `frontend/src/i18n/messages/en.ts`
  - 补菜单状态、资源页 tabs 与工作上下文 UI 所需文案

## 4. Chunk 1：组合式菜单运行时

### Task 1：定义菜单组合结果合同

**Files:**
- Modify: `backend/pkg/sdk/menu.go`
- Modify: `frontend/src/kernel/contracts/menuContribution.ts`
- Test: `backend/internal/plugins/menu_composer_test.go`
- Test: `frontend/src/kernel/runtime/composeKernelNavigation.test.ts`

- [ ] **Step 1: 先写 backend failing test，覆盖 grouped composed menu descriptors**
- [ ] **Step 2: 运行 `mise exec go@1.25.1 -- go test ./backend/internal/plugins/...`，确认因缺少 grouped menu fields 而失败**
- [ ] **Step 3: 写最小 backend descriptor 结构，加入 menu group、availability state 与 fallback-entry 支持**
- [ ] **Step 4: 先写 frontend failing test，覆盖 grouped navigation 输入**
- [ ] **Step 5: 运行 `cd frontend && npm test -- --run composeKernelNavigation`，确认失败**
- [ ] **Step 6: 写最小 frontend 类型更新**
- [ ] **Step 7: 重新运行受影响的 backend 与 frontend 测试直到通过**
- [ ] **Step 8: Commit**

### Task 2：实现 backend 菜单 blueprint 与 mount 解析

**Files:**
- Create: `backend/internal/plugins/menu_blueprint.go`
- Create: `backend/internal/plugins/menu_mounts.go`
- Create: `backend/internal/plugins/menu_resolution.go`
- Modify: `backend/internal/plugins/menu_composer.go`
- Modify: `backend/internal/plugins/kernel_snapshot.go`
- Test: `backend/internal/plugins/menu_composer_test.go`

- [ ] **Step 1: 先写 failing tests，覆盖 blueprint groups、mount placement、disabled-unavailable entries 与 `CRDs` fallback**
- [ ] **Step 2: 运行 `mise exec go@1.25.1 -- go test ./backend/internal/plugins/...`，确认失败原因与缺失的组合规则一致**
- [ ] **Step 3: 实现最小 backend blueprint、mount conversion 与 resolution 路径**
- [ ] **Step 4: 重新运行 backend plugin tests 直到通过**
- [ ] **Step 5: Commit**

### Task 3：在 frontend shell 中渲染组合式菜单结果

**Files:**
- Create: `frontend/src/kernel/runtime/menu/types.ts`
- Create: `frontend/src/kernel/runtime/menu/resolveMenuState.ts`
- Modify: `frontend/src/kernel/runtime/composeKernelNavigation.ts`
- Modify: `frontend/src/App.tsx`
- Test: `frontend/src/kernel/runtime/composeKernelNavigation.test.ts`
- Test: `frontend/src/App.test.tsx`

- [ ] **Step 1: 先写 failing tests，覆盖 grouped rendering、disabled unavailable entries 与稳定的 `CRDs` fallback 显示**
- [ ] **Step 2: 运行 `cd frontend && npm test -- --run`，确认失败**
- [ ] **Step 3: 实现最小 grouped navigation 解析与 shell 渲染**
- [ ] **Step 4: 重新运行 frontend tests 直到通过**
- [ ] **Step 5: Commit**

## 5. Chunk 2：共享工作上下文运行时

### Task 4：恢复产品级工作上下文

**Files:**
- Create: `frontend/src/kernel/runtime/context/types.ts`
- Create: `frontend/src/kernel/runtime/context/reducer.ts`
- Create: `frontend/src/kernel/runtime/context/selectors.ts`
- Modify: `frontend/src/kernel/runtime/KernelRuntimeContext.tsx`
- Test: `frontend/src/kernel/runtime/context/reducer.test.ts`

- [ ] **Step 1: 先写 reducer failing tests，覆盖 cluster switch、namespace-scope update、workflow entry 与 resource entry**
- [ ] **Step 2: 运行 `cd frontend && npm test -- --run reducer`，确认失败**
- [ ] **Step 3: 实现最小 shared context reducer 与 selectors**
- [ ] **Step 4: 重新运行 reducer tests 直到通过**
- [ ] **Step 5: Commit**

### Task 5：将菜单导航接入工作上下文

**Files:**
- Modify: `frontend/src/kernel/runtime/KernelRuntimeContext.tsx`
- Modify: `frontend/src/App.tsx`
- Test: `frontend/src/App.test.tsx`

- [ ] **Step 1: 先写 app failing tests，覆盖菜单导航时 workflow-domain continuity 与 namespace continuity**
- [ ] **Step 2: 运行 `cd frontend && npm test -- --run App`，确认因缺少 context transitions 失败**
- [ ] **Step 3: 实现最小 navigation-to-context 接线**
- [ ] **Step 4: 重新运行 frontend app tests 直到通过**
- [ ] **Step 5: Commit**

## 6. Chunk 3：资源页外壳运行时

### Task 6：定义资源页 capability types

**Files:**
- Create: `frontend/src/kernel/resource-pages/types.ts`
- Create: `frontend/src/kernel/resource-pages/tabs.ts`
- Test: `frontend/src/kernel/resource-pages/tabs.test.ts`

- [ ] **Step 1: 先写 failing tests，覆盖默认 `Overview` 与 `YAML` tab 解析**
- [ ] **Step 2: 运行 `cd frontend && npm test -- --run tabs`，确认失败**
- [ ] **Step 3: 实现最小 capability 与 default-tab 定义**
- [ ] **Step 4: 重新运行 tab tests 直到通过**
- [ ] **Step 5: Commit**

### Task 7：实现共享资源页外壳

**Files:**
- Create: `frontend/src/kernel/resource-pages/ResourcePageShell.tsx`
- Create: `frontend/src/kernel/resource-pages/DefaultOverviewTab.tsx`
- Create: `frontend/src/kernel/resource-pages/DefaultYamlTab.tsx`
- Create: `frontend/src/kernel/resource-pages/resolveResourcePage.ts`
- Test: `frontend/src/kernel/resource-pages/ResourcePageShell.test.tsx`

- [ ] **Step 1: 先写 failing tests，覆盖 shell 渲染与默认 tabs 可见性**
- [ ] **Step 2: 运行 `cd frontend && npm test -- --run ResourcePageShell`，确认失败**
- [ ] **Step 3: 实现最小 shared shell 与 default tabs 路径**
- [ ] **Step 4: 重新运行 resource-page tests 直到通过**
- [ ] **Step 5: Commit**

### Task 8：将一条资源路径接入共享 shell

**Files:**
- Modify: `frontend/src/kernel/builtins/pages/WorkloadsPage.tsx`
- Modify: `frontend/src/App.tsx`
- Test: `frontend/src/App.test.tsx`

- [ ] **Step 1: 先写 failing tests，覆盖从 workload list 进入资源页并看到 `Overview` 与 `YAML`**
- [ ] **Step 2: 运行 `cd frontend && npm test -- --run App`，确认失败**
- [ ] **Step 3: 实现最小资源进入路径，并接入共享 shell**
- [ ] **Step 4: 重新运行 frontend app tests 直到通过**
- [ ] **Step 5: Commit**

## 7. Chunk 4：最低 i18n 运行时补全

### Task 9：补最小产品级 locale state

**Files:**
- Modify: `frontend/src/i18n/types.ts`
- Modify: `frontend/src/i18n/copy.ts`
- Create: `frontend/src/i18n/localeContext.tsx`
- Test: `frontend/src/i18n/copy.test.ts`

- [ ] **Step 1: 先写 failing tests，覆盖 locale-context-driven copy access**
- [ ] **Step 2: 运行 `cd frontend && npm test -- --run copy`，确认失败**
- [ ] **Step 3: 实现最小 locale boundary，不扩展成完整 i18n 产品系统**
- [ ] **Step 4: 重新运行 i18n tests 直到通过**
- [ ] **Step 5: Commit**

## 8. Chunk 5：接回首条真实工作流

### Task 10：在修正后的基础设施上恢复首条工作流

**Files:**
- Modify: `frontend/src/kernel/builtins/pages/HomepagePage.tsx`
- Modify: `frontend/src/kernel/builtins/pages/WorkloadsPage.tsx`
- Modify: `frontend/src/kernel/runtime/KernelRuntimeContext.tsx`
- Modify: `backend/internal/plugins/workload_provider.go`
- Modify: `backend/internal/plugins/action_executor.go`
- Test: `frontend/src/App.test.tsx`
- Test: `backend/internal/api/kernel_handler_test.go`

- [ ] **Step 1: 先写 integration failing tests，覆盖 `Homepage -> Workloads -> Resource Page -> Action -> Result -> Return`**
- [ ] **Step 2: 运行受影响的 frontend 与 backend tests，确认失败**
- [ ] **Step 3: 在新的菜单与资源页系统之上实现最小工作流恢复**
- [ ] **Step 4: 重新运行受影响的 frontend 与 backend tests 直到通过**
- [ ] **Step 5: Commit**

## 9. 验证命令

每个 chunk 收尾时使用以下命令验证：

- `cd frontend && npm test -- --run`
- `cd frontend && npm run build`
- `export GOPATH=/tmp/go GOMODCACHE=/tmp/go/pkg/mod GOCACHE=/tmp/go/build-cache GOTOOLCHAIN=local && mise exec go@1.25.1 -- go test ./...`
- `export GOPATH=/tmp/go GOMODCACHE=/tmp/go/pkg/mod GOCACHE=/tmp/go/build-cache GOTOOLCHAIN=local && mise exec go@1.25.1 -- go build ./...`

## 10. 退出条件

只有在以下条件都成立时，本补全计划才算完成：

- 菜单运行时已经从平面列表升级为组合结果，
- 共享工作上下文已经是真实状态而不是硬编码，
- 资源页外壳与默认 tabs 已经成立，
- 最低 i18n runtime state 已经存在，
- 首条工作流运行在这些层之上，而不是绕开它们。
