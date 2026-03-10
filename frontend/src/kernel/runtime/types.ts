import type { ActionContribution } from '../contracts/actionContribution';
import type { MenuContribution } from '../contracts/menuContribution';
import type { PageContribution } from '../contracts/pageContribution';
import type { SlotContribution } from '../contracts/slotContribution';

export interface KernelRegistrySnapshot {
  pages: PageContribution[];
  menus: MenuContribution[];
  actions: ActionContribution[];
  slots: SlotContribution[];
}
