# Test Plan: Invite Acceptance Membership Delta

## Functional
- Accept invite returns `membership` object.
- Membership user id equals `auth.LocalUserID(username)`.

## Permissions
- `role_hint=member` maps to `grp-<tenant>-viewer`.
- `role_hint=admin` maps to `grp-<tenant>-admin`.

## Smoke
- `cd backend && go test ./...`
