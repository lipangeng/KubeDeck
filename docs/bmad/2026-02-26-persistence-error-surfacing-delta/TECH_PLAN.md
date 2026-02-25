# Tech Plan Delta

- Added strict persist helpers in `iam_persistence.go`:
  - `persistIAMGroupsStrict`
  - `persistIAMMembershipsStrict`
  - `persistIAMInvitesStrict`
  - `persistAuthSessionsStrict`
- Added `persistenceInitError` helper to surface init failure when persistence is enabled.
- Rebuild flow now uses strict helpers and restores snapshots on persist failure.
