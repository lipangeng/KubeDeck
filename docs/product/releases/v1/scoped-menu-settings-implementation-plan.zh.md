# Scoped Menu Settings 实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**目标：** 为 KubeDeck 增加首批 scoped menu settings 能力，使系统能够在工作菜单、系统配置菜单、集群配置菜单之间切换，并支持 scoped 的 `pin`、`hide`、`reset` 操作。

**架构方式：** 继续保持一套统一的动态菜单系统，并将其扩展为 scope-aware 组合模型。后端仍然是 blueprint、mount、override 与最终 menu groups 的权威来源；前端只负责切换 scope、按 active cluster 加载组合后的菜单结果，并通过 preferences API 提交最小菜单定制操作。

**技术栈：** Go backend、React + TypeScript frontend、Vitest、Go test、MUI、现有 kernel runtime 与菜单组合模型。

---

## 文件结构

### Backend

- 修改：`backend/internal/plugins/menu_model.go`
  为 override 增加明确的 scope 字段与顺序支持字段。
- 修改：`backend/internal/plugins/menu_composer.go`
  将菜单组合切换为基于 scope 的 override 输入。
- 修改：`backend/internal/plugins/menu_resolution.go`
  在 `pin` 和 `hide` 之外，解析 `groupOrderOverrides` 与 `itemOrderOverrides`。
- 修改：`backend/internal/plugins/kernel_snapshot.go`
  为工作菜单、系统配置菜单、集群配置菜单输出 scoped menu results。
- 修改：`backend/internal/api/kernel_handler.go`
  在 snapshot 与 menu preference API 中接收 `scope`。
- 修改：`backend/internal/api/router.go`
  保持 menu preference API 继续挂在当前路由族下。
- 修改：`backend/internal/storage/repo_interfaces.go`
  让菜单 override 以 user、cluster、scope 为键进行存取。
- 测试：`backend/internal/plugins/menu_composer_test.go`
  锁定 scoped 组合规则。
- 测试：`backend/internal/api/kernel_handler_test.go`
  锁定 scoped preference API 行为。

### Frontend

- 修改：`frontend/src/kernel/runtime/transport.ts`
  扩展 remote contract，支持 scoped preference payload 与 scoped menu metadata。
- 修改：`frontend/src/kernel/runtime/fetchKernelMetadata.ts`
  请求时同时携带 cluster 与 active menu scope。
- 修改：`frontend/src/kernel/runtime/KernelRuntimeContext.tsx`
  增加当前 menu scope、scope 切换、scoped metadata reload 与 scoped preference 更新入口。
- 修改：`frontend/src/App.tsx`
  增加可见的 scope 入口：
  - 右上角系统配置入口
  - 左侧底部集群配置入口
  - 配置空间中的 `Back to Work`
- 新增：`frontend/src/kernel/runtime/updateMenuPreferences.ts`
  专门负责保存 scoped menu settings 的 transport helper。
- 新增：`frontend/src/features/menu-settings/MenuSettingsPanel.tsx`
  提供首批 `pin`、`hide`、`reset current scope` 的设置 UI。
- 新增：`frontend/src/features/menu-settings/MenuSettingsPanel.test.tsx`
  独立验证设置 UI 行为。
- 修改：`frontend/src/App.test.tsx`
  验证 scope 切换与菜单刷新。
- 修改：`frontend/src/i18n/messages/en.ts`
  为菜单设置与 scope 切换增加新文案。

### Documentation

- 修改：`docs/product/architecture/cluster-aware-menu-composition.md`
- 修改：`docs/product/architecture/cluster-aware-menu-composition.zh.md`
  如果执行过程中行为细节有变化，及时同步架构文档。

## Chunk 1：扩展 Scoped Override 合同

### Task 1：先写 backend 失败测试，锁定 scoped override 字段

**文件：**
- 修改：`backend/internal/plugins/menu_composer_test.go`

- [ ] **Step 1: 先写失败测试**

