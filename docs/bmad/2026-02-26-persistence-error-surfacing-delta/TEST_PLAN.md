# Test Plan Delta

- `TestIAMRebuildGroupIDsEndpointReturns500OnPersistenceFailure`
- `TestIAMRebuildGroupIDsEndpoint`
- `TestIAMRebuildGroupIDsEndpointRejectsCanonicalConflict`
- Full regression: `cd backend && go test ./... && go build ./...`
