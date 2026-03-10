import type { MenuContribution } from '../contracts/menuContribution';
import type { KernelNavigationGroup } from './menu/types';

const GROUP_ORDER: Record<string, number> = {
  core: 10,
  platform: 20,
  extensions: 30,
  resources: 40,
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
    entries: groupEntries,
  }));
}
