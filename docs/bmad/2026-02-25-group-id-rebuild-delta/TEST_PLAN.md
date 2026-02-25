# Test Plan Delta

- `TestIAMGroupSameNameAllowedAcrossTenants`
- `TestIAMGroupPatchRejectsDuplicateNameInTenant`
- `TestGroupIDRebuildRewritesMembershipRefs`
- `TestGroupIDRebuildRejectsCanonicalNameConflict`
- Full regression: `cd backend && go test ./... && go build ./...`
