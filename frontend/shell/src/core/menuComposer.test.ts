import { describe, expect, it } from 'vitest';
import { composeMenus } from './menuComposer';
import type { MenuItem } from '../sdk/types';

function menu(partial: Partial<MenuItem>): MenuItem {
  return {
    id: partial.id ?? 'id',
    group: partial.group ?? 'default',
    title: partial.title ?? 'title',
    targetType: partial.targetType ?? 'page',
    targetRef: partial.targetRef ?? 'ref',
    source: partial.source ?? 'system',
    order: partial.order ?? 100,
    visible: partial.visible ?? true,
  };
}

describe('composeMenus', () => {
  it('merges and sorts visible items from system, user and dynamic sources', () => {
    const systemMenus = [
      menu({ id: 'sys-hidden', source: 'system', order: 1, visible: false }),
      menu({ id: 'sys-2', source: 'system', order: 20 }),
    ];
    const userMenus = [menu({ id: 'user-1', source: 'user', order: 10 })];
    const dynamicMenus = [
      menu({ id: 'dyn-1', source: 'dynamic', order: 15 }),
      menu({ id: 'dyn-2', source: 'dynamic', order: 20 }),
    ];

    const result = composeMenus(systemMenus, userMenus, dynamicMenus);

    expect(result.map((item) => item.id)).toEqual([
      'user-1',
      'dyn-1',
      'dyn-2',
      'sys-2',
    ]);
    expect(result.every((item) => item.visible)).toBe(true);
  });
});
