import type { MenuItem } from '../sdk/types';

function clone(item: MenuItem): MenuItem {
  return { ...item };
}

export function composeMenus(
  systemMenus: MenuItem[],
  userMenus: MenuItem[],
  dynamicMenus: MenuItem[],
): MenuItem[] {
  return [...systemMenus, ...userMenus, ...dynamicMenus]
    .map(clone)
    .filter((item) => item.visible)
    .sort((a, b) => {
      if (a.order !== b.order) {
        return a.order - b.order;
      }

      return a.id.localeCompare(b.id);
    });
}
