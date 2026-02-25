# Group ID Full Rebuild Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rebuild IAM group IDs to tenant-scoped hash format and rewrite membership references safely within a maintenance-window style migration.

**Architecture:** Implement deterministic canonicalization and ID generation in IAM handler helpers, add an explicit rebuild function that rewrites in-memory groups/memberships atomically, then persist. Keep rollback by snapshotting current maps before mutation and restoring on persistence failure. Validate via focused API tests and full backend regression.

**Tech Stack:** Go (`backend/internal/api`), existing in-memory + persistence repo model, Go test framework.

---

### Task 1: Add Failing Tests For Tenant-Scoped Group IDs

**Files:**
- Modify: `backend/internal/api/meta_handler_test.go`
- Test: `backend/internal/api/meta_handler_test.go`

**Step 1: Write failing test for cross-tenant same-name groups**

```go
func TestIAMGroupSameNameAllowedAcrossTenants(t *testing.T) {
    // seed admin memberships for tenant-dev and tenant-qa
    // create "ops" in dev, switch tenant, create "ops" in qa
    // assert both created and IDs differ
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/api -run TestIAMGroupSameNameAllowedAcrossTenants -v`
Expected: FAIL (duplicate conflict or same ID)

**Step 3: Commit test-only change**

```bash
git add backend/internal/api/meta_handler_test.go
git commit -m "test(iam): add cross-tenant same-name group id coverage"
```

### Task 2: Add Failing Test For Membership Group Replacement Using Returned IDs

**Files:**
- Modify: `backend/internal/api/meta_handler_test.go`

**Step 1: Update membership replace flow to use created group IDs dynamically**

```go
// unmarshal create group responses into iamGroup
replacePayload := []byte(`{"group_ids":["` + adminsGroup.ID + `","` + devopsGroup.ID + `"]}`)
```

**Step 2: Run test to verify current implementation mismatch**

Run: `cd backend && go test ./internal/api -run TestMembershipListAndReplaceGroupsFlow -v`
Expected: FAIL before ID logic fix

**Step 3: Commit test adjustment**

```bash
git add backend/internal/api/meta_handler_test.go
git commit -m "test(iam): use returned group ids in membership replace flow"
```

### Task 3: Implement Tenant-Scoped Stable Group ID Generator

**Files:**
- Modify: `backend/internal/api/iam_handler.go`
- Test: `backend/internal/api/meta_handler_test.go`

**Step 1: Add canonical + hash helper**

```go
func canonicalGroupName(name string) string { /* trim/lower/space collapse */ }
func shortGroupHash(canonical string) string { /* sha256 -> first 8 hex */ }
func groupIDForTenantName(tenantID, groupName string) string {
    return "grp-" + tenantID + "-" + shortGroupHash(canonicalGroupName(groupName))
}
```

**Step 2: Replace createGroup ID construction**

```go
id := groupIDForTenantName(session.ActiveTenantID, normalizedName)
```

**Step 3: Run focused tests**

Run:
- `cd backend && go test ./internal/api -run TestIAMGroupSameNameAllowedAcrossTenants -v`
- `cd backend && go test ./internal/api -run TestMembershipListAndReplaceGroupsFlow -v`
Expected: PASS

**Step 4: Commit implementation**

```bash
git add backend/internal/api/iam_handler.go backend/internal/api/meta_handler_test.go
git commit -m "feat(iam): use tenant-scoped hash group ids"
```

### Task 4: Enforce Patch Name Uniqueness Within Tenant

**Files:**
- Modify: `backend/internal/api/iam_handler.go`
- Modify: `backend/internal/api/meta_handler_test.go`

**Step 1: Add helper excluding current group ID**

```go
func groupNameExistsByTenantExcept(tenantID, groupName, excludeID string) bool { ... }
```

**Step 2: Validate patch name**

```go
if groupNameExistsByTenantExcept(session.ActiveTenantID, nextName, groupID) {
    writeJSONError(w, http.StatusConflict, "group_exists")
    return
}
```

**Step 3: Add/adjust test for patch duplicate conflict**

```go
func TestIAMGroupPatchRejectsDuplicateNameInTenant(t *testing.T) { ... }
```

**Step 4: Run focused tests**

Run: `cd backend && go test ./internal/api -run TestIAMGroup -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/api/iam_handler.go backend/internal/api/meta_handler_test.go
git commit -m "feat(iam): enforce per-tenant name uniqueness on group patch"
```

### Task 5: Implement Full Rebuild Function (Maintenance Mode Operation)

**Files:**
- Modify: `backend/internal/api/iam_handler.go`
- Modify: `backend/internal/api/iam_persistence.go`
- Modify: `backend/internal/api/meta_handler_test.go`

**Step 1: Add rebuild function with snapshot rollback**

```go
func rebuildGroupIDsForAllTenants() error {
    // snapshot groups/memberships
    // build old->new map
    // detect canonical conflicts per tenant
    // rewrite groups and membership group_ids
    // persist groups then memberships
    // on persist error: restore snapshot and return error
}
```

**Step 2: Wire execution path (internal call in tests first)**

```go
if err := rebuildGroupIDsForAllTenants(); err != nil { ... }
```

**Step 3: Add migration behavior test**

```go
func TestGroupIDRebuildRewritesMembershipRefs(t *testing.T) { ... }
```

**Step 4: Run focused migration tests**

Run: `cd backend && go test ./internal/api -run TestGroupIDRebuild -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/api/iam_handler.go backend/internal/api/iam_persistence.go backend/internal/api/meta_handler_test.go
git commit -m "feat(iam): add full group-id rebuild with membership rewrite"
```

### Task 6: BMAD Delta Docs For Rebuild Iteration

**Files:**
- Create: `docs/bmad/2026-02-25-group-id-rebuild-delta/PRD.md`
- Create: `docs/bmad/2026-02-25-group-id-rebuild-delta/API_CONTRACT.md`
- Create: `docs/bmad/2026-02-25-group-id-rebuild-delta/TECH_PLAN.md`
- Create: `docs/bmad/2026-02-25-group-id-rebuild-delta/TEST_PLAN.md`
- Create: `docs/bmad/2026-02-25-group-id-rebuild-delta/SECURITY_REVIEW.md`
- Create: `docs/bmad/2026-02-25-group-id-rebuild-delta/OBSERVABILITY_SPEC.md`
- Create: `docs/bmad/2026-02-25-group-id-rebuild-delta/RELEASE_NOTES.md`

**Step 1: Write concise delta artifacts**

```markdown
Include: scope, migration steps, rollback, test evidence, audit/ops notes.
```

**Step 2: Commit docs**

```bash
git add docs/bmad/2026-02-25-group-id-rebuild-delta
git commit -m "docs(bmad): add group-id rebuild iteration artifacts"
```

### Task 7: Full Verification And Push

**Files:**
- Modify: none (verification)

**Step 1: Run full backend verification**

Run:
- `cd backend && go test ./...`
- `cd backend && go build ./...`
Expected: all PASS

**Step 2: Inspect git status/log**

Run:
- `git status --short`
- `git log --oneline -n 10`
Expected: clean tree, expected commits present

**Step 3: Push branch**

Run: `git push`
Expected: push success to `temp/merge-validation`
