# API Contract: Auth Invite Acceptance Delta

## Endpoint
### POST /api/auth/accept-invite
Request body:
- `token` string (required)
- `username` string (required)
- `password` string (required)

Response 200:
- `status`, `tenant_id`, `username`, `membership`

Errors:
- 400: `token_required`, `username_password_required`
- 404/409/410 as existing invite lifecycle errors
