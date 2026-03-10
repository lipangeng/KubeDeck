import { describe, expect, it } from 'vitest';
import {
  parseClustersResponse,
  parseMenusResponse,
  parseRegistryResponse,
  parseWorkloadsResponse,
} from './metaApi';

describe('parseMenusResponse', () => {
  it('parses valid menu response', () => {
    const parsed = parseMenusResponse({
      cluster: 'dev',
      menus: [
        {
          id: 'workloads',
          group: 'system',
          title: 'Workloads',
          targetType: 'page',
          targetRef: '/workloads',
          source: 'system',
          order: 10,
          visible: true,
        },
      ],
    });

    expect(parsed.cluster).toBe('dev');
    expect(parsed.menus).toHaveLength(1);
    expect(parsed.menus[0].id).toBe('workloads');
  });

  it('throws when required fields are missing', () => {
    expect(() =>
      parseMenusResponse({
        cluster: 'dev',
        menus: [
          {
            id: 'workloads',
            title: 'Workloads',
            source: 'system',
            order: 10,
            visible: true,
          },
        ],
      }),
    ).toThrowError(/group is required/);
  });
});

describe('parseClustersResponse', () => {
  it('parses valid clusters response', () => {
    const parsed = parseClustersResponse({ clusters: ['default', 'dev'] });
    expect(parsed.clusters).toEqual(['default', 'dev']);
  });
});

describe('parseRegistryResponse', () => {
  it('parses typed registry response', () => {
    const parsed = parseRegistryResponse({
      cluster: 'default',
      resourceTypes: [
        {
          id: 'apps.v1.deployments',
          group: 'apps',
          version: 'v1',
          kind: 'Deployment',
          plural: 'deployments',
          namespaced: true,
          preferredVersion: 'v1',
          source: 'system',
        },
      ],
    });

    expect(parsed.cluster).toBe('default');
    expect(parsed.resourceTypes[0].kind).toBe('Deployment');
  });
});

describe('parseWorkloadsResponse', () => {
  it('parses typed workloads response', () => {
    const parsed = parseWorkloadsResponse({
      cluster: 'dev',
      namespace: 'default',
      items: [
        {
          id: 'dev-default-api',
          name: 'api',
          kind: 'Deployment',
          namespace: 'default',
          status: 'Running',
          health: 'Healthy',
          updatedAt: '2026-03-10T06:30:00Z',
        },
      ],
    });

    expect(parsed.cluster).toBe('dev');
    expect(parsed.items[0].name).toBe('api');
    expect(parsed.items[0].status).toBe('Running');
  });
});
