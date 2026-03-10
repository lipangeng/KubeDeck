import type {
  ContributionIdentity,
  ContributionVisibility,
  LocalizedText,
  WorkflowDomainId,
} from './types';

export type MenuPlacement = 'primary' | 'secondary' | 'context';
export type MenuAvailability = 'enabled' | 'disabled-unavailable' | 'hidden';

export interface MenuContribution extends ContributionVisibility {
  identity: ContributionIdentity;
  workflowDomainId: WorkflowDomainId;
  entryKey: string;
  groupKey: string;
  placement: MenuPlacement;
  availability: MenuAvailability;
  isFallback?: boolean;
  title: LocalizedText;
  description?: LocalizedText;
  route?: string;
}
