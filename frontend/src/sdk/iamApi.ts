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
