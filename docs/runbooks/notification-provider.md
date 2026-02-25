# Notification Provider (Invite Delivery)

KubeDeck invite delivery supports configurable providers.

## Provider Selection

- `KUBEDECK_NOTIFICATION_PROVIDER=stub`
  - Default mode.
  - Email send is no-op; SMS returns not-implemented.
- `KUBEDECK_NOTIFICATION_PROVIDER=webhook`
  - Sends invite notifications to webhook endpoint.
  - Requires `KUBEDECK_NOTIFICATION_WEBHOOK_URL`.

## Required Variables

- `KUBEDECK_NOTIFICATION_PROVIDER`
- `KUBEDECK_NOTIFICATION_WEBHOOK_URL` (required when provider is `webhook`)

## Webhook Payload

Email payload:

```json
{"type":"email","to":"user@example.com","subject":"KubeDeck Invite","body":"#/accept-invite?token=..."}
```

SMS payload:

```json
{"type":"sms","phone":"+123456","content":"#/accept-invite?token=..."}
```

## Notes

- Webhook endpoint should return HTTP 2xx for success.
- Non-2xx response is treated as delivery failure.
- This is an MVP adapter and can be replaced by SMTP/SMS provider adapters later.
