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
  priority?: number;
  origin?: 'local' | 'remote';
  createTab: (options: ResolveDefaultTabsOptions) => ResourcePageTab;
}

export interface ResourcePageTakeoverExtension {
  kind: string;
  capabilityType: 'page-takeover';
  priority?: number;
  origin?: 'local' | 'remote';
  renderPage: (options: ResolveDefaultTabsOptions) => ReactNode;
}

export interface ResourcePageSummarySlotExtension {
  kind: string;
  capabilityType: 'slot';
  placement: 'summary';
  priority?: number;
  origin?: 'local' | 'remote';
  renderSlot: (options: ResolveDefaultTabsOptions) => ReactNode;
}

export interface ResourcePageResolvedAction {
  id: string;
  title: string;
  actionId: string;
}

export interface ResourcePageActionExtension {
  kind: string;
  capabilityType: 'action';
  actionId: string;
  priority?: number;
  origin?: 'local' | 'remote';
  createAction: (options: ResolveDefaultTabsOptions) => ResourcePageResolvedAction;
}

export type ResourcePageExtension =
  | ResourceTabExtension
  | ResourcePageTakeoverExtension
  | ResourcePageSummarySlotExtension
  | ResourcePageActionExtension;

export interface ResolvedResourcePage {
  tabs: ResourcePageTab[];
  takeoverContent: ReactNode | null;
  summaryContent: ReactNode[];
  actions: ResourcePageResolvedAction[];
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
