# PRD: Group ID Rebuild API Delta

## Goals
- Provide controlled API entry to trigger group-id rebuild.
- Return explicit conflict error when canonical-name collision exists.

## Acceptance Criteria
- AC1: `POST /api/iam/groups/rebuild-ids` returns 200 on successful rebuild.
- AC2: canonical conflict returns 409 `group_name_conflict`.
- AC3: audit event `iam.group.rebuild_ids` records success/denied outcomes.
