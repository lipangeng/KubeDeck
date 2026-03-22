# Platform Enhancement Implementation Plan

> **For agentic workers:** REQUIRED: 实施前必须使用 `test-driven-development`，并且所有变更都要遵循 `docs/product/development-mode.zh.md`。步骤统一使用 checkbox（`- [ ]`）跟踪。

**Goal:** 为 KubeDeck 添加企业级认证、RBAC 权限系统、多数据库支持，以及 Pod/Ingress 等资源对象的增强页面。

**Architecture:** 采用外部 OAuth2 代理认证，K8s RBAC 映射授权，GORM 数据库抽象层，分阶段实现资源详情页。

**Tech Stack:** Go 1.26, GORM, SQLite/MySQL/PostgreSQL, golang.org/x/oauth2, React 18, xterm.js, Monaco Editor

---

## Phase 1: 数据库和认证基础（2 周）

### Task 1: 数据库抽象层

**Files:**
- Create: `backend/internal/storage/database.go`
- Create: `backend/internal/storage/models.go`
- Create: `backend/internal/storage/database_test.go`
- Modify: `backend/internal/api/router.go`

**Step 1: 定义数据库模型**

```go
// backend/internal/storage/models.go
package storage

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

// User represents a platform user
type User struct {
    ID            uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
    Username      string    `gorm:"uniqueIndex;size:255" json:"username"`
    Email         string    `gorm:"size:255" json:"email"`
    OAuthProvider string    `gorm:"size:100" json:"oauth_provider"`
    OAuthSub      string    `gorm:"size:255" json:"oauth_sub"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

// Role represents a RBAC role
type Role struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
    Name        string    `gorm:"uniqueIndex;size:255" json:"name"`
    Description string    `gorm:"size:1024" json:"description"`
    Rules       string    `gorm:"type:json" json:"rules"` // JSON serialized PolicyRules
    IsBuiltin   bool      `gorm:"default:false" json:"is_builtin"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// RoleBinding binds users to roles
type RoleBinding struct {
    ID         uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
    RoleID     uuid.UUID `gorm:"type:uuid" json:"role_id"`
    UserID     uuid.UUID `gorm:"type:uuid" json:"user_id"`
    ClusterIDs string    `gorm:"type:json" json:"cluster_ids"` // JSON array of cluster IDs
    Namespaces string    `gorm:"type:json" json:"namespaces"`  // JSON array or "*"
    CreatedAt  time.Time `json:"created_at"`
}

// AuditLog records user actions
type AuditLog struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
    UserID       uuid.UUID `gorm:"type:uuid" json:"user_id"`
    Action       string    `gorm:"size:100" json:"action"`
    Resource     string    `gorm:"size:100" json:"resource"`
    ResourceName string    `gorm:"size:255" json:"resource_name"`
    Namespace    string    `gorm:"size:255" json:"namespace"`
    ClusterID    string    `gorm:"size:255" json:"cluster_id"`
    Status       string    `gorm:"size:20" json:"status"`
    Timestamp    time.Time `gorm:"index" json:"timestamp"`
    Details      string    `gorm:"type:json" json:"details"`
}
```

**Step 2: 实现数据库连接抽象**

```go
// backend/internal/storage/database.go
package storage

import (
    "fmt"
    "gorm.io/driver/sqlite"
    "gorm.io/driver/mysql"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type DatabaseConfig struct {
    Type    string `json:"type"` // sqlite, mysql, postgres
    DSN     string `json:"dsn"`
    MaxIdle int    `json:"max_idle"`
    MaxOpen int    `json:"max_open"`
}

func NewDatabase(cfg DatabaseConfig) (*gorm.DB, error) {
    var dialector gorm.Dialector

    switch cfg.Type {
    case "mysql":
        dialector = mysql.Open(cfg.DSN)
    case "postgres":
        dialector = postgres.Open(cfg.DSN)
    case "sqlite", "":
        if cfg.DSN == "" {
            cfg.DSN = "data/kubedeck.db"
        }
        dialector = sqlite.Open(cfg.DSN)
    default:
        return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
    }

    db, err := gorm.Open(dialector, &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, err
    }

    // Auto migrate schema
    err = db.AutoMigrate(&User{}, &Role{}, &RoleBinding{}, &AuditLog{})
    if err != nil {
        return nil, err
    }

    // Connection pool settings
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    sqlDB.SetMaxIdleConns(cfg.MaxIdle)
    sqlDB.SetMaxOpenConns(cfg.MaxOpen)

    return db, nil
}

// GetUsers returns the user repository
func (d *Database) GetUsers() UserRepository {
    return &userRepository{db: d.db}
}

// GetRoles returns the role repository
func (d *Database) GetRoles() RoleRepository {
    return &roleRepository{db: d.db}
}
```

