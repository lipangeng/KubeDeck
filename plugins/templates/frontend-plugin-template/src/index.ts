import type {
  FrontendCapabilityModule,
  MenuContribution,
  PageContribution,
  SlotContribution,
} from '../../../../frontend/src/kernel/sdk';

const pages = (): PageContribution[] => [];
const menus = (): MenuContribution[] => [];
const slots = (): SlotContribution[] => [];

const plugin: FrontendCapabilityModule = {
  pluginId: 'example-frontend-plugin',
  registerPages: pages,
  registerMenus: menus,
  registerActions: () => [],
  registerSlots: slots,
};

export default plugin;
