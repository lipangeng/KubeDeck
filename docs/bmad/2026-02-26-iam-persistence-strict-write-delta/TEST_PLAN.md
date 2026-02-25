# Test Plan Delta

## New Tests
- `TestIAMCreateGroupReturns500OnPersistenceFailure`
- `TestIAMCreateTenantMemberReturns500OnPersistenceFailure`
- `TestIAMCreateInviteReturns500OnPersistenceFailure`

## Validation Commands
- `cd backend && go test ./...`
- `cd backend && go build ./...`
