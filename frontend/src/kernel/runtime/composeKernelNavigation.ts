import type { MenuContribution } from '../contracts/menuContribution';

function defaultVisible(entry: MenuContribution): boolean {
  return entry.visible ?? true;
}

function defaultOrder(entry: MenuContribution): number {
  return entry.order ?? 0;
}

export function composeKernelNavigation(entries: MenuContribution[]): MenuContribution[] {
  return entries
    .filter(defaultVisible)
    .sort((left, right) => {
      const orderDifference = defaultOrder(left) - defaultOrder(right);
      if (orderDifference !== 0) {
        return orderDifference;
      }
      return left.entryKey.localeCompare(right.entryKey);
    });
}
