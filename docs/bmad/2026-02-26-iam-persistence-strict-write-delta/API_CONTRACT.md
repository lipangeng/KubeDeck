# API Contract Delta

## Changed Error Behavior
For affected IAM write endpoints:
- Status: `500 Internal Server Error`
- Body: JSON error containing `persistence_failed`

## Affected APIs
- `POST /api/iam/groups`
- `PATCH /api/iam/groups/{groupID}`
- `DELETE /api/iam/groups/{groupID}`
- `PUT /api/iam/groups/{groupID}/permissions`
- `POST /api/iam/tenants/{tenantID}/members`
- `DELETE /api/iam/tenants/{tenantID}/members`
- `PUT /api/iam/memberships/{membershipID}/groups`
- `PUT /api/iam/memberships/{membershipID}/validity`
- `POST /api/iam/invites`
- `DELETE /api/iam/invites/{inviteID}`
