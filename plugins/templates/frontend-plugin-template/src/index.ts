import type {
  FrontendCapabilityModule,
  MenuContribution,
  PageContribution,
  ResourcePageExtension,
  SlotContribution,
} from '../../../../frontend/src/kernel/sdk';

const pages = (): PageContribution[] => [];
const menus = (): MenuContribution[] => [];
const slots = (): SlotContribution[] => [];
const resourcePageExtensions = (): ResourcePageExtension[] => [];

const plugin: FrontendCapabilityModule = {
  pluginId: 'example-frontend-plugin',
  registerPages: pages,
  registerMenus: menus,
  registerActions: () => [],
  registerSlots: slots,
  registerResourcePageExtensions: resourcePageExtensions,
};

export default plugin;
