# Tech Plan: Invite Acceptance Membership Delta

## BE Plan
- Export `auth.LocalUserID(username)` for consistent local identity mapping.
- Add `upsertMembershipForAcceptedInvite` in auth handler.
- Add `inviteRoleGroupIDs` mapping helper.
- Extend tests for membership creation and role mapping.

## Definition of Done
- [x] AC covered by tests
- [x] Minimal smoke pass
- [x] Security review delta
- [x] Observability delta
