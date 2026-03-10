import type { ActionContribution } from '../contracts/actionContribution';

export function registerBuiltInActions(): ActionContribution[] {
  return [
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.workloads',
        contributionId: 'action.create',
      },
      workflowDomainId: 'workloads',
      actionId: 'create',
      title: { key: 'actions.create', fallback: 'Create' },
      surface: 'drawer',
      order: 10,
    },
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.workloads',
        contributionId: 'action.apply',
      },
      workflowDomainId: 'workloads',
      actionId: 'apply',
      title: { key: 'actions.apply', fallback: 'Apply' },
      surface: 'drawer',
      order: 20,
    },
  ];
}