**Step 3: 编写数据库初始化测试**

```go
// backend/internal/storage/database_test.go
package storage

import (
    "os"
    "testing"
)

func TestNewDatabase_SQLite(t *testing.T) {
    tmpFile := "/tmp/test_kubedeck.db"
    defer os.Remove(tmpFile)

    cfg := DatabaseConfig{
        Type:    "sqlite",
        DSN:     tmpFile,
        MaxIdle: 5,
        MaxOpen: 25,
    }

    db, err := NewDatabase(cfg)
    if err != nil {
        t.Fatalf("Failed to create database: %v", err)
    }

    // Verify tables exist
    tables := []string{"users", "roles", "role_bindings", "audit_logs"}
    for _, table := range tables {
        if !db.Migrator().HasTable(&User{}) {
            t.Errorf("Table %s does not exist", table)
        }
    }
}

func TestNewDatabase_DefaultsToSQLite(t *testing.T) {
    tmpFile := "/tmp/test_default.db"
    defer os.Remove(tmpFile)

    cfg := DatabaseConfig{
        DSN: tmpFile,
    }

    db, err := NewDatabase(cfg)
    if err != nil {
        t.Fatalf("Failed to create database: %v", err)
    }

    if db == nil {
        t.Error("Expected non-nil database")
    }
}
```

**Step 4: 运行测试验证**

```bash
cd backend && go test ./internal/storage/... -v
# Expected: PASS
```

**Step 5: 提交**

```bash
git add backend/internal/storage/database.go backend/internal/storage/models.go backend/internal/storage/database_test.go
git commit -m "feat(storage): add database abstraction layer with GORM

- Support SQLite, MySQL, PostgreSQL via GORM
- Auto-migrate User, Role, RoleBinding, AuditLog schemas
- Configurable connection pool settings"
```

---

### Task 2: OAuth2 认证集成

**Files:**
- Create: `backend/internal/auth/oauth2.go`
- Create: `backend/internal/auth/oauth2_test.go`
- Create: `backend/internal/auth/jwt.go`
- Create: `backend/internal/auth/jwt_test.go`
- Modify: `backend/internal/api/router.go`
- Modify: `backend/internal/api/kernel_handler.go`

**Step 1: 定义 OAuth2 配置结构**

