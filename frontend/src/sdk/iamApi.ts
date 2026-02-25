function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

export interface IAMPermission {
  code: string;
  scope: string;
}

export interface IAMGroup {
  id: string;
  tenantID: string;
  name: string;
  description: string;
  permissions: string[];
}

export interface IAMMembership {
  id: string;
  tenantID: string;
  userID: string;
  userLabel: string;
  groupIDs: string[];
}

export interface IAMInvite {
  id: string;
  tenantID: string;
  tenantCode: string;
  inviteeEmail: string;
  inviteePhone: string;
  roleHint: string;
  token: string;
  inviteLink: string;
  expiresAt: string;
  status: string;
}

export function parsePermissionsResponse(value: unknown): IAMPermission[] {
  if (!isObject(value) || !Array.isArray(value.permissions)) {
    throw new Error('invalid iam permissions response');
  }
  return value.permissions
    .filter((item) => isObject(item))
    .map((item) => ({
      code: String(item.code ?? ''),
      scope: String(item.scope ?? ''),
    }))
    .filter((item) => item.code !== '');
}

export function parseGroupsResponse(value: unknown): IAMGroup[] {
  if (!isObject(value) || !Array.isArray(value.groups)) {
    throw new Error('invalid iam groups response');
  }
  return value.groups.filter((item) => isObject(item)).map(parseGroup);
}

export function parseGroup(value: Record<string, unknown>): IAMGroup {
  return {
    id: String(value.id ?? ''),
    tenantID: String(value.tenant_id ?? ''),
    name: String(value.name ?? ''),
    description: String(value.description ?? ''),
    permissions: Array.isArray(value.permissions)
      ? value.permissions
          .map((item) => String(item))
          .filter((item) => item.trim() !== '')
      : [],
  };
}

export function parseMembershipsResponse(value: unknown): IAMMembership[] {
  if (!isObject(value) || !Array.isArray(value.memberships)) {
    throw new Error('invalid iam memberships response');
  }
  return value.memberships.filter((item) => isObject(item)).map((item) => ({
    id: String(item.id ?? ''),
    tenantID: String(item.tenant_id ?? ''),
    userID: String(item.user_id ?? ''),
    userLabel: String(item.user_label ?? ''),
    groupIDs: Array.isArray(item.group_ids)
      ? item.group_ids.map((groupID) => String(groupID)).filter((groupID) => groupID !== '')
      : [],
  }));
}

export function parseInvitesResponse(value: unknown): IAMInvite[] {
  if (!isObject(value) || !Array.isArray(value.invites)) {
    throw new Error('invalid iam invites response');
  }
  return value.invites.filter((item) => isObject(item)).map((item) => ({
    id: String(item.id ?? ''),
    tenantID: String(item.tenant_id ?? ''),
    tenantCode: String(item.tenant_code ?? ''),
    inviteeEmail: String(item.invitee_email ?? ''),
    inviteePhone: String(item.invitee_phone ?? ''),
    roleHint: String(item.role_hint ?? ''),
    token: String(item.token ?? ''),
    inviteLink: String(item.invite_link ?? ''),
    expiresAt: String(item.expires_at ?? ''),
    status: String(item.status ?? ''),
  }));
}
