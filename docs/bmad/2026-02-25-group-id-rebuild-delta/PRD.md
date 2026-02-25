# PRD: Group ID Tenant-Scoped Rebuild Delta

## Goals
- P0: Make IAM group IDs tenant-scoped and stable via canonical-name hash.
- P0: Provide rebuild operation to rewrite membership group references.

## Acceptance Criteria
- AC1: Same group name can exist across tenants without ID conflict.
- AC2: Rebuild rewrites membership `group_ids` from old IDs to new IDs.
- AC3: Canonical-name conflicts within one tenant are detected and migration aborts safely.
