import type { MenuContribution } from '../contracts/menuContribution';

export function registerBuiltInMenus(): MenuContribution[] {
  return [
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.homepage',
        contributionId: 'menu.homepage',
      },
      workflowDomainId: 'homepage',
      entryKey: 'homepage',
      placement: 'primary',
      order: 10,
      title: { key: 'homepage.title', fallback: 'Homepage' },
      route: '/',
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.workloads',
        contributionId: 'menu.workloads',
      },
      workflowDomainId: 'workloads',
      entryKey: 'workloads',
      placement: 'primary',
      order: 20,
      title: { key: 'workloads.title', fallback: 'Workloads' },
      route: '/workloads',
    },
  ];
}
