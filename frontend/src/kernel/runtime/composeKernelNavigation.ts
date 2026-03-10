import type { MenuContribution } from '../contracts/menuContribution';
import type { KernelNavigationGroup } from './menu/types';

const GROUP_ORDER: Record<string, number> = {
  core: 10,
  platform: 20,
  extensions: 30,
  resources: 40,
};

const GROUP_TITLE: Record<string, { key: string; fallback: string }> = {
  core: { key: 'menu.group.core', fallback: 'Core' },
  platform: { key: 'menu.group.platform', fallback: 'Platform' },
  extensions: { key: 'menu.group.extensions', fallback: 'Extensions' },
  resources: { key: 'menu.group.resources', fallback: 'Resources' },
};

function defaultVisible(entry: MenuContribution): boolean {
  const available = entry.availability ?? 'enabled';
  return (entry.visible ?? true) && available !== 'hidden';
}

function defaultOrder(entry: MenuContribution): number {
  return entry.order ?? 0;
}

export function composeKernelNavigation(entries: MenuContribution[]): KernelNavigationGroup[] {
  const visibleEntries = entries
    .filter(defaultVisible)
    .sort((left, right) => {
      const leftGroupOrder = GROUP_ORDER[left.groupKey] ?? 1000;
      const rightGroupOrder = GROUP_ORDER[right.groupKey] ?? 1000;
      if (leftGroupOrder !== rightGroupOrder) {
        return leftGroupOrder - rightGroupOrder;
      }
      const orderDifference = defaultOrder(left) - defaultOrder(right);
      if (orderDifference !== 0) {
        return orderDifference;
      }
      return left.entryKey.localeCompare(right.entryKey);
    });

  const groups = new Map<string, MenuContribution[]>();
  for (const entry of visibleEntries) {
    const groupEntries = groups.get(entry.groupKey) ?? [];
    groupEntries.push(entry);
    groups.set(entry.groupKey, groupEntries);
  }

  return Array.from(groups.entries()).map(([key, groupEntries]) => ({
    key,
    order: GROUP_ORDER[key] ?? 1000,
    title: GROUP_TITLE[key] ?? { key: `menu.group.${key}`, fallback: key },
    entries: groupEntries,
  }));
}
