import type { ActionContribution } from '../contracts/actionContribution';
import type { MenuContribution } from '../contracts/menuContribution';
import type { PageContribution } from '../contracts/pageContribution';
import type { SlotContribution } from '../contracts/slotContribution';
import type { KernelNavigationGroup } from './menu/types';

export interface KernelRegistrySnapshot {
  pages: PageContribution[];
  menus: MenuContribution[];
  menuGroups: KernelNavigationGroup[];
  actions: ActionContribution[];
  slots: SlotContribution[];
}
