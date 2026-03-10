import type { MenuContribution } from '../../contracts/menuContribution';

export interface KernelNavigationGroup {
  key: string;
  entries: MenuContribution[];
}
