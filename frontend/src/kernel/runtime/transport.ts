export interface RemoteTextRef {
  Key: string;
  Fallback: string;
  Description?: string;
}

export interface RemotePageDescriptor {
  ID: string;
  WorkflowDomainID: string;
  Route: string;
  EntryKey: string;
  Title: RemoteTextRef;
  Description?: RemoteTextRef;
}

export interface RemoteMenuDescriptor {
  ID: string;
  WorkflowDomainID: string;
  EntryKey: string;
  GroupKey: string;
  Route: string;
  Placement: 'primary' | 'secondary' | 'context';
  Availability: 'enabled' | 'disabled-unavailable' | 'hidden';
  IsFallback?: boolean;
  Order: number;
  Visible: boolean;
  Title: RemoteTextRef;
  Description?: RemoteTextRef;
}

export interface RemoteMenuBlueprintGroup {
  key: string;
  order: number;
  title: RemoteTextRef;
}

export interface RemoteMenuBlueprintEntry {
  entryKey: string;
  workflowDomainId: string;
  defaultGroupKey: string;
  route: string;
  order: number;
  placement: 'primary' | 'secondary' | 'context';
  title: RemoteTextRef;
  sourceType: 'builtin' | 'plugin' | 'crd' | 'fallback';
  isFallback?: boolean;
}

export interface RemoteMenuBlueprint {
  groups: RemoteMenuBlueprintGroup[];
  entries: RemoteMenuBlueprintEntry[];
}

export interface RemoteMenuMount {
  id: string;
  capabilityId: string;
  sourceType: 'builtin' | 'plugin' | 'crd' | 'fallback';
  workflowDomainId: string;
  entryKey: string;
  groupKey: string;
  route: string;
  placement: 'primary' | 'secondary' | 'context';
  availability: 'enabled' | 'disabled-unavailable' | 'hidden';
  isFallback?: boolean;
  order: number;
  visible: boolean;
  title: RemoteTextRef;
  description?: RemoteTextRef;
}

export interface RemoteMenuOverride {
  scope: 'global' | 'cluster' | 'work-global' | 'work-cluster' | 'system' | 'cluster';
  hiddenEntryKeys?: string[];
  moveEntryKeys?: Record<string, string>;
  pinEntryKeys?: string[];
  groupOrderOverrides?: string[];
  itemOrderOverrides?: Record<string, string[]>;
}

export interface RemoteMenuPreferences {
  globalOverrides: RemoteMenuOverride[];
  clusterOverrides: RemoteMenuOverride[];
}

export interface RemoteResolvedMenuEntry {
  ID: string;
  CapabilityID: string;
  SourceType: 'builtin' | 'plugin' | 'crd' | 'fallback';
  WorkflowDomainID: string;
  EntryKey: string;
  GroupKey: string;
  Route: string;
  Placement: 'primary' | 'secondary' | 'context';
  Availability: 'enabled' | 'disabled-unavailable' | 'hidden';
  IsFallback?: boolean;
  Order: number;
  Visible: boolean;
  Title: RemoteTextRef;
  Description?: RemoteTextRef;
  Mounted: boolean;
  Configured: boolean;
  Pinned?: boolean;
}

export interface RemoteMenuGroup {
  key: string;
  order: number;
  title: RemoteTextRef;
  entries: RemoteResolvedMenuEntry[];
}

export interface RemoteActionDescriptor {
  ID: string;
  WorkflowDomainID: string;
  Surface: 'drawer' | 'dialog' | 'inline' | 'page';
  Visible: boolean;
  PermissionHint?: string;
  Title: RemoteTextRef;
  Description?: RemoteTextRef;
}

export interface RemoteSlotDescriptor {
  ID: string;
  WorkflowDomainID: string;
  SlotID: string;
  Placement: 'summary' | 'panel' | 'toolbar' | 'context';
  Visible: boolean;
  Title?: RemoteTextRef;
}

export interface RemoteResourcePageExtensionDescriptor {
  Kind: string;
  CapabilityType: 'tab' | 'tab-replace' | 'page-takeover' | 'action' | 'slot';
  TargetTabID?: string;
  TabID?: string;
  ActionID?: string;
  Placement?: 'summary';
  Priority?: number;
  Title: RemoteTextRef;
  ContentFallback?: string;
}

export interface RemoteKernelMetadata {
  pages: RemotePageDescriptor[];
  menus: RemoteMenuDescriptor[];
  menuBlueprint?: RemoteMenuBlueprint;
  menuMounts?: RemoteMenuMount[];
  menuOverrides?: RemoteMenuOverride[];
  menuGroups?: RemoteMenuGroup[];
  actions: RemoteActionDescriptor[];
  slots: RemoteSlotDescriptor[];
  resourcePageExtensions?: RemoteResourcePageExtensionDescriptor[];
}
