# Group ID Full Rebuild Design (Maintenance Window)

## Context
KubeDeck IAM currently allows group IDs derived from mutable group names. This can cause ID drift and collision risks across tenants when names are changed or normalized inconsistently. We will rebuild IDs to a tenant-scoped, hash-based immutable scheme.

## Goals
- Rebuild all IAM group IDs to immutable format: `grp-<tenantID>-<hash8(canonical-name)>`.
- Rewrite all membership `group_ids` references from old IDs to new IDs.
- Keep group semantic fields unchanged: `name`, `description`, `permissions`.
- Complete in a short maintenance window with clear rollback.

## Out of Scope
- No dual-write transition.
- No old-ID compatibility lookup.
- No frontend protocol change beyond returned ID values.

## Chosen Approach
- Option selected: **full rebuild in maintenance window**.
- Rationale: cleanest long-term state, no transitional complexity, deterministic migration.

## Migration Flow
1. Enter maintenance window
- Freeze IAM write operations (group/membership/invite mutations).

2. Pre-check
- Build mapping on `tenant_id + canonical(name)`.
- If canonical-name conflicts exist in same tenant, abort with report.

3. Generate new IDs
- `canonical(name) = trim + lowercase + whitespace collapse`.
- `new_id = grp-<tenantID>-<hash8(canonical(name))>`.

4. Full rebuild
- Rebuild `iam_groups` with new IDs.
- Rewrite `iam_memberships.group_ids` using old->new mapping.

5. Verification
- No dangling membership group IDs.
- IAM group list/create/patch and membership group replacement pass smoke checks.

6. Exit / rollback
- Keep pre-migration snapshot (groups + memberships).
- On any failure, restore snapshot and exit maintenance mode.

## Data Safety and Rollback
- Snapshot must be taken before mutation.
- Migration is all-or-nothing at repository level.
- Rollback path is deterministic: restore snapshot, re-enable writes.

## Test Strategy
- Unit
  - canonical-name normalization.
  - ID generation determinism and tenant scoping.
  - old->new mapping and rewrite behavior.
- Integration
  - group CRUD and membership group replacement after rebuild.
- Regression
  - invite acceptance membership mapping remains valid.
  - permission resolution remains correct.

## Acceptance Criteria
- Same-tenant canonical duplicate names are detected and migration aborts safely.
- All membership `group_ids` reference existing groups after rebuild.
- IAM write paths recover after maintenance window.
- Backend tests and build pass.

## Deliverables
- Design doc (this file).
- Implementation plan doc.
- Migration code + rollback notes.
- BMAD artifacts update (PRD/API/TECH/TEST/SEC/OBS/RELEASE).
