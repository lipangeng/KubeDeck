# API Contract: <Service Name>

## Auth
- mechanism:
- required scopes/roles:

## Common Models
- Error: { code, message, details? }

## Endpoints

### GET /...
Request:
- query:
Response:
- 200:
Errors:
- 401/403/404/500:

### POST /...
Request:
- body:
Response:
- 200/201:
Errors:
- 400/401/403/409/500:

## Notes
- pagination:
- idempotency:
- rate limits:
