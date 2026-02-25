# PRD Delta - IAM Write Persistence Strict

## Scope
- Ensure IAM write APIs fail fast when persistence initialization/write cannot proceed.
- Target endpoints: group create/patch/delete/permissions, tenant member create/delete, membership group/validity replace, invite create/revoke.

## Acceptance Criteria
- On persistence init/write failure, endpoint returns `500` with error code `persistence_failed`.
- No success response is emitted when persistence fails.
- Existing success behavior remains unchanged.
