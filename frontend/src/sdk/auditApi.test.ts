import { describe, expect, it } from 'vitest';
import { parseAuditEventsResponse } from './auditApi';

describe('auditApi', () => {
  it('parses audit events response', () => {
    const events = parseAuditEventsResponse({
      events: [
        {
          tenant_id: 'tenant-dev',
          actor_id: 'u1',
          action: 'auth.login',
          target_type: 'session',
          target_id: 'token-1',
          result: 'allowed',
          reason: '',
          created_at: '2026-02-25T00:00:00Z',
        },
      ],
    });
    expect(events).toHaveLength(1);
    expect(events[0]?.action).toBe('auth.login');
  });
});
