import { describe, expect, it } from 'vitest';
import type { MenuItem } from '../sdk/types';
import { groupMenusBySource } from './menuGrouping';

function menu(partial: Partial<MenuItem>): MenuItem {
  return {
    id: partial.id ?? 'id',
    group: partial.group ?? 'default',
    title: partial.title ?? 'Title',
    targetType: partial.targetType ?? 'page',
    targetRef: partial.targetRef ?? '/stub',
    source: partial.source ?? 'system',
    order: partial.order ?? 0,
    visible: partial.visible ?? true,
  };
}

describe('groupMenusBySource', () => {
  it('groups menus by source and keeps ordering', () => {
    const grouped = groupMenusBySource([
      menu({ id: 'dyn', source: 'dynamic', order: 30 }),
      menu({ id: 'sys', source: 'system', order: 10 }),
      menu({ id: 'usr', source: 'user', order: 20 }),
    ]);

    expect(grouped.system.map((item) => item.id)).toEqual(['sys']);
    expect(grouped.user.map((item) => item.id)).toEqual(['usr']);
    expect(grouped.dynamic.map((item) => item.id)).toEqual(['dyn']);
  });
});
