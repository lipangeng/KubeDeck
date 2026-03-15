import { describe, expect, it } from 'vitest';
import { hydrateKernelSnapshot } from './hydrateKernelSnapshot';
import type { KernelRegistrySnapshot } from './types';
import { resolveResourcePage } from '../resource-pages/resolveResourcePage';

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
        {
          Kind: 'Deployment',
          CapabilityType: 'action',
          ActionID: 'restart-rollout',
          Title: { Key: 'deployment.restart', Fallback: 'Restart Rollout' },
        },
        {
          Kind: 'Deployment',
          CapabilityType: 'slot',
          Placement: 'summary',
          Title: { Key: 'deployment.summary', Fallback: 'Deployment Summary' },
          ContentFallback: 'Remote deployment summary',
        },
      ],
    });

    expect(hydrated.resourcePageExtensions).toHaveLength(4);
    expect(
      hydrated.resourcePageExtensions.filter(
        (extension) =>
          extension.kind === 'Deployment' && extension.capabilityType === 'tab',
      ),
    ).toHaveLength(1);
    expect(
      hydrated.resourcePageExtensions.some((extension) => extension.kind === 'Service'),
    ).toBe(true);
    expect(
      hydrated.resourcePageExtensions.some(
        (extension) =>
          extension.kind === 'Deployment' &&
          extension.capabilityType === 'action' &&
          extension.actionId === 'restart-rollout',
      ),
    ).toBe(true);
    expect(
      hydrated.resourcePageExtensions.some(
        (extension) =>
          extension.kind === 'Deployment' &&
          extension.capabilityType === 'slot' &&
          extension.placement === 'summary',
      ),
    ).toBe(true);
  });

  it('prefers a local takeover over an equal-priority remote takeover', () => {
    const localSnapshot: KernelRegistrySnapshot = {
      pages: [],
      menus: [],
      menuGroups: [],
      actions: [],
      slots: [],
      resourcePageExtensions: [
        {
          kind: 'StatefulSet',
          capabilityType: 'page-takeover',
          priority: 20,
          renderPage: () => 'Local takeover',
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
          Kind: 'StatefulSet',
          CapabilityType: 'page-takeover',
          TabID: 'statefulset.takeover',
          Title: { Key: 'statefulset.takeover', Fallback: 'StatefulSet takeover' },
          ContentFallback: 'Remote takeover',
          Priority: 20,
        },
      ],
    });

    const resolution = resolveResourcePage({
      resource: {
        kind: 'StatefulSet',
        name: 'db',
        namespace: 'default',
      },
      extensions: hydrated.resourcePageExtensions,
    });

    expect(resolution.takeoverContent).toBe('Local takeover');
  });

  it('allows a higher-priority remote takeover to override a local takeover', () => {
    const localSnapshot: KernelRegistrySnapshot = {
      pages: [],
      menus: [],
      menuGroups: [],
      actions: [],
      slots: [],
      resourcePageExtensions: [
        {
          kind: 'StatefulSet',
          capabilityType: 'page-takeover',
          priority: 20,
          renderPage: () => 'Local takeover',
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
          Kind: 'StatefulSet',
          CapabilityType: 'page-takeover',
          TabID: 'statefulset.takeover',
          Title: { Key: 'statefulset.takeover', Fallback: 'StatefulSet takeover' },
          ContentFallback: 'Remote takeover',
          Priority: 80,
        },
      ],
    });

    const resolution = resolveResourcePage({
      resource: {
        kind: 'StatefulSet',
        name: 'db',
        namespace: 'default',
      },
      extensions: hydrated.resourcePageExtensions,
    });

    expect(resolution.takeoverContent).toBe('Remote takeover for db');
  });

  it('keeps local workflow actions and summary slots when backend metadata does not replace them', () => {
    const localSnapshot: KernelRegistrySnapshot = {
      pages: [],
      menus: [],
      menuGroups: [],
      actions: [
        {
          identity: {
            source: 'plugin',
            capabilityId: 'plugin.workflow-shell',
            contributionId: 'action.inspect',
          },
          workflowDomainId: 'workloads',
          actionId: 'inspect',
          title: { key: 'actions.inspect', fallback: 'Inspect' },
          surface: 'inline',
          visible: true,
        },
      ],
      slots: [
        {
          identity: {
            source: 'plugin',
            capabilityId: 'plugin.workflow-shell',
            contributionId: 'slot.workloads.summary.workflow-shell',
          },
          workflowDomainId: 'workloads',
          slotId: 'workloads.summary.workflow-shell',
          placement: 'summary',
          visible: true,
          component: () => null,
        },
      ],
      resourcePageExtensions: [],
    };

    const hydrated = hydrateKernelSnapshot(localSnapshot, {
      pages: [],
      menus: [],
      actions: [],
      slots: [],
      resourcePageExtensions: [],
    });

    expect(hydrated.actions.some((action) => action.actionId === 'inspect')).toBe(true);
    expect(hydrated.slots.some((slot) => slot.slotId === 'workloads.summary.workflow-shell')).toBe(true);
  });
});
