import type {
  ContributionIdentity,
  ContributionVisibility,
  LocalizedText,
  WorkflowDomainId,
} from './types';

export type ActionSurface = 'drawer' | 'dialog' | 'inline' | 'page';

export interface ActionContribution extends ContributionVisibility {
  identity: ContributionIdentity;
  workflowDomainId: WorkflowDomainId;
  actionId: string;
  title: LocalizedText;
  description?: LocalizedText;
  surface: ActionSurface;
  permissionHint?: string;
}
