# API Contract Delta

## POST /api/iam/groups/rebuild-ids
Auth: requires `iam:write`

Success (200):
- `{ "status": "ok" }`

Errors:
- 409: `group_name_conflict`
- 403: `permission_denied`
- 500: `rebuild_failed`