增加测试，验证：
- `MenuOverride` 接受 `scope`
- `groupOrderOverrides` 能生效
- `itemOrderOverrides` 能在组内生效
- `system` 与 `cluster` 是合法的组合目标 scope

- [ ] **Step 2: 运行测试，确认失败**

运行：

```bash
cd backend && GOPATH=/tmp/go GOMODCACHE=/tmp/go/pkg/mod GOCACHE=/tmp/go/build-cache GOTOOLCHAIN=local mise exec go@1.25.1 -- go test ./internal/plugins -run TestComposeMenus
```

预期：FAIL，因为当前 override 模型还不支持新的 scoped 顺序字段。

- [ ] **Step 3: 编写最小实现**

修改：
- `backend/internal/plugins/menu_model.go`
- `backend/internal/plugins/menu_resolution.go`

增加：
- scoped override 结构
- `groupOrderOverrides`
- `itemOrderOverrides`
- 在 blueprint 和 mount 解析后的顺序应用逻辑

- [ ] **Step 4: 再次运行测试，确认通过**

运行相同命令并确认 PASS。

- [ ] **Step 5: 提交**

```bash
git add backend/internal/plugins/menu_model.go backend/internal/plugins/menu_resolution.go backend/internal/plugins/menu_composer_test.go
git commit -m "feat(menu): extend scoped override model"
```

## Chunk 2：补齐 Scoped Preference 持久化与 Snapshot 解析

### Task 2：先写 backend API 失败测试，锁定 `scope`

**文件：**
- 修改：`backend/internal/api/kernel_handler_test.go`
- 修改：`backend/internal/storage/repo_interfaces.go`

- [ ] **Step 1: 先写失败测试**

增加测试，验证：
- `GET /api/preferences/menu?scope=system`
- `PUT /api/preferences/menu?scope=cluster&cluster=prod-eu1`
- `GET /api/meta/kernel?scope=cluster&cluster=prod-eu1`

每个测试都必须验证响应反映了请求的 scope，而不是继续走之前默认的工作菜单路径。

- [ ] **Step 2: 运行测试，确认失败**

运行：

```bash
cd backend && GOPATH=/tmp/go GOMODCACHE=/tmp/go/pkg/mod GOCACHE=/tmp/go/build-cache GOTOOLCHAIN=local mise exec go@1.25.1 -- go test ./internal/api -run 'TestKernelHandler(MenuPreferences|Snapshot)'
```

预期：FAIL，因为 API 还不理解 scoped menu settings。

- [ ] **Step 3: 编写最小实现**

修改：
- `backend/internal/api/kernel_handler.go`
- `backend/internal/storage/repo_interfaces.go`
- `backend/internal/plugins/kernel_snapshot.go`

增加：
- scope-aware preference lookup
- scope-aware snapshot composition
- 按 user + cluster + scope 进行持久化

- [ ] **Step 4: 再次运行测试，确认通过**

运行相同命令并确认 PASS。

- [ ] **Step 5: 提交**

```bash
git add backend/internal/api/kernel_handler.go backend/internal/api/kernel_handler_test.go backend/internal/plugins/kernel_snapshot.go backend/internal/storage/repo_interfaces.go
git commit -m "feat(menu): persist scoped menu preferences"
```

## Chunk 3：补齐 Frontend Scope Switching Runtime

### Task 3：先写 frontend runtime 失败测试，锁定 scope 切换

**文件：**
- 修改：`frontend/src/App.test.tsx`
- 修改：`frontend/src/kernel/runtime/KernelRuntimeContext.tsx`

- [ ] **Step 1: 先写失败测试**

增加测试，验证：
- 点击右上角系统配置入口后，菜单切换为 system scope
- 点击左侧底部集群配置入口后，菜单切换为 cluster scope
- `Back to Work` 能返回工作菜单
- metadata 请求中会带上 active scope

- [ ] **Step 2: 运行测试，确认失败**

运行：

