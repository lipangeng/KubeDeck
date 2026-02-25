# PRD: Auth Audit + Tenant Code Validation Delta

## Goals
- P0: Reject malformed `tenant_code` in login/switch/oauth callback.
- P0: Standardize login audit lifecycle: attempted, failed, succeeded.

## Acceptance Criteria
- AC1: malformed tenant_code returns 400 `invalid_tenant_code`.
- AC2: login failure writes `auth.login.failed`, success writes `auth.login.succeeded`.
- AC3: each login request writes `auth.login.attempted`.
