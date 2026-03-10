import type { ActionContribution } from './contracts/actionContribution';
import type { MenuContribution } from './contracts/menuContribution';
import type { PageContribution } from './contracts/pageContribution';
import type { SlotContribution } from './contracts/slotContribution';

export type { ActionContribution, ActionSurface } from './contracts/actionContribution';
export type { MenuContribution } from './contracts/menuContribution';
export type { PageContribution } from './contracts/pageContribution';
export type { SlotContribution, SlotPlacement } from './contracts/slotContribution';

export interface FrontendCapabilityModule {
  pluginId: string;
  registerPages?: () => PageContribution[];
  registerMenus?: () => MenuContribution[];
  registerActions?: () => ActionContribution[];
  registerSlots?: () => SlotContribution[];
}
