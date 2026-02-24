# KubeDeck Architecture Design (Phase 1-3)

Date: 2026-02-25
Status: Approved (Brainstorming output)
Source baseline: `PLAN.md` (SSOT)

## 1. System Blueprint

Adopt a contract-first, layered, incremental strategy.

- Backend Core (microkernel) provides framework capabilities only:
  - Plugin lifecycle/runtime
  - Registry
  - Cluster manager
  - Unified resource access layer
  - Auth framework
  - Metadata APIs
- Frontend Shell (microkernel) provides container capabilities only:
  - Plugin registration host
  - Routing and replacement resolution
  - Sidebar composition (system/user/dynamic)
  - Global state (cluster/namespace/theme/i18n/permission hints)
  - Slot renderer
- Business features are all plugins, including built-in features.
- Phase mapping:
  - Phase 1: skeleton + minimal registry loop + single cluster + basic dynamic CRD menu + multi-YAML create
  - Phase 2: multi-cluster context and per-cluster registry + Pod replacement + slot injection
  - Phase 3: global search + resource relationship graph + terminal + security governance + OAuth management UI

## 2. Directory and Module Responsibilities

```txt
KubeDeck/
  docs/
    PLAN.md
    plans/
  frontend/
    shell/
      src/
        core/            # plugin host, registry client, slot renderer
        app/             # layout/sidebar/router
        state/           # cluster/ns/theme/i18n/permission hints
        sdk/             # registerPages/registerExtensions/...
        plugins/         # built-in frontend plugins
    packages/
      plugin-types/      # shared TS types (manifest/meta dto)
  backend/
    cmd/kubedeck/
    internal/
      core/              # plugin runtime, service wiring
      registry/          # registry model + compose pipeline
      cluster/           # kubeconfig discovery + client cache + discovery cache
      resource/          # typed/dynamic unified accessor
      auth/              # local auth + oauth provider interface + rbac
      storage/           # storage abstraction and drivers (sqlite/mysql/postgres)
      api/               # http handlers (meta/menu/resource/apply/logs)
      plugins/           # built-in backend plugins
    pkg/sdk/             # backend plugin sdk contract
  plugins/
    templates/
      frontend-plugin-template/
      backend-plugin-template/
```

High-risk module ownership details:

- `backend/internal/cluster`
  - `manager.go`: cluster list/switching and client/discovery cache lifecycle
  - `discovery_sync.go`: per-cluster API groups/resources/CRD refresh
- `backend/internal/registry`
  - `types.go`: ResourceType/Page/Slot/Menu model
  - `builder.go`: compose core + plugins + dynamic CRD into registry snapshot
- `backend/internal/auth`
  - `provider_local.go`: local auth
  - `provider_oauth.go`: OAuth provider interface adapter (stub allowed in MVP)
  - `rbac_eval.go`: platform RBAC + cluster/ns allowlist evaluation
- `backend/internal/storage`
  - `repo_interfaces.go`: repository contracts (`UserMenuRepo`, `UserPreferenceRepo`, `PluginConfigRepo`, future `RBACRepo`)
  - `driver_sqlite.go`: default embedded sqlite implementation
  - `driver_mysql.go`: optional mysql implementation
  - `driver_postgres.go`: optional postgresql implementation
  - `migrations/*`: dialect-specific migration scripts with unified versioning
- `frontend/shell/src/core`
  - `pluginHost.ts`: plugin registration lifecycle
  - `routeResolver.ts`: replacement priority resolution
  - `menuComposer.ts`: merge/sort/visibility for system/user/dynamic menus
- `frontend/shell/src/state`
  - `clusterContext.ts`: clear cluster-bound cache and reload registry/menu on switch
  - `namespaceFilter.ts`: namespace defaulting rules linked with create dialog
- `plugins/templates/*`
  - mandatory `plugin.manifest.json`, `README.md`, `src/index` placeholder

## 3. Data Models and API Contracts

### 3.1 Registry models (backend authority)

```ts
type ResourceType = {
  id: string;
  group: string;
  version: string;
  kind: string;
  plural: string;
  namespaced: boolean;
  preferredVersion: string;
  source: "system" | "plugin" | "dynamic-crd";
};

type PageMeta = {
  pageId: string;
  route: string;
  pluginId: string;
  replacementFor?: string;
  slots?: string[];
};

type SlotMeta = {
  slotId: string;
  pageId: string;
  accepts: "tab" | "panel" | "action" | "widget";
  ordering: "weight" | "append";
};

type MenuItem = {
  id: string;
  group: string;
  titleI18nKey: string;
  targetType: "page" | "resource";
  targetRef: string;
  source: "system" | "user" | "dynamic";
  order: number;
  visible: boolean;
  permissionHints?: string[];
};
```