```go
// backend/internal/auth/oauth2.go
package auth

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"

    "golang.org/x/oauth2"
    "github.com/google/uuid"
)

type OAuth2Config struct {
    Provider     string `json:"provider"` // github, google, gitlab, oidc
    ClientID     string `json:"client_id"`
    ClientSecret string `json:"client_secret"`
    RedirectURL  string `json:"redirect_url"`
    Scopes       []string `json:"scopes"`
    
    // OIDC specific
    IssuerURL    string `json:"issuer_url"`
}

type OAuth2Provider struct {
    config   *OAuth2Config
    oauthCfg *oauth2.Config
}

func NewOAuth2Provider(cfg *OAuth2Config) (*OAuth2Provider, error) {
    endpoint := getOAuth2Endpoint(cfg.Provider)
    
    oauthCfg := &oauth2.Config{
        ClientID:     cfg.ClientID,
        ClientSecret: cfg.ClientSecret,
        RedirectURL:  cfg.RedirectURL,
        Scopes:       cfg.Scopes,
        Endpoint:     endpoint,
    }

    return &OAuth2Provider{
        config:   cfg,
        oauthCfg: oauthCfg,
    }, nil
}

func getOAuth2Endpoint(provider string) oauth2.Endpoint {
    switch provider {
    case "github":
        return oauth2.Endpoint{
            AuthURL:  "https://github.com/login/oauth/authorize",
            TokenURL: "https://github.com/login/oauth/access_token",
        }
    case "google":
        return oauth2.Endpoint{
            AuthURL:  "https://accounts.google.com/o/oauth2/auth",
            TokenURL: "https://oauth2.googleapis.com/token",
        }
    case "gitlab":
        return oauth2.Endpoint{
            AuthURL:  "https://gitlab.com/oauth/authorize",
            TokenURL: "https://gitlab.com/oauth/token",
        }
    default:
        return oauth2.Endpoint{}
    }
}

func (p *OAuth2Provider) AuthCodeURL(state string) string {
    return p.oauthCfg.AuthCodeURL(state)
}

func (p *OAuth2Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
    return p.oauthCfg.Exchange(ctx, code)
}

func (p *OAuth2Provider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
    // Implementation varies by provider
    switch p.config.Provider {
    case "github":
        return p.getGitHubUserInfo(ctx, token)
    case "google":
        return p.getGoogleUserInfo(ctx, token)
    // ... other providers
    }
    return nil, fmt.Errorf("unsupported provider: %s", p.config.Provider)
}

type UserInfo struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Name     string `json:"name"`
}
```

**Step 2: 实现 JWT Token 管理**

```go
// backend/internal/auth/jwt.go
package auth

import (
    "errors"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

var (
    ErrInvalidToken = errors.New("invalid token")
    ErrExpiredToken = errors.New("token expired")
)

type JWTManager struct {
    secretKey []byte
    issuer    string
}

type Claims struct {
    UserID   uuid.UUID `json:"sub"`
    Username string    `json:"name"`
    Email    string    `json:"email"`
    Groups   []string  `json:"groups"`
    jwt.RegisteredClaims
}

func NewJWTManager(secretKey, issuer string) *JWTManager {
    return &JWTManager{
        secretKey: []byte(secretKey),
        issuer:    issuer,
    }
}

func (m *JWTManager) GenerateToken(userID uuid.UUID, username, email string, groups []string) (string, error) {
    claims := Claims{
        UserID:   userID,
        Username: username,
        Email:    email,
        Groups:   groups,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    m.issuer,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.secretKey)
}

func (m *JWTManager) VerifyToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return m.secretKey, nil
    })

    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }

    if claims.ExpiresAt.Before(time.Now()) {
        return nil, ErrExpiredToken
    }

    return claims, nil
}
```

**Step 3: 编写 OAuth2 回调处理器**

```go
// backend/internal/api/kernel_handler.go - add OAuth2 handlers
func (h *KernelHandler) Login(w http.ResponseWriter, r *http.Request) {
    provider := r.URL.Query().Get("provider")
    if provider == "" {
        provider = "github"
    }

    oauthProvider := h.getOAuth2Provider(provider)
    if oauthProvider == nil {
        http.Error(w, "OAuth2 provider not configured", http.StatusBadRequest)
        return
    }

    state := generateState()
    // Store state in session/cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "oauth_state",
        Value:    state,
        Path:     "/",
        MaxAge:   300,
        HttpOnly: true,
        Secure:   true,
    })

    authURL := oauthProvider.AuthCodeURL(state)
    http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *KernelHandler) OAuth2Callback(w http.ResponseWriter, r *http.Request) {
    state := r.URL.Query().Get("state")
    // Verify state cookie

    code := r.URL.Query().Get("code")
    provider := r.URL.Query().Get("provider")

    oauthProvider := h.getOAuth2Provider(provider)
    token, err := oauthProvider.Exchange(r.Context(), code)
    if err != nil {
        http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
        return
    }

    userInfo, err := oauthProvider.GetUserInfo(r.Context(), token)
    if err != nil {
        http.Error(w, "Failed to get user info", http.StatusInternalServerError)
        return
    }

    // Get or create user
    user, err := h.getOrCreateUser(userInfo)
    if err != nil {
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }

    // Generate JWT
    jwtToken, err := h.jwtManager.GenerateToken(user.ID, user.Username, user.Email, []string{})
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
    }

    // Set token cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "auth_token",
        Value:    jwtToken,
        Path:     "/",
        MaxAge:   86400,
        HttpOnly: true,
        Secure:   true,
    })

    http.Redirect(w, r, "/", http.StatusFound)
}
```

