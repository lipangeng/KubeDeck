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

export interface ResourceTabExtension {
  kind: string;
  capabilityType?: 'tab' | 'tab-replace';
  targetTabId?: string;
  tabId?: string;
  createTab: (options: ResolveDefaultTabsOptions) => ResourcePageTab;
}

export interface ResourcePageTakeoverExtension {
  kind: string;
  capabilityType: 'page-takeover';
  renderPage: (options: ResolveDefaultTabsOptions) => ReactNode;
}

export type ResourcePageExtension = ResourceTabExtension | ResourcePageTakeoverExtension;

export interface ResolvedResourcePage {
  tabs: ResourcePageTab[];
  takeoverContent: ReactNode | null;
}

export interface ResolveDefaultTabsOptions {
  resource?: ResourcePageIdentity;
  overviewContent?: ReactNode;
  yamlContent?: ReactNode;
  yamlVariantContent?: ReactNode;
  runtimeContent?: ReactNode;
  logsContent?: ReactNode;
  extensions?: ResourcePageExtension[];
}
