import { describe, expect, it } from 'vitest';
import type { ActionContribution } from '../contracts/actionContribution';
import type { MenuContribution } from '../contracts/menuContribution';
import type { PageContribution } from '../contracts/pageContribution';
import { KernelRegistry } from './kernelRegistry';

function createPage(id: string): PageContribution {
  return {
    identity: { source: 'builtin', capabilityId: 'core.homepage', contributionId: id },
    workflowDomainId: id,
    route: `/${id}`,
    entryKey: id,
    title: { key: `${id}.title`, fallback: id },
    component: () => null,
  };
}

function createMenu(id: string): MenuContribution {
  return {
    identity: { source: 'builtin', capabilityId: 'core.homepage', contributionId: id },
    workflowDomainId: id,
    entryKey: id,
    placement: 'primary',
    title: { key: `${id}.title`, fallback: id },
  };
}

function createAction(id: string): ActionContribution {
  return {
    identity: { source: 'builtin', capabilityId: 'core.actions', contributionId: id },
    workflowDomainId: 'workloads',
    actionId: id,
    title: { key: `${id}.title`, fallback: id },
    surface: 'drawer',
  };
}

describe('KernelRegistry', () => {
  it('aggregates registered contribution modules into one snapshot', () => {
    const registry = new KernelRegistry();
    registry.register({
      pages: [createPage('homepage')],
      menus: [createMenu('homepage')],
    });
    registry.register({
      pages: [createPage('workloads')],
      actions: [createAction('apply')],
    });

    const snapshot = registry.snapshot();
    expect(snapshot.pages).toHaveLength(2);
    expect(snapshot.menus).toHaveLength(1);
    expect(snapshot.actions).toHaveLength(1);
    expect(snapshot.slots).toHaveLength(0);
  });
});
