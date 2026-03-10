import type { SlotContribution, SlotPlacement } from '../contracts/slotContribution';

export function resolveWorkflowSlots(
  workflowDomainId: string,
  slots: SlotContribution[],
  placement?: SlotPlacement,
): SlotContribution[] {
  return slots
    .filter((slot) => slot.workflowDomainId === workflowDomainId)
    .filter((slot) => slot.visible !== false)
    .filter((slot) => (placement ? slot.placement === placement : true));
}
