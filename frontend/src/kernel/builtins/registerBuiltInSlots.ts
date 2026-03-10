import type { SlotContribution } from '../contracts/slotContribution';

function EmptySlot() {
  return null;
}

export function registerBuiltInSlots(): SlotContribution[] {
  return [
    {
      identity: {
        source: 'builtin',
        capabilityId: 'core.workloads',
        contributionId: 'slot.workloads.summary',
      },
      workflowDomainId: 'workloads',
      slotId: 'workloads.summary',
      placement: 'summary',
      component: EmptySlot,
      visible: false,
    },
  ];
}
