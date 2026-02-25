# Tech Plan Delta

## Design
- Replace non-strict persistence calls in write handlers with strict variants:
  - `persistIAMGroupsStrict`
  - `persistIAMMembershipsStrict`
  - `persistIAMInvitesStrict`
- On error, return `persistence_failed` and stop handler execution before audit success writes.

## Files
- `backend/internal/api/iam_handler.go`
- `backend/internal/api/meta_handler_test.go`
