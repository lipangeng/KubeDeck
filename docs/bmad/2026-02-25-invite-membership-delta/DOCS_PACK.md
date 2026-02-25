# Docs Pack: Invite Acceptance Membership Delta

## Developer Notes
- Updated backend auth/iam flow in `auth_handler.go` and local id helper in `provider_local.go`.

## API Example
- `POST /api/auth/accept-invite` now returns `membership` in success payload.
