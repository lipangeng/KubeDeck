import type { ComponentType } from 'react';
import type {
  ContributionIdentity,
  ContributionVisibility,
  LocalizedText,
  WorkflowDomainId,
} from './types';

export type SlotPlacement = 'summary' | 'panel' | 'toolbar' | 'context';

export interface SlotContribution extends ContributionVisibility {
  identity: ContributionIdentity;
  workflowDomainId: WorkflowDomainId;
  slotId: string;
  placement: SlotPlacement;
  title?: LocalizedText;
  component: ComponentType;
}
