export interface NamespaceScope {
  kind: 'single' | 'all';
  namespaces: string[];
}

export interface ResourceIdentity {
  kind: string;
  name: string;
  namespace?: string;
}

export interface WorkingContextState {
  activeCluster: string;
  namespaceScope: NamespaceScope;
  currentWorkflowDomainId: string | null;
  currentRoute: string | null;
  currentResource: ResourceIdentity | null;
}

export type WorkingContextAction =
  | { type: 'request_cluster_switch'; cluster: string }
  | { type: 'update_namespace_scope'; namespaceScope: NamespaceScope }
  | { type: 'enter_workflow_domain'; workflowDomainId: string; route: string }
  | { type: 'enter_resource'; resource: ResourceIdentity };