### 3.2 Metadata APIs

- `GET /api/meta/registry?cluster=<id>`
  - returns `resourceTypes/pages/slots/replacements`
- `GET /api/meta/menus?cluster=<id>`
  - returns menu model (recommended: merged output with source annotations)
- `GET /api/meta/clusters`
  - returns accessible cluster list + default cluster
- `GET /api/user/preferences`
- `PUT /api/user/preferences`

### 3.3 Resource APIs (MVP required)

- `POST /api/resources/apply?cluster=<id>&defaultNs=<ns>`
  - body: multi-document YAML
  - response: per-document results (`success/error/reason/ref`)
- `GET /api/resources/list`
- `GET /api/resources/get`
- `DELETE /api/resources/delete`
- `GET /api/logs/pod`

### 3.4 Auth contracts (dual-track)

- `POST /api/auth/login` (local)
- `GET /api/auth/me`
- `POST /api/auth/logout`
- reserved OAuth endpoints:
  - `GET /api/auth/oauth/providers`
  - `GET /api/auth/oauth/callback/<provider>`

Rule: backend authorization is authoritative; UI permission hints are display-only.

### 3.5 Storage contracts (multi-database)

- config:
  - `storage.driver=sqlite|mysql|postgres`
  - `storage.dsn=<driver-specific-dsn>`
- default: embedded sqlite
- drivers are swappable without business-layer code changes
- unified repository interfaces and migration version semantics

## 4. Critical Flows

### 4.1 Cluster switching

1. Frontend triggers `setCluster(newClusterId)`.
2. Clear old-cluster-bound resource/page caches.
3. Fetch registry and menus for new cluster.
4. Restore namespace filter for that cluster; fallback to `default`.
5. Re-resolve routing for replacement/slots.
6. All subsequent API calls carry new cluster context.

Constraint: no cross-cluster state reuse.

### 4.2 Dynamic CRD menu generation

Backend per cluster:

1. Run discovery to fetch CRD list and preferred version.
2. Register CRDs as `ResourceType(source=dynamic-crd)`.
3. Bind generic resource pages when no replacement exists.
4. Generate dynamic menu items (default grouping by CRD group).
5. Apply visibility policies (allow/deny/default rules; MVP can be minimal).
6. Return menus for frontend composition.

Frontend only merges/renders system + user + dynamic.

### 4.3 Multi-YAML create (`---`)

1. Frontend submits raw YAML + `defaultNs`.
2. Backend splits YAML into ordered documents.
3. Parse each doc GVK and scope.
4. For namespaced resources, fill `metadata.namespace` if missing.
5. Apply each document independently.
6. Return summary (`total/succeeded/failed/partial`) + per-doc result.

Failure policy: no automatic rollback; partial failure is explicit.

## 5. Error Handling, Observability, Testing, Acceptance

### 5.1 Error handling

- Unified API error shape: `code/message/details/requestId`.
- Multi-YAML returns request-level + doc-level status.
- Auth failures use consistent `401/403` with auditable error codes.
- Distinguish cluster-unreachable, discovery-failure, CRD-parse-failure.

### 5.2 Platform observability

- Structured logs include `requestId/user/cluster/plugin`.
- Metrics baseline: request latency/error rate, discovery refresh latency, registry build latency, apply success rate.
- Audit logs: resource mutations, login/logout, authorization denials, plugin registration failures.
- Probes:
  - `/healthz` (process)
  - `/readyz` (dependencies ready: storage/cluster manager)

### 5.3 Test strategy

- Unit tests:
  - registry composition rules
  - cluster switch cache cleanup
  - multi-YAML namespace defaulting and per-doc results
  - auth/rbac matrix
  - storage contract parity across sqlite/mysql/postgres drivers
- Integration tests:
  - per-cluster registry/menu differences
  - dynamic CRD menu updates on cluster switch
  - replacement/slot precedence behavior
- Minimal E2E for MVP:
  - switch cluster -> menu refresh -> open resource page -> multi-YAML create -> inspect per-doc feedback

### 5.4 Acceptance criteria by phase

- Phase 1: skeleton, contracts, templates, and single-cluster dynamic-resource loop are runnable.
- Phase 2: multi-cluster isolation and per-cluster registry work; replacement/slots extensibility works.
- Phase 3: search/graph/terminal/OAuth management delivered without breaking plugin contracts.

## 6. Selected Decisions (for traceability)

- Scope: full Phase 1-3 long-term architecture design.
- Detail level: high-level overall + implementation-level detail on high-risk modules.
- Auth strategy: MVP local auth + OAuth-ready interfaces (dual-track).
- Plugin loading start point: compile-time built-in plugins only.
- Data storage: multi-database support from architecture start; default sqlite, switchable to mysql/postgresql.
