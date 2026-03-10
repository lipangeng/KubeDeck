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

export interface RemoteKernelMetadata {
  pages: RemotePageDescriptor[];
  menus: RemoteMenuDescriptor[];
  actions: RemoteActionDescriptor[];
  slots: RemoteSlotDescriptor[];
}
