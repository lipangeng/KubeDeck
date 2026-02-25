import { describe, expect, it } from 'vitest';
import { parseGroup, parseGroupsResponse, parsePermissionsResponse } from './iamApi';

describe('iamApi', () => {
  it('parses permissions response', () => {
    const permissions = parsePermissionsResponse({
      permissions: [
        { code: 'iam:read', scope: 'platform' },
        { code: 'resource:write', scope: 'cluster' },
      ],
    });
    expect(permissions).toHaveLength(2);
    expect(permissions[0]?.code).toBe('iam:read');
  });

  it('parses groups response', () => {
    const groups = parseGroupsResponse({
      groups: [
        {
          id: 'grp-admin',
          tenant_id: 'tenant-dev',
          name: 'admins',
          description: 'platform admins',
          permissions: ['iam:write'],
        },
      ],
    });
    expect(groups).toHaveLength(1);
    expect(groups[0]?.tenantID).toBe('tenant-dev');
  });

  it('parses single group', () => {
    const group = parseGroup({
      id: 'grp-viewer',
      tenant_id: 'tenant-dev',
      name: 'viewer',
      permissions: ['resource:read'],
    });
    expect(group.id).toBe('grp-viewer');
    expect(group.permissions).toEqual(['resource:read']);
  });
});
