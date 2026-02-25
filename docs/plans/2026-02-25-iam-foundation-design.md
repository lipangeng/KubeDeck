# KubeDeck IAM Foundation Design (Phase 1)

## 1. Goal
Build minimum-but-real IAM foundation first, before more platform extensions:
- local username/password login
- multi-tenant membership and tenant switching
- group-based RBAC (backend authoritative)
- invite onboarding (email first, SMS interface reserved)
- structured audit trail

## 2. Scope
Included in Phase 1:
- Auth/session APIs: login, me, logout, switch tenant
- IAM domain: users, tenants, memberships, groups, permissions, bindings
- Invite domain: create/list/accept invite, invite-link generation
- Audit events for auth/IAM/resource-sensitive operations
- Storage-first abstraction compatible with SQLite now, MySQL/PostgreSQL later

Deferred:
- OAuth/OIDC provider integration
- ABAC expression engine
- advanced audit UI analytics

## 3. Core Constraints
- Backend authorization is authoritative.
- Single account can join multiple tenants.
- Login supports `tenant_code` for direct tenant targeting.
- Membership has validity window; expired membership is denied for login/switch/access.
- Audit records are mandatory for critical flows; write failure does not block business response.

## 4. Data Model
- `tenants(id, code, name, status, created_at, updated_at)`
- `users(id, username, email, phone, password_hash, status, created_at, updated_at)`
- `tenant_memberships(id, tenant_id, user_id, membership_role, status, effective_from, effective_to)`
- `groups(id, tenant_id, name, description, created_at, updated_at)`
- `permissions(id, code, description, scope)`
- `group_permissions(id, tenant_id, group_id, permission_id)`
- `membership_groups(id, tenant_id, membership_id, group_id)`
- `sessions(id, user_id, active_tenant_id, token_hash, expires_at, last_seen_at, ip, user_agent)`
- `invites(id, tenant_id, invitee_email, invitee_phone, role_hint, token_hash, expires_at, status, inviter_user_id, created_at)`
- `audit_events(id, tenant_id, actor_user_id, action, target_type, target_id, result, reason, metadata_json, created_at)`

Key uniqueness:
- `tenants.code` unique
- `tenant_memberships(tenant_id,user_id)` unique
- `groups(tenant_id,name)` unique
- `group_permissions(tenant_id,group_id,permission_id)` unique
- `membership_groups(tenant_id,membership_id,group_id)` unique

## 5. API Contract (MVP)
Auth:
- `POST /api/auth/login` (supports `tenant_code`)
- `GET /api/auth/me`
- `POST /api/auth/switch-tenant`
- `POST /api/auth/logout`

IAM:
- `GET /api/iam/permissions`
- `GET/POST/PATCH/DELETE /api/iam/groups`
- `PUT /api/iam/groups/:id/permissions`
- `PUT /api/iam/memberships/:id/groups`
- `GET /api/iam/tenants`
- `GET/POST/DELETE /api/iam/tenants/:id/members`

Invite:
- `POST /api/iam/invites`
- `GET /api/iam/invites`
- `POST /api/auth/accept-invite`

Audit:
- `GET /api/audit/events`

## 6. RBAC Baseline
Permission codes:
- `iam:read`, `iam:write`
- `tenant:read`, `tenant:write`
- `audit:read`
- `menu:read`, `menu:write`
- `resource:read`, `resource:write`, `resource:apply`, `resource:delete`
- `cluster:switch`

Default tenant groups:
- `tenant-owner`
- `tenant-admin`
- `tenant-viewer`

## 7. Membership Validity
Membership is valid when:
- `now >= effective_from`
- and (`effective_to is null` or `now < effective_to`)

Denied flows if invalid:
- login with tenant targeting
- switch tenant
- protected API authorization checks

## 8. Invite + Notification Model
- Invite can target email or phone (at least one required).
- Email sending is implemented in Phase 1.
- SMS provider is interface-only in Phase 1.
- Admin can copy invite link and share manually.

Notification provider abstraction:
- `SendEmail(to, subject, body)`
- `SendSMS(phone, content)`

## 9. Audit Requirements
Must audit:
- auth login success/failure
- logout
- tenant switch success/failure
- invite create/send/accept/expire/revoke
- IAM mutation operations
- sensitive resource mutation operations

Required fields:
- tenant, actor, action, target, result, timestamp

## 10. Delivery Order
- M1: Auth + session + tenant context + membership validity + core audit
- M2: RBAC group-permission-member binding + protected APIs
- M3: Invite onboarding + email provider + invite audit chain
