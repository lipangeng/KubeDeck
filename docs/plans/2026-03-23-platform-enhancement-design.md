# KubeDeck 平台增强设计

**日期:** 2026-03-23  
**版本:** v2.0.0  
**状态:** Approved

---

## 1. 概述

本文档定义 KubeDeck v2.0.0 的增强功能设计，包括数据库集成、OAuth2 认证、RBAC 权限系统，以及 Pod、Ingress 等 Kubernetes 资源对象的定制页面。

### 1.1 设计目标

1. **企业级认证** - 集成外部 OAuth2 Provider，实现单点登录
2. **细粒度权限** - 基于 K8s RBAC 的跨集群权限管理
3. **多数据库支持** - SQLite/MySQL/PostgreSQL 抽象层
4. **资源管理增强** - Pod 日志/Terminal、Ingress 规则可视化等

### 1.2 参考设计

- [kite-org/kite](https://github.com/kite-org/kite) - OAuth2 和 RBAC 实现参考
- Kite Pod 详情页 - 日志和 Terminal 交互设计
- Kite Ingress 管理 - 规则可视化配置

---

## 2. 架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        KubeDeck Platform                         │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Frontend  │  │   Backend   │  │   External Services     │  │
│  │  (React)    │  │    (Go)     │  │                         │  │
│  │             │  │             │  │  ┌──────────────────┐   │  │
│  │  ┌───────┐  │  │  ┌───────┐  │  │  │ OAuth2 Provider  │   │  │
│  │  │ Pod   │  │  │  │ Auth  │◄─┼─────►│ (GitHub/Google)  │   │  │
│  │  │ Page  │  │  │  │ Layer │  │  │  └──────────────────┘   │  │
│  │  ├───────┤  │  │  ├───────┤  │  │  ┌──────────────────┐   │  │
│  │  │Ingress│  │  │  │ RBAC  │  │  │  │  Kubernetes API  │   │  │
│  │  │ Page  │  │  │  │ Layer │◄─┼─────►│   (RBAC AuthZ)   │   │  │
│  │  ├───────┤  │  │  ├───────┤  │  │  └──────────────────┘   │  │
│  │  │ ...   │  │  │  │  DB   │  │  │  ┌──────────────────┐   │  │
│  │  └───────┘  │  │  │ Layer │  │  │  │  Database        │   │  │
│  └─────────────┘  │  └───────┘  │  │  │ (SQLite/MySQL/PG)│   │  │
│                   └─────────────┘  │  └──────────────────┘   │  │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 组件职责

| 组件 | 职责 | 技术选型 |
|------|------|----------|
| Auth Layer | OAuth2 回调、JWT 验证、会话管理 | golang.org/x/oauth2 |
| RBAC Layer | 角色管理、权限映射、K8s Binding | Kubernetes client-go |
| DB Layer | 用户存储、配置存储、审计日志 | GORM + SQLite/MySQL/PG |
| Pod Page | 日志流、Exec Terminal、文件管理 | WebSocket + xterm.js |
| Ingress Page | 规则可视化、TLS 配置 | Monaco Editor |

---

## 3. 数据库设计

### 3.1 数据库抽象层

采用 GORM 作为 ORM 框架，支持运行时切换数据库类型。

```go
// Database configuration
type DatabaseConfig struct {
    Type     string // sqlite, mysql, postgres
    DSN      string
    MaxIdle  int
    MaxOpen  int
}

// Initialize database
func NewDatabase(cfg DatabaseConfig) (*gorm.DB, error) {
    var dialector gorm.Dialector
    switch cfg.Type {
    case "sqlite":
        dialector = sqlite.Open(cfg.DSN)
    case "mysql":
        dialector = mysql.Open(cfg.DSN)
    case "postgres":
        dialector = postgres.Open(cfg.DSN)
    }
    return gorm.Open(dialector, &gorm.Config{})
}
```

### 3.2 数据表设计

#### 3.2.1 用户表 (users)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| username | string | 用户名（唯一） |
| email | string | 邮箱 |
| oauth_provider | string | OAuth Provider 名称 |
| oauth_sub | string | OAuth Provider 用户 ID |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

#### 3.2.2 角色表 (roles)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | string | 角色名称 |
| description | string | 角色描述 |
| rules | JSON | 权限规则（K8s PolicyRule） |
| is_builtin | boolean | 是否内置角色 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

#### 3.2.3 角色绑定表 (role_bindings)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| role_id | UUID | 角色 ID |
| user_id | UUID | 用户 ID |
| cluster_ids | JSON | 绑定的集群 ID 列表 |
| namespaces | JSON | 命名空间范围（* 表示全部） |
| created_at | timestamp | 创建时间 |

#### 3.2.4 审计日志表 (audit_logs)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| user_id | UUID | 用户 ID |
| action | string | 操作类型 |
| resource | string | 资源类型 |
| resource_name | string | 资源名称 |
| namespace | string | 命名空间 |
| cluster_id | string | 集群 ID |
| status | string | 操作状态（success/failed） |
| timestamp | timestamp | 操作时间 |
| details | JSON | 详细信息 |

---

## 4. OAuth2 认证设计

### 4.1 架构设计

采用外部 OAuth2 代理模式，后端仅验证 JWT Token。

```
┌──────────┐     ┌─────────────┐     ┌──────────────────┐
│  User    │────►│ OAuth2      │────►│ KubeDeck Backend │
│  Browser │     │ Provider    │     │                  │
└──────────┘     └─────────────┘     └──────────────────┘
                      │                      │
                      │                      ▼
                      │              ┌──────────────────┐
                      │              │ Kubernetes API   │
                      │              │ (RBAC AuthZ)     │
                      └─────────────►│                  │
                                     └──────────────────┘
```

### 4.2 支持的 OAuth2 Provider

- GitHub
- Google
- GitLab
- Microsoft (Azure AD / Entra ID)
- OIDC (通用)

### 4.3 认证流程

1. 用户访问 `/login`
2. 重定向到 OAuth2 Provider 授权页面
3. 用户授权后回调到 `/auth/callback`
4. 后端验证 code，获取 access_token 和 id_token
5. 解析 JWT，提取用户信息
6. 创建/更新本地用户记录
7. 生成会话 Cookie，重定向到首页

### 4.4 JWT Token 结构

```json
{
  "sub": "user-uuid",
  "name": "John Doe",
  "email": "john@example.com",
  "groups": ["admin", "developer"],
  "iat": 1711180800,
  "exp": 1711267200
}
```

---

## 5. RBAC 权限系统设计

### 5.1 权限模型

采用 K8s RBAC 映射模型：

```
KubeDeck Role → K8s ClusterRole/Role
KubeDeck RoleBinding → K8s ClusterRoleBinding/RoleBinding
```

### 5.2 内置角色

| 角色 | 权限范围 | 说明 |
|------|----------|------|
| admin | 所有资源所有操作 | 集群管理员 |
| editor | 大部分资源可写 | 开发者 |
| viewer | 所有资源只读 | 观察者 |
| namespace-admin | 命名空间内所有操作 | 命名空间管理员 |

### 5.3 跨集群角色绑定

```yaml
apiVersion: kubedeck.io/v1
kind: GlobalRoleBinding
metadata:
  name: developer-binding
subjects:
- kind: User
  name: john@example.com
roleRef:
  kind: ClusterRole
  name: developer
clusters:
- cluster-a
- cluster-b
namespaces:
- default
- production
```

### 5.4 角色配置页面设计

```
┌─────────────────────────────────────────────────────────────┐
│  Create/Edit Role                              [Save] [Cancel] │
├─────────────────────────────────────────────────────────────┤
│  Role Name: [________________]  Description: [______________] │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ Permission Rules                                         │  │
│  ├─────────────────────────────────────────────────────────┤  │
│  │  ┌─────────────────────┐  ┌──────────────────────────┐  │  │
│  │  │ Resource Types      │  │ Operations               │  │  │
│  │  │ ☑ Pods              │  │ ☑ get    ☑ list         │  │  │
│  │  │ ☑ Deployments       │  │ ☑ create ☑ update       │  │  │
│  │  │ ☑ Services          │  │ ☑ delete ☑ watch        │  │  │
│  │  │ ☑ ConfigMaps        │  │                          │  │  │
│  │  │ ☑ Secrets           │  │ [Select All] [Clear All] │  │  │
│  │  │ ☑ Ingresses         │  │                          │  │  │
│  │  │ ☐ PersistentVolumes │  │                          │  │  │
│  │  │ ☐ Nodes             │  │                          │  │  │
│  │  └─────────────────────┘  └──────────────────────────┘  │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                                 │
│  Cluster Binding: ☑ All Clusters  ○ Select Clusters...        │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ YAML Preview                                             │  │
│  │ apiVersion: rbac.authorization.k8s.io/v1                │  │
│  │ kind: ClusterRole                                        │  │
│  │ metadata:                                                │  │
│  │   name: custom-role                                      │  │
│  │ rules:                                                   │  │
│  │ - apiGroups: [""]                                        │  │
│  │   resources: ["pods", "services"]                        │  │
│  │   verbs: ["get", "list", "create"]                       │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                                 │
│  [Use Visual Editor]  [Switch to YAML]                         │
└─────────────────────────────────────────────────────────────┘
```

---

## 6. Pod 详情页设计

### 6.1 页面结构

```
┌─────────────────────────────────────────────────────────────┐
│ Pod: api-server-7d8f9b6c5-xk2p1    Namespace: default       │
│ Status: Running    Node: node-1    IP: 10.244.1.15         │
├─────────────────────────────────────────────────────────────┤
│ [Overview] [Logs] [Terminal] [Files] [Events] [YAML]       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Containers (3)                                             │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Name       Image          Status    Restarts    Age    │ │
│  ├────────────────────────────────────────────────────────┤ │
│  │ api        myapp:v1.2.3   Running   0          2d      │ │
│  │ sidecar    fluent-bit     Running   1          2d      │ │
│  │ init       busybox        Terminated 0         2d      │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  Environment Variables                                       │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ DATABASE_URL    postgres://db:5432/app                 │ │
│  │ LOG_LEVEL       info                                   │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  Volumes                                                     │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ config-volume    /etc/config    configmap/app-config   │ │
│  │ data-volume      /data          pvc/data-pvc           │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 6.2 日志 Tab 设计

```
┌─────────────────────────────────────────────────────────────┐
│ Container: [api ▼]  Pod: api-7d8f9b6c5-xk2p1                │
├─────────────────────────────────────────────────────────────┤
│ [▶ Follow] [⏸ Pause] [⬇ Download]  Filter: [____________]  │
│ Timestamp: ☑ Show  │  Wrap: ☑ On                           │
├─────────────────────────────────────────────────────────────┤
│ 2026-03-23 10:15:32.123 INFO  Starting application...       │
│ 2026-03-23 10:15:32.456 INFO  Connected to database         │
│ 2026-03-23 10:15:33.789 WARN  High memory usage detected    │
│ 2026-03-23 10:15:34.012 ERROR Failed to connect to cache    │
│                                                            │
└─────────────────────────────────────────────────────────────┘
```

### 6.3 Terminal Tab 设计

```
┌─────────────────────────────────────────────────────────────┐
│ Container: [api ▼]  Shell: [/bin/bash ▼]                    │
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ root@api-7d8f9b6c5-xk2p1:/app# ls -la                   │ │
│ │ total 48                                                │ │
│ │ drwxr-xr-x  4 root root 4096 Mar 23 10:15 .            │ │
│ │ drwxr-xr-x  3 root root 4096 Mar 23 10:15 ..           │ │
│ │ -rw-r--r--  1 root root  234 Mar 23 10:15 main.go      │ │
│ │ root@api-7d8f9b6c5-xk2p1:/app# _                       │ │
│ └─────────────────────────────────────────────────────────┘ │
│ [Disconnect] [Clear] [Fullscreen]                          │
└─────────────────────────────────────────────────────────────┘
```

### 6.4 实现阶段

| 阶段 | 功能 | 工期 |
|------|------|------|
| Phase 1 | 基础信息 + 日志查看 | 1 周 |
| Phase 2 | Web Terminal (Exec) | 1 周 |
| Phase 3 | 文件管理 | 1 周 |

---

## 7. Ingress 详情页设计

### 7.1 页面结构

```
┌─────────────────────────────────────────────────────────────┐
│ Ingress: api-ingress    Namespace: default                  │
│ Class: nginx    Address: 192.168.1.100                      │
├─────────────────────────────────────────────────────────────┤
│ [Rules] [TLS] [Events] [YAML]                               │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Rules                                                       │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Host          Path        Service        Port          │ │
│  ├────────────────────────────────────────────────────────┤ │
│  │ api.example.com  /api    api-service     8080    [Edit]│ │
│  │ *.example.com    /       web-service     80      [Edit]│ │
│  │                                                         │ │
│  │ [+ Add Rule]                                            │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  TLS                                                         │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Hosts                    Secret         Status          │ │
│  ├────────────────────────────────────────────────────────┤ │
│  │ api.example.com         api-tls        Valid           │ │
│  │                                                         │ │
│  │ [+ Add TLS]                                             │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 7.2 规则编辑弹窗

```
┌─────────────────────────────────────────────────────────────┐
│ Edit Ingress Rule                               [Save] [×]  │
├─────────────────────────────────────────────────────────────┤
│  Host: [api.example.com________________]                    │
│                                                              │
│  Path: [/api________________]  Type: [Prefix ▼]             │
│                                                              │
│  Backend Service                                             │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Service: [api-service ▼]  Port: [8080 ▼]               │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  Advanced Options                                            │
│  ☑ Rewrite Target: [/(.*)_______]                           │
│  ☐ SSL Redirect                                              │
│  ☐ Rate Limiting: [___] req/s                               │
└─────────────────────────────────────────────────────────────┘
```

---

## 8. 其他资源对象设计

### 8.1 Deployment 详情页

- **基础信息**: 副本数、策略、选择器
- **操作**: 扩缩容、重启、回滚
- **Pods Tab**: 关联的 Pod 列表
- **Events Tab**: 相关事件

### 8.2 Service 详情页

- **基础信息**: 类型、ClusterIP、Ports
- **Endpoints Tab**: 后端 Pod 列表
- **操作**: 编辑、删除

### 8.3 ConfigMap/Secret 详情页

- **数据列表**: Key-Value 展示
- **操作**: 编辑、复制、下载
- **引用查看**: 哪些 Pod 在使用

---

## 9. API 设计

### 9.1 认证相关

```
POST   /api/auth/login          # 登录
POST   /api/auth/logout         # 登出
GET    /api/auth/callback       # OAuth2 回调
GET    /api/auth/me             # 当前用户信息
```

### 9.2 用户管理

```
GET    /api/users               # 用户列表
GET    /api/users/:id           # 用户详情
PUT    /api/users/:id/roles     # 更新用户角色
DELETE /api/users/:id           # 删除用户
```

### 9.3 角色管理

```
GET    /api/roles               # 角色列表
POST   /api/roles               # 创建角色
PUT    /api/roles/:id           # 更新角色
DELETE /api/roles/:id           # 删除角色
POST   /api/roles/:id/bind      # 绑定用户/集群
```

### 9.4 Pod 相关

```
GET    /api/pods/:ns/:name      # Pod 详情
GET    /api/pods/:ns/:name/logs # Pod 日志 (WebSocket)
POST   /api/pods/:ns/:name/exec # Exec Terminal (WebSocket)
GET    /api/pods/:ns/:name/files# 文件列表
PUT    /api/pods/:ns/:name/files# 上传文件
```

### 9.5 Ingress 相关

```
GET    /api/ingresses/:ns/:name # Ingress 详情
PUT    /api/ingresses/:ns/:name # 更新 Ingress
```

---

## 10. 前端组件设计

### 10.1 新增组件

```
frontend/src/
├── features/
│   ├── auth/
│   │   ├── LoginPage.tsx
│   │   ├── OAuthCallback.tsx
│   │   └── UserProfile.tsx
│   ├── rbac/
│   │   ├── RoleList.tsx
│   │   ├── RoleEditor.tsx
│   │   ├── RoleBindingEditor.tsx
│   │   └── ClusterSelector.tsx
│   ├── pods/
│   │   ├── PodDetail.tsx
│   │   ├── PodLogs.tsx
│   │   ├── PodTerminal.tsx
│   │   └── PodFiles.tsx
│   └── ingress/
│       ├── IngressDetail.tsx
│       ├── IngressRules.tsx
│       └── IngressTLS.tsx
├── components/
│   ├── yaml-editor/
│   │   ├── YamlEditor.tsx
│   │   └── YamlPreview.tsx
│   ├── terminal/
│   │   ├── XtermWrapper.tsx
│   │   └── TerminalSession.tsx
│   └── log-viewer/
│       ├── LogViewer.tsx
│       └── LogFilters.tsx
└── kernel/
    └── builtins/
        └── pages/
            ├── PodPage.tsx
            ├── IngressPage.tsx
            ├── DeploymentPage.tsx
            └── ServicePage.tsx
```

### 10.2 新增依赖

```json
{
  "dependencies": {
    "@xterm/xterm": "^5.0.0",
    "@xterm/addon-fit": "^0.8.0",
    "@xterm/addon-attach": "^0.8.0",
    "@monaco-editor/react": "^4.6.0",
    "js-yaml": "^4.1.0"
  }
}
```

---

## 11. 安全设计

### 11.1 认证安全

- JWT Token 有效期 24 小时
- Refresh Token 机制
- HTTPS 强制
- CSRF 保护

### 11.2 授权安全

- 所有 API 请求经过 RBAC 校验
- K8s API 请求使用用户 Token 代理
- 审计日志记录所有操作

### 11.3 数据安全

- 数据库加密存储（可选）
- Secret 数据脱敏显示
- 终端会话加密（WebSocket over WSS）

---

## 12. 实施计划

### Phase 1: 基础架构（2 周）

1. 数据库抽象层
2. OAuth2 认证集成
3. 用户管理基础

### Phase 2: RBAC 系统（2 周）

1. 角色管理 CRUD
2. 角色绑定和跨集群同步
3. K8s RBAC 映射

### Phase 3: Pod 详情页（3 周）

1. Phase 1: 基础信息 + 日志
2. Phase 2: Web Terminal
3. Phase 3: 文件管理

### Phase 4: Ingress 和其他资源（2 周）

1. Ingress 详情页
2. Deployment/Service 详情页
3. ConfigMap/Secret 详情页

---

## 13. 验收标准

### 13.1 认证和授权

- [ ] 可通过 GitHub OAuth 登录
- [ ] 可创建自定义角色
- [ ] 角色可绑定到多个集群
- [ ] 权限限制生效

### 13.2 Pod 详情页

- [ ] 可查看 Pod 基础信息
- [ ] 可实时查看日志
- [ ] 可使用 Web Terminal
- [ ] 可管理容器文件

### 13.3 Ingress 详情页

- [ ] 可可视化编辑路由规则
- [ ] 可配置 TLS
- [ ] 可切换 YAML 编辑

---

## 14. 参考资料

- [Kite GitHub](https://github.com/kite-org/kite)
- [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [OAuth2 Spec](https://oauth.net/2/)
- [GORM Documentation](https://gorm.io/docs/)
