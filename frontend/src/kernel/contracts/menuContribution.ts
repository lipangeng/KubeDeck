import type {
  ContributionIdentity,
  ContributionVisibility,
  LocalizedText,
  WorkflowDomainId,
} from './types';

export type MenuPlacement = 'primary' | 'secondary' | 'context';

export interface MenuContribution extends ContributionVisibility {
  identity: ContributionIdentity;
  workflowDomainId: WorkflowDomainId;
  entryKey: string;
  placement: MenuPlacement;
  title: LocalizedText;
  description?: LocalizedText;
  route?: string;
}