```bash
cd frontend && npm test -- --run src/App.test.tsx
```

预期：FAIL，因为 runtime 当前还没有 scoped menu-space 切换能力。

- [ ] **Step 3: 编写最小实现**

修改：
- `frontend/src/kernel/runtime/fetchKernelMetadata.ts`
- `frontend/src/kernel/runtime/KernelRuntimeContext.tsx`
- `frontend/src/App.tsx`

增加：
- 当前 menu scope 状态
- scope 切换动作
- 右上角 `System Settings`
- 左下角 `Cluster Settings`
- `Back to Work`

- [ ] **Step 4: 再次运行测试，确认通过**

运行相同命令并确认 PASS。

- [ ] **Step 5: 提交**

```bash
git add frontend/src/kernel/runtime/fetchKernelMetadata.ts frontend/src/kernel/runtime/KernelRuntimeContext.tsx frontend/src/App.tsx frontend/src/App.test.tsx
git commit -m "feat(menu): add scoped menu space switching"
```

## Chunk 4：补齐首版 Menu Settings Panel

### Task 4：先写 UI 失败测试，锁定 `pin`、`hide`、`reset`

**文件：**
- 新增：`frontend/src/features/menu-settings/MenuSettingsPanel.test.tsx`
- 新增：`frontend/src/features/menu-settings/MenuSettingsPanel.tsx`
- 新增：`frontend/src/kernel/runtime/updateMenuPreferences.ts`
- 修改：`frontend/src/i18n/messages/en.ts`

- [ ] **Step 1: 先写失败测试**

增加测试，验证：
- 面板能显示当前 scope 下的最终组合菜单
- `Pin` 会发送新的 scoped override payload
- `Hide` 会发送新的 scoped override payload
- `Reset current scope` 会清空该 scope 的 override payload

- [ ] **Step 2: 运行测试，确认失败**

运行：

```bash
cd frontend && npm test -- --run src/features/menu-settings/MenuSettingsPanel.test.tsx
```

预期：FAIL，因为设置面板与更新 transport 还不存在。

- [ ] **Step 3: 编写最小实现**

新增：
- `updateMenuPreferences.ts`
- `MenuSettingsPanel.tsx`
- `en.ts` 中的新文案键

只把面板接入 system 和 cluster 两种配置空间。

- [ ] **Step 4: 再次运行测试，确认通过**

运行相同命令并确认 PASS。

- [ ] **Step 5: 提交**

```bash
git add frontend/src/features/menu-settings/MenuSettingsPanel.tsx frontend/src/features/menu-settings/MenuSettingsPanel.test.tsx frontend/src/kernel/runtime/updateMenuPreferences.ts frontend/src/i18n/messages/en.ts
git commit -m "feat(menu): add first scoped menu settings panel"
```

## Chunk 5：全量验证与文档同步

### Task 5：验证并同步文档

**文件：**
- 修改：`docs/product/architecture/cluster-aware-menu-composition.md`
- 修改：`docs/product/architecture/cluster-aware-menu-composition.zh.md`

- [ ] **Step 1: 运行 backend 验证**

运行：

```bash
cd backend && GOPATH=/tmp/go GOMODCACHE=/tmp/go/pkg/mod GOCACHE=/tmp/go/build-cache GOTOOLCHAIN=local mise exec go@1.25.1 -- go test ./... && mise exec go@1.25.1 -- go build ./...
```

预期：PASS。

- [ ] **Step 2: 运行 frontend 验证**

运行：

```bash
cd frontend && npm test -- --run && npm run build
```

预期：PASS。

- [ ] **Step 3: 如实现细节变化则同步文档**

如果执行过程中 scope 名称或入口路径有调整，及时更新架构文档。

- [ ] **Step 4: 提交**

```bash
git add docs/product/architecture/cluster-aware-menu-composition.md docs/product/architecture/cluster-aware-menu-composition.zh.md
git commit -m "docs(menu): sync scoped menu settings implementation"
```
