import { describe, expect, it } from 'vitest';
import type { MenuContribution } from '../contracts/menuContribution';
import { composeKernelNavigation } from './composeKernelNavigation';

describe('composeKernelNavigation', () => {
  it('filters hidden entries and sorts by order then key', () => {
    const entries: MenuContribution[] = [
      {
        identity: { source: 'builtin', capabilityId: 'core', contributionId: 'workloads' },
        workflowDomainId: 'workloads',
        entryKey: 'workloads',
        placement: 'primary',
        title: { key: 'workloads.title', fallback: 'Workloads' },
        order: 20,
      },
      {
        identity: { source: 'builtin', capabilityId: 'core', contributionId: 'hidden' },
        workflowDomainId: 'hidden',
        entryKey: 'hidden',
        placement: 'secondary',
        title: { key: 'hidden.title', fallback: 'Hidden' },
        visible: false,
      },
      {
        identity: { source: 'builtin', capabilityId: 'core', contributionId: 'homepage' },
        workflowDomainId: 'homepage',
        entryKey: 'homepage',
        placement: 'primary',
        title: { key: 'homepage.title', fallback: 'Homepage' },
        order: 10,
      },
    ];

    expect(composeKernelNavigation(entries).map((entry) => entry.entryKey)).toEqual([
      'homepage',
      'workloads',
    ]);
  });
});
