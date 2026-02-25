import type { MenuItem } from '../sdk/types';

export interface GroupedMenus {
  system: MenuItem[];
  user: MenuItem[];
  dynamic: MenuItem[];
}

export function groupMenusBySource(items: MenuItem[]): GroupedMenus {
  const grouped: GroupedMenus = {
    system: [],
    user: [],
    dynamic: [],
  };

  for (const item of items) {
    grouped[item.source].push(item);
  }

  return grouped;
}
