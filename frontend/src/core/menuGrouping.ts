import type { MenuItem } from '../sdk/types';

export interface MenuGroup {
  name: string;
  items: MenuItem[];
}

function normalizeGroupName(name: string): string {
  const trimmed = name.trim();
  return trimmed === '' ? 'General' : trimmed;
}

export function groupMenusByGroup(items: MenuItem[]): MenuGroup[] {
  const grouped = new Map<string, MenuItem[]>();
  for (const item of items) {
    const groupName = normalizeGroupName(item.group);
    if (!grouped.has(groupName)) {
      grouped.set(groupName, []);
    }
    grouped.get(groupName)?.push(item);
  }

  return Array.from(grouped.entries()).map(([name, groupedItems]) => ({
    name,
    items: groupedItems,
  }));
}
