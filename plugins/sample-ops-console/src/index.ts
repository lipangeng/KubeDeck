import { createElement } from 'react';
import type {
  FrontendCapabilityModule,
  MenuContribution,
  PageContribution,
  SlotContribution,
} from '../../../frontend/src/kernel/sdk';

function SampleOpsConsolePage() {
  return createElement('div', null, 'Sample Ops Console Plugin');
}

const pages = (): PageContribution[] => [
  {
    identity: {
      source: 'plugin',
      capabilityId: 'plugin.sample-ops-console',
      contributionId: 'page.sample-ops-console',
    },
    workflowDomainId: 'sample-ops-console',
    route: '/sample-ops-console',
    entryKey: 'sample-ops-console',
    title: {
      key: 'sampleOpsConsole.title',
      fallback: 'Sample Ops Console',
    },
    description: {
      key: 'sampleOpsConsole.description',
      fallback: 'A sample plugin contribution used to validate end-to-end plugin discovery.',
    },
    component: SampleOpsConsolePage,
  },
];

const menus = (): MenuContribution[] => [
  {
    identity: {
      source: 'plugin',
      capabilityId: 'plugin.sample-ops-console',
      contributionId: 'menu.sample-ops-console',
    },
    workflowDomainId: 'sample-ops-console',
    entryKey: 'sample-ops-console',
    placement: 'primary',
    route: '/sample-ops-console',
    title: {
      key: 'sampleOpsConsole.title',
      fallback: 'Sample Ops Console',
    },
  },
];

const slots = (): SlotContribution[] => [
  {
    identity: {
      source: 'plugin',
      capabilityId: 'plugin.sample-ops-console',
      contributionId: 'slot.sample-ops-console.summary',
    },
    workflowDomainId: 'sample-ops-console',
    slotId: 'sample-ops-console.summary',
    placement: 'summary',
    title: {
      key: 'sampleOpsConsole.slots.summary',
      fallback: 'Sample Ops Console Summary',
    },
    visible: true,
    component: () => createElement('div', null, 'Sample Ops Console Summary'),
  },
];

const plugin: FrontendCapabilityModule = {
  pluginId: 'plugin.sample-ops-console',
  registerPages: pages,
  registerMenus: menus,
  registerActions: () => [],
  registerSlots: slots,
};

export default plugin;
