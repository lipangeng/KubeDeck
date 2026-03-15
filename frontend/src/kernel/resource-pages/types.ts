import type { ReactNode } from 'react';

export type ResourceExtensionCapabilityType =
  | 'tab'
  | 'tab-replace'
  | 'page-takeover'
  | 'action'
  | 'slot'
  | 'section';

export interface ResourcePageTab {
  id: string;
  title: string;
  capabilityType: ResourceExtensionCapabilityType;
  content: ReactNode;
}

export interface ResourcePageIdentity {
  kind: string;
  name: string;
  namespace?: string;
}

export interface ResolveDefaultTabsOptions {
  resource?: ResourcePageIdentity;
  overviewContent?: ReactNode;
  yamlContent?: ReactNode;
  runtimeContent?: ReactNode;
}