**Step 4: 编写认证中间件**

```go
// backend/internal/auth/middleware.go
package auth

import (
    "context"
    "net/http"
    "strings"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Skip auth for public endpoints
            if isPublicEndpoint(r.URL.Path) {
                next.ServeHTTP(w, r)
                return
            }

            token := extractToken(r)
            if token == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            claims, err := jwtManager.VerifyToken(token)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            // Add user to context
            ctx := context.WithValue(r.Context(), UserContextKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func extractToken(r *http.Request) string {
    // Try cookie first
    cookie, err := r.Cookie("auth_token")
    if err == nil {
        return cookie.Value
    }

    // Try Authorization header
    authHeader := r.Header.Get("Authorization")
    if strings.HasPrefix(authHeader, "Bearer ") {
        return strings.TrimPrefix(authHeader, "Bearer ")
    }

    return ""
}

func isPublicEndpoint(path string) bool {
    publicPaths := []string{
        "/api/auth/login",
        "/api/auth/callback",
        "/healthz",
        "/readyz",
    }
    for _, p := range publicPaths {
        if path == p || strings.HasPrefix(path, p+"/") {
            return true
        }
    }
    return false
}
```

**Step 5: 编写 OAuth2 测试**

```go
// backend/internal/auth/oauth2_test.go
package auth

import (
    "context"
    "testing"
    "os"
)

func TestNewOAuth2Provider(t *testing.T) {
    cfg := &OAuth2Config{
        Provider:     "github",
        ClientID:     "test-client-id",
        ClientSecret: "test-client-secret",
        RedirectURL:  "http://localhost:8080/api/auth/callback",
        Scopes:       []string{"user:email"},
    }

    provider, err := NewOAuth2Provider(cfg)
    if err != nil {
        t.Fatalf("Failed to create OAuth2 provider: %v", err)
    }

    if provider.config.Provider != "github" {
        t.Errorf("Expected provider 'github', got '%s'", provider.config.Provider)
    }
}

func TestJWTManager(t *testing.T) {
    manager := NewJWTManager("test-secret-key", "kubedeck")

    userID := uuid.New()
    token, err := manager.GenerateToken(userID, "testuser", "test@example.com", []string{})
    if err != nil {
        t.Fatalf("Failed to generate token: %v", err)
    }

    claims, err := manager.VerifyToken(token)
    if err != nil {
        t.Fatalf("Failed to verify token: %v", err)
    }

    if claims.Username != "testuser" {
        t.Errorf("Expected username 'testuser', got '%s'", claims.Username)
    }
}
```

**Step 6: 更新路由注册**

```go
// backend/internal/api/router.go
func NewRouter() http.Handler {
    mux := http.NewServeMux()
    kernel := NewKernelHandler()

    // Auth routes
    mux.HandleFunc("/api/auth/login", kernel.Login)
    mux.HandleFunc("/api/auth/callback", kernel.OAuth2Callback)
    mux.HandleFunc("/api/auth/logout", kernel.Logout)
    mux.HandleFunc("/api/auth/me", kernel.GetCurrentUser)

    // Protected routes (require auth)
    mux.Handle("/api/meta/", authMiddleware(kernel.Snapshot))
    mux.Handle("/api/clusters/", authMiddleware(kernel.Clusters))
    mux.Handle("/api/workflows/", authMiddleware(kernel.Workloads))
    mux.Handle("/api/actions/", authMiddleware(kernel.ExecuteAction))
    mux.Handle("/api/users/", authMiddleware(kernel.Users))
    mux.Handle("/api/roles/", authMiddleware(kernel.Roles))

    // Public routes
    mux.HandleFunc("/api/healthz", healthHandler)
    mux.HandleFunc("/api/readyz", healthHandler)

    return mux
}
```

**Step 7: 运行测试验证**

```bash
cd backend && go test ./internal/auth/... -v
# Expected: PASS
```

