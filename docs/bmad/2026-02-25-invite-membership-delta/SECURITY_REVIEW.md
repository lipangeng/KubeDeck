# Security Review: Invite Acceptance Membership Delta

## AuthN/AuthZ
- Invite token validation unchanged.
- Membership write scoped to invite tenant only.

## Data Handling
- Membership creation does not persist raw invite token.

## Residual Risk
- Username collisions across auth providers remain a known limitation for MVP.
