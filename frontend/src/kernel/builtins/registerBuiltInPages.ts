import type { PageContribution } from '../contracts/pageContribution';
import { HomepagePage } from './pages/HomepagePage';
import { WorkloadsPage } from './pages/WorkloadsPage';

export function registerBuiltInPages(): PageContribution[] {
  return [
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.homepage',
        contributionId: 'page.homepage',
      },
      workflowDomainId: 'homepage',
      route: '/',
      entryKey: 'homepage',
      title: { key: 'homepage.title', fallback: 'Homepage' },
      description: {
        key: 'homepage.description',
        fallback: 'Enter the core built-in workflow domains from the kernel shell.',
      },
      component: HomepagePage,
      order: 10,
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.workloads',
        contributionId: 'page.workloads',
      },
      workflowDomainId: 'workloads',
      route: '/workloads',
      entryKey: 'workloads',
      title: { key: 'workloads.title', fallback: 'Workloads' },
      description: {
        key: 'workloads.description',
        fallback: 'View the first built-in workload domain through the kernel registry.',
      },
      component: WorkloadsPage,
      order: 20,
    },
  ];
}
