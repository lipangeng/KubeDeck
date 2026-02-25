function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

export interface AuditEvent {
  tenantID: string;
  actorID: string;
  action: string;
  targetType: string;
  targetID: string;
  result: string;
  reason: string;
  createdAt: string;
}

export function parseAuditEventsResponse(value: unknown): AuditEvent[] {
  if (!isObject(value) || !Array.isArray(value.events)) {
    throw new Error('invalid audit events response');
  }
  return value.events.filter((item) => isObject(item)).map((item) => ({
    tenantID: String(item.tenant_id ?? ''),
    actorID: String(item.actor_id ?? ''),
    action: String(item.action ?? ''),
    targetType: String(item.target_type ?? ''),
    targetID: String(item.target_id ?? ''),
    result: String(item.result ?? ''),
    reason: String(item.reason ?? ''),
    createdAt: String(item.created_at ?? ''),
  }));
}
