import { describe, expect, it } from 'vitest';
import type { MenuItem } from '../sdk/types';
import { groupMenusByGroup } from './menuGrouping';

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

describe('groupMenusByGroup', () => {
  it('groups merged menus by group name and keeps item order', () => {
    const grouped = groupMenusByGroup([
      menu({ id: 'pod', group: 'WORKLOAD', source: 'system', order: 10 }),
      menu({ id: 'deploy', group: 'WORKLOAD', source: 'user', order: 20 }),
      menu({ id: 'fav', group: 'FAVORITES', source: 'user', order: 30 }),
    ]);

    expect(grouped).toHaveLength(2);
    expect(grouped[0].name).toBe('WORKLOAD');
    expect(grouped[0].items.map((item) => item.id)).toEqual(['pod', 'deploy']);
    expect(grouped[1].name).toBe('FAVORITES');
  });

  it('falls back to General when menu group is empty', () => {
    const grouped = groupMenusByGroup([menu({ id: 'misc', group: '  ' })]);
    expect(grouped[0].name).toBe('General');
  });
});
