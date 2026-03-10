import type { SlotContribution } from '../contracts/slotContribution';

export function resolveSlotContributions(
  workflowDomainId: string,
  slotId: string,
  slots: SlotContribution[],
): SlotContribution[] {
  return slots
    .filter((slot) => slot.workflowDomainId === workflowDomainId && slot.slotId === slotId)
    .filter((slot) => slot.visible ?? true)
    .sort((left, right) => (left.order ?? 0) - (right.order ?? 0));
}
