export interface ContributionIdentity {
  source: 'builtin' | 'plugin';
  capabilityId: string;
  contributionId: string;
}

export interface LocalizedText {
  key: string;
  fallback: string;
  description?: string;
}

export type WorkflowDomainId = string;

export interface ContributionVisibility {
  visible?: boolean;
  order?: number;
}
