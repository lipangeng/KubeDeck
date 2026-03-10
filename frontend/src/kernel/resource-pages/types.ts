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

export interface ResolveDefaultTabsOptions {
  overviewContent?: ReactNode;
  yamlContent?: ReactNode;
}
