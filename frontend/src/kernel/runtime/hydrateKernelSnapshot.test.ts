import { describe, expect, it } from 'vitest';
import { hydrateKernelSnapshot } from './hydrateKernelSnapshot';
import type { KernelRegistrySnapshot } from './types';

describe('hydrateKernelSnapshot', () => {
  it('keeps local built-in resource page extensions and appends remote-only extensions', () => {
    const localSnapshot: KernelRegistrySnapshot = {
      pages: [],
      menus: [],
      menuGroups: [],
      actions: [],
      slots: [],
      resourcePageExtensions: [
        {
          kind: 'Deployment',
          capabilityType: 'tab',
          tabId: 'runtime',
          createTab: () => ({
            id: 'runtime',
            title: 'Runtime',
            capabilityType: 'tab',
            content: null,
          }),
        },
      ],
    };

    const hydrated = hydrateKernelSnapshot(localSnapshot, {
      pages: [],
      menus: [],
      actions: [],
      slots: [],
      resourcePageExtensions: [
        {
          Kind: 'Deployment',
          CapabilityType: 'tab',
          TabID: 'runtime',
          Title: { Key: 'workloads.runtime', Fallback: 'Runtime' },
          ContentFallback: 'Remote runtime',
        },
        {
          Kind: 'Service',
          CapabilityType: 'tab',
          TabID: 'endpoints',
          Title: { Key: 'service.endpoints', Fallback: 'Endpoints' },
          ContentFallback: 'Remote endpoints',
        },
      ],
    });

    expect(hydrated.resourcePageExtensions).toHaveLength(2);
    expect(
      hydrated.resourcePageExtensions.filter((extension) => extension.kind === 'Deployment'),
    ).toHaveLength(1);
    expect(
      hydrated.resourcePageExtensions.some((extension) => extension.kind === 'Service'),
    ).toBe(true);
  });
});
