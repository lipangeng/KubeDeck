# Architecture Spec: Invite -> Membership Link

## System Overview
- Keep invite state in IAM invite store.
- On accept, atomically mark invite accepted, then upsert IAM membership.

## Key Decisions
- Use stable local user id derivation from username for membership user_id.
- Reuse existing tenant default groups for role mapping.

## NFR
- Security: no raw invite token persisted in membership path.
- Operability: no new service dependency.
