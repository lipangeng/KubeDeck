import type { ComponentType } from 'react';
import type {
  ContributionIdentity,
  ContributionVisibility,
  LocalizedText,
  WorkflowDomainId,
} from './types';

export interface PageContribution extends ContributionVisibility {
  identity: ContributionIdentity;
  workflowDomainId: WorkflowDomainId;
  route: string;
  entryKey: string;
  title: LocalizedText;
  description?: LocalizedText;
  component: ComponentType;
}