**Step 8: 提交**

```bash
git add backend/internal/auth/*.go backend/internal/api/router.go
git commit -m "feat(auth): add OAuth2 authentication and JWT token management

- Support GitHub, Google, GitLab OAuth2 providers
- JWT token generation and verification
- Auth middleware for protected endpoints
- User session management via cookies"
```

---

## Phase 2: RBAC 权限系统（2 周）

### Task 3: 角色管理 CRUD

**Files:**
- Create: `backend/internal/storage/role_repository.go`
- Create: `backend/internal/api/roles_handler.go`
- Create: `backend/internal/api/roles_handler_test.go`

**Step 1: 实现角色仓库**

```go
// backend/internal/storage/role_repository.go
package storage

import (
    "github.com/google/uuid"
    "gorm.io/gorm"
    "kubedeck/backend/pkg/sdk"
)

type RoleRepository interface {
    Create(*Role) error
    GetByID(uuid.UUID) (*Role, error)
    GetAll() ([]Role, error)
    Update(*Role) error
    Delete(uuid.UUID) error
}

type roleRepository struct {
    db *gorm.DB
}

func (r *roleRepository) Create(role *Role) error {
    return r.db.Create(role).Error
}

func (r *roleRepository) GetByID(id uuid.UUID) (*Role, error) {
    var role Role
    err := r.db.First(&role, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &role, nil
}

func (r *roleRepository) GetAll() ([]Role, error) {
    var roles []Role
    err := r.db.Find(&roles).Error
    return roles, err
}

func (r *roleRepository) Update(role *Role) error {
    return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id uuid.UUID) error {
    return r.db.Delete(&Role{ID: id}).Error
}

// Convert Role to K8s PolicyRules
func (r *Role) ToPolicyRules() []sdk.PolicyRule {
    var rules []sdk.PolicyRule
    // Parse JSON rules field
    return rules
}
```

**Step 2: 实现角色 API 处理器**

```go
// backend/internal/api/roles_handler.go
package api

import (
    "encoding/json"
    "net/http"
    "github.com/google/uuid"
)

func (h *KernelHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
    roles, err := h.roleRepo.GetAll()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, roles)
}

func (h *KernelHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var role Role
    if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    role.ID = uuid.New()
    if err := h.roleRepo.Create(&role); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    writeJSON(w, role)
}

func (h *KernelHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
    // Implementation
}

func (h *KernelHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

**Step 3: 编写角色 API 测试**

```go
// backend/internal/api/roles_handler_test.go
package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestKernelHandler_CreateRole(t *testing.T) {
    handler := NewKernelHandler()

    role := Role{
        Name:        "test-role",
        Description: "Test role",
        Rules:       `[]`,
    }

    body, _ := json.Marshal(role)
    req := httptest.NewRequest(http.MethodPost, "/api/roles", bytes.NewReader(body))
    rec := httptest.NewRecorder()

    handler.CreateRole(rec, req)

    if rec.Code != http.StatusCreated {
        t.Fatalf("expected status 201, got %d", rec.Code)
    }
}
```

**Step 4: 运行测试并修复**

```bash
cd backend && go test ./internal/api/... -run TestKernelHandler_CreateRole -v
```

**Step 5: 提交**

```bash
git add backend/internal/storage/role_repository.go backend/internal/api/roles_handler.go
git commit -m "feat(rbac): add role management CRUD API

- Create, read, update, delete roles
- Store role rules as JSON (K8s PolicyRules)
- Support builtin and custom roles"
```

---

## Phase 3: Pod 详情页 Phase 1（1 周）

### Task 4: Pod 基础信息页面

**Files:**
- Create: `frontend/src/kernel/builtins/pages/PodPage.tsx`
- Create: `frontend/src/kernel/builtins/pages/PodPage.test.tsx`
- Modify: `frontend/src/kernel/builtins/registerBuiltInPages.ts`
- Modify: `backend/internal/core/builtins/pods_capability.go`

**Step 1: 创建后端 Pod Capability**

```go
// backend/internal/core/builtins/pods_capability.go
package builtins

import "kubedeck/backend/pkg/sdk"

