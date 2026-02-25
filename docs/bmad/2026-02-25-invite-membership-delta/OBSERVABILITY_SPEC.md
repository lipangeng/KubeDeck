# Observability Spec: Invite Acceptance Membership Delta

## Logging
- Existing audit event `auth.accept_invite` remains the primary signal.

## Metrics (future)
- `invite_accept_total{result}`
- `invite_accept_membership_upsert_total{role_hint}`
