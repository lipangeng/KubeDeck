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
const resourcePageExtensions = (): ResourcePageExtension[] => [
  {
    kind: 'StatefulSet',
    capabilityType: 'page-takeover',
    priority: 50,
    renderPage: (options) =>
      `Example frontend plugin StatefulSet takeover for ${options.resource?.name ?? 'statefulset'}`,
  },
];

const plugin: FrontendCapabilityModule = {
  pluginId: 'example-frontend-plugin',
  registerPages: pages,
  registerMenus: menus,
  registerActions: () => [],
  registerSlots: slots,
  registerResourcePageExtensions: resourcePageExtensions,
};

export default plugin;
