# Test Plan Delta

- `TestAuthLoginRejectsInvalidTenantCode`
- `TestSwitchTenantRejectsInvalidTenantCode`
- `TestAuthLoginWritesAttemptedAndOutcomeAuditEvents`
- Full regression: `cd backend && go test ./...`
