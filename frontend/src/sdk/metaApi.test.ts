import { describe, expect, it } from 'vitest';
import { parseMenusResponse } from './metaApi';

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