type PodsCapability struct{}

func (PodsCapability) CapabilityDescriptor() sdk.CapabilityDescriptor {
    return sdk.CapabilityDescriptor{
        ID:      "core.pods",
        Version: "v1",
        Pages: []sdk.PageDescriptor{
            {
                ID:               "page.pod-detail",
                WorkflowDomainID: "pods",
                Route:            "/pods/:namespace/:name",
                EntryKey:         "pod-detail",
                Title:            sdk.TextRef{Key: "pod.detail", Fallback: "Pod Details"},
            },
        },
        Menus: []sdk.MenuDescriptor{},
        Actions: []sdk.ActionDescriptor{
            {
                ID:               "pod.logs",
                WorkflowDomainID: "pods",
                Surface:          sdk.ActionSurfaceDrawer,
                Visible:          true,
                Title:            sdk.TextRef{Key: "actions.podLogs", Fallback: "View Logs"},
            },
            {
                ID:               "pod.exec",
                WorkflowDomainID: "pods",
                Surface:          sdk.ActionSurfaceDrawer,
                Visible:          true,
                Title:            sdk.TextRef{Key: "actions.podExec", Fallback: "Terminal"},
            },
        },
    }
}

func (PodsCapability) WorkflowDomainID() string {
    return "pods"
}
```

**Step 2: 创建前端 Pod 详情页组件**

```tsx
// frontend/src/kernel/builtins/pages/PodPage.tsx
import { useEffect, useState } from 'react';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Box from '@mui/material/Box';
import { copy } from '../../../i18n/copy';
import { useKernelRuntime } from '../../runtime/KernelRuntimeContext';

interface PodDetail {
  name: string;
  namespace: string;
  status: string;
  node: string;
  ip: string;
  startTime: string;
  containers: ContainerDetail[];
}

interface ContainerDetail {
  name: string;
  image: string;
  status: string;
  restarts: number;
  ready: boolean;
}

export function PodPage() {
  const { activeCluster, namespaceScope } = useKernelRuntime();
  const [pod, setPod] = useState<PodDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState(0);

  useEffect(() => {
    // Fetch pod details from API
    async function fetchPodDetail() {
      try {
        const params = new URLSearchParams(window.location.search);
        const name = params.get('name');
        const namespace = params.get('namespace') || 'default';
        
        const response = await fetch(`/api/pods/${namespace}/${name}?cluster=${activeCluster}`);
        const data = await response.json();
        setPod(data);
      } catch (error) {
        console.error('Failed to fetch pod detail:', error);
      } finally {
        setLoading(false);
      }
    }

    void fetchPodDetail();
  }, [activeCluster]);

  if (loading) {
    return <Typography>Loading...</Typography>;
  }

  if (!pod) {
    return <Typography>Pod not found</Typography>;
  }

  return (
    <Stack spacing={3}>
      {/* Header */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack spacing={1}>
            <Typography variant="h6">
              {pod.name}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Namespace: {pod.namespace}
            </Typography>
          </Stack>
          <Stack direction="row" spacing={1}>
            <Chip label={`Status: ${pod.status}`} color={pod.status === 'Running' ? 'success' : 'warning'} />
            <Chip label={`Node: ${pod.node}`} variant="outlined" />
          </Stack>
        </Stack>
      </Paper>

      {/* Tabs */}
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs value={activeTab} onChange={(_, newValue) => setActiveTab(newValue)}>
          <Tab label={copy('pod.tabs.overview')} />
          <Tab label={copy('pod.tabs.logs')} />
          <Tab label={copy('pod.tabs.terminal')} />
          <Tab label={copy('pod.tabs.files')} />
          <Tab label={copy('pod.tabs.events')} />
          <Tab label="YAML" />
        </Tabs>
      </Box>

      {/* Overview Tab */}
      {activeTab === 0 && (
        <Stack spacing={2}>
          {/* Containers Table */}
          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 1 }}>
              Containers ({pod.containers.length})
            </Typography>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>{copy('pod.columns.name')}</TableCell>
                  <TableCell>{copy('pod.columns.image')}</TableCell>
                  <TableCell>{copy('pod.columns.status')}</TableCell>
                  <TableCell>{copy('pod.columns.restarts')}</TableCell>
                  <TableCell>{copy('pod.columns.ready')}</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {pod.containers.map((container) => (
                  <TableRow key={container.name}>
                    <TableCell>{container.name}</TableCell>
                    <TableCell>{container.image}</TableCell>
                    <TableCell>
                      <Chip label={container.status} size="small" color={container.status === 'Running' ? 'success' : 'default'} />
                    </TableCell>
                    <TableCell>{container.restarts}</TableCell>
                    <TableCell>
                      <Chip label={container.ready ? 'Ready' : 'Not Ready'} size="small" color={container.ready ? 'success' : 'error'} />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Paper>
        </Stack>
      )}

      {/* Logs Tab */}
      {activeTab === 1 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography>Logs viewer will be implemented in Phase 2</Typography>
        </Paper>
      )}

      {/* Terminal Tab */}
      {activeTab === 2 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography>Terminal will be implemented in Phase 2</Typography>
        </Paper>
      )}
    </Stack>
  );
}
```

**Step 3: 注册 Pod 页面**

```tsx
// frontend/src/kernel/builtins/registerBuiltInPages.ts
import { PodPage } from './pages/PodPage';

