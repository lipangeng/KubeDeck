import { describe, expect, it } from 'vitest';
import {
  parseGroup,
  parseGroupsResponse,
  parseInvitesResponse,
  parseMembershipsResponse,
  parsePermissionsResponse,
} from './iamApi';

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

  it('parses memberships response', () => {
    const memberships = parseMembershipsResponse({
      memberships: [
        {
          id: 'mbr-u1-tenant-dev',
          tenant_id: 'tenant-dev',
          user_id: 'u1',
          user_label: 'alice',
          group_ids: ['grp-admin'],
        },
      ],
    });
    expect(memberships).toHaveLength(1);
    expect(memberships[0]?.userLabel).toBe('alice');
  });

  it('parses invites response', () => {
    const invites = parseInvitesResponse({
      invites: [
        {
          id: 'inv-1',
          tenant_id: 'tenant-dev',
          tenant_code: 'dev',
          invitee_email: 'a@example.com',
          role_hint: 'member',
          token: 'abc',
          invite_link: '/accept-invite?token=abc',
          expires_at: '2026-02-25T00:00:00Z',
          status: 'pending',
        },
      ],
    });
    expect(invites).toHaveLength(1);
    expect(invites[0]?.inviteLink).toContain('/accept-invite');
  });
});
