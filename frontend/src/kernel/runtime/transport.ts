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
  Route: string;
  Placement: 'primary' | 'secondary' | 'context';
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

export interface RemoteKernelMetadata {
  pages: RemotePageDescriptor[];
  menus: RemoteMenuDescriptor[];
  actions: RemoteActionDescriptor[];
}