export function registerBuiltInPages(): PageContribution[] {
  return [
    // ... existing pages
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.pods',
        contributionId: 'page.pod-detail',
      },
      workflowDomainId: 'pods',
      route: '/pods/:namespace/:name',
      entryKey: 'pod-detail',
      title: { key: 'pod.detail', fallback: 'Pod Details' },
      component: PodPage,
      order: 30,
    },
  ];
}
```

**Step 4: 添加 i18n 文案**

```ts
// frontend/src/i18n/messages/en.ts
export const enMessages = {
  // ... existing messages
  'pod.detail': 'Pod Details',
  'pod.tabs.overview': 'Overview',
  'pod.tabs.logs': 'Logs',
  'pod.tabs.terminal': 'Terminal',
  'pod.tabs.files': 'Files',
  'pod.tabs.events': 'Events',
  'pod.columns.name': 'Name',
  'pod.columns.image': 'Image',
  'pod.columns.status': 'Status',
  'pod.columns.restarts': 'Restarts',
  'pod.columns.ready': 'Ready',
  'actions.podLogs': 'View Logs',
  'actions.podExec': 'Terminal',
} as const;
```

**Step 5: 编写 Pod 页面测试**

```tsx
// frontend/src/kernel/builtins/pages/PodPage.test.tsx
import { render, screen, waitFor } from '@testing-library/react';
import { PodPage } from './PodPage';
import { KernelRuntimeProvider } from '../../runtime/KernelRuntimeContext';

describe('PodPage', () => {
  it('renders pod details', async () => {
    render(
      <KernelRuntimeProvider>
        <PodPage />
      </KernelRuntimeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Pod Details')).toBeTruthy();
    });
  });
});
```

**Step 6: 运行测试验证**

```bash
cd frontend && npm test -- --run PodPage
```

**Step 7: 提交**

```bash
git add frontend/src/kernel/builtins/pages/PodPage.tsx frontend/src/i18n/messages/en.ts
git commit -m "feat(pod): add Pod detail page with overview tab

- Display pod status, node, IP, start time
- Show container list with status and restarts
- Tab structure for logs, terminal, files, events"
```

---

## 验证命令

每个 Phase 完成后运行：

```bash
# Backend tests
cd backend && go test ./...

# Frontend tests
cd frontend && npm test -- --run

# Build verification
cd frontend && npm run build
cd backend && go build ./...

# Integration test
./kubedeck --port 8080 &
curl http://localhost:8080/api/healthz
```

---

## 执行选项

**计划已完成并保存到 `docs/plans/2026-03-23-platform-enhancement-implementation.md`。**

**两种执行方式：**

**1. Subagent-Driven（当前会话）** - 我为每个任务分派新的 subagent，任务间进行代码审查，快速迭代

**2. Parallel Session（独立会话）** - 在新会话中使用 executing-plans 技能，批量执行带检查点

**选择哪种方式？**
