import { describe, expect, it } from 'vitest';
import type {
  FrontendCapabilityModule,
  MenuContribution,
  PageContribution,
  ResourcePageExtension,
  SlotContribution,
} from '../sdk';
import { createLocalKernelSnapshot } from './createLocalKernelSnapshot';

function createPluginPage(): PageContribution {
  return {
    identity: {
      source: 'plugin',
      capabilityId: 'plugin.ops-console',
      contributionId: 'page.ops-console',
    },
    workflowDomainId: 'ops-console',
    route: '/ops-console',
    entryKey: 'ops-console',
    title: { key: 'opsConsole.title', fallback: 'Operations Console' },
    component: () => null,
  };
}

function createPluginMenu(): MenuContribution {
  return {
    identity: {
      source: 'plugin',
      capabilityId: 'plugin.ops-console',
      contributionId: 'menu.ops-console',
    },
    workflowDomainId: 'ops-console',
    entryKey: 'ops-console',
    groupKey: 'extensions',
    placement: 'primary',
    availability: 'enabled',
    route: '/ops-console',
    title: { key: 'opsConsole.title', fallback: 'Operations Console' },
  };
}

function createPluginSlot(): SlotContribution {
  return {
    identity: {
      source: 'plugin',
      capabilityId: 'plugin.ops-console',
      contributionId: 'slot.ops-console.summary',
    },
    workflowDomainId: 'ops-console',
    slotId: 'ops-console.summary',
    placement: 'summary',
    component: () => null,
    visible: true,
    title: { key: 'opsConsole.slot.summary', fallback: 'Operations Summary' },
  };
}

describe('createLocalKernelSnapshot', () => {
  it('registers external frontend capability modules alongside built-ins', () => {
    const resourcePageExtension: ResourcePageExtension = {
      kind: 'Service',
      tabId: 'endpoints',
      createTab: () => ({
        id: 'endpoints',
        title: 'Endpoints',
        capabilityType: 'tab',
        content: null,
      }),
    };

    const pluginModule: FrontendCapabilityModule = {
      pluginId: 'plugin.ops-console',
      registerPages: () => [createPluginPage()],
      registerMenus: () => [createPluginMenu()],
      registerActions: () => [],
      registerSlots: () => [createPluginSlot()],
      registerResourcePageExtensions: () => [resourcePageExtension],
    };

    const snapshot = createLocalKernelSnapshot([pluginModule]);

    expect(
      snapshot.pages.some((page) => page.identity.capabilityId === 'plugin.ops-console'),
    ).toBe(true);
    expect(
      snapshot.menus.some((menu) => menu.identity.capabilityId === 'plugin.ops-console'),
    ).toBe(true);
    expect(
      snapshot.slots.some((slot) => slot.identity.capabilityId === 'plugin.ops-console'),
    ).toBe(true);
    expect(snapshot.resourcePageExtensions).toContain(resourcePageExtension);
  });

  it('keeps takeover extensions registered by frontend capability modules', () => {
    const takeoverExtension: ResourcePageExtension = {
      kind: 'StatefulSet',
      capabilityType: 'page-takeover',
      priority: 60,
      renderPage: () => 'Plugin takeover component',
    };

    const pluginModule: FrontendCapabilityModule = {
      pluginId: 'plugin.ops-console',
      registerPages: () => [],
      registerMenus: () => [],
      registerActions: () => [],
      registerSlots: () => [],
      registerResourcePageExtensions: () => [takeoverExtension],
    };

    const snapshot = createLocalKernelSnapshot([pluginModule]);

    expect(snapshot.resourcePageExtensions).toContain(takeoverExtension);
  });
});
