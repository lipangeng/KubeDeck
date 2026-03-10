import { describe, expect, it } from 'vitest';
import type { MenuContribution } from '../contracts/menuContribution';
import { composeKernelNavigation } from './composeKernelNavigation';

describe('composeKernelNavigation', () => {
  it('groups entries by blueprint order and filters hidden entries', () => {
    const entries: MenuContribution[] = [
      {
        identity: { source: 'builtin', capabilityId: 'core', contributionId: 'workloads' },
        workflowDomainId: 'workloads',
        entryKey: 'workloads',
        groupKey: 'core',
        placement: 'primary',
        availability: 'enabled',
        title: { key: 'workloads.title', fallback: 'Workloads' },
        order: 20,
      },
      {
        identity: { source: 'builtin', capabilityId: 'core', contributionId: 'hidden' },
        workflowDomainId: 'hidden',
        entryKey: 'hidden',
        groupKey: 'extensions',
        placement: 'secondary',
        availability: 'hidden',
        title: { key: 'hidden.title', fallback: 'Hidden' },
      },
      {
        identity: { source: 'builtin', capabilityId: 'core', contributionId: 'homepage' },
        workflowDomainId: 'homepage',
        entryKey: 'homepage',
        groupKey: 'core',
        placement: 'primary',
        availability: 'enabled',
        title: { key: 'homepage.title', fallback: 'Homepage' },
        order: 10,
      },
      {
        identity: { source: 'builtin', capabilityId: 'core', contributionId: 'crds' },
        workflowDomainId: 'crds',
        entryKey: 'crds',
        groupKey: 'resources',
        placement: 'secondary',
        availability: 'disabled-unavailable',
        isFallback: true,
        title: { key: 'resources.crds.title', fallback: 'CRDs' },
        order: 999,
      },
    ];

    expect(composeKernelNavigation(entries)).toEqual([
      {
        key: 'core',
        order: 10,
        title: { key: 'menu.group.core', fallback: 'Core' },
        entries: [
          expect.objectContaining({ entryKey: 'homepage' }),
          expect.objectContaining({ entryKey: 'workloads' }),
        ],
      },
      {
        key: 'resources',
        order: 40,
        title: { key: 'menu.group.resources', fallback: 'Resources' },
        entries: [expect.objectContaining({ entryKey: 'crds', availability: 'disabled-unavailable' })],
      },
    ]);
  });
});
