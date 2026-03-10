export type ClusterStatus = 'ready' | 'switching' | 'failed';

export interface ActiveCluster {
  id: string;
  status: ClusterStatus;
  lastStableId?: string;
}

export type NamespaceScopeMode = 'single' | 'multiple' | 'all';
export type NamespaceScopeSource = 'default' | 'restored' | 'user_selected';

export interface NamespaceScope {
  mode: NamespaceScopeMode;
  values: string[];
  source: NamespaceScopeSource;
}

export type WorkflowDomainId = 'homepage' | 'workloads';
export type WorkflowDomainSource =
  | 'system_default'
  | 'homepage_entry'
  | 'direct_navigation'
  | 'return_flow';

export interface WorkflowDomain {
  id: WorkflowDomainId;
  source: WorkflowDomainSource;
}

export type SortDirection = 'asc' | 'desc';

export interface ListContext {
  searchText?: string;
  statusFilters?: string[];
  subtypeFilters?: string[];
  sortKey?: string;
  sortDirection?: SortDirection;
}

export type ActionType = 'create' | 'apply';
export type ActionStatus =
  | 'idle'
  | 'editing'
  | 'validating'
  | 'submitting'
  | 'success'
  | 'partial_failure'
  | 'failure';

export interface NamespaceExecutionTarget {
  kind: 'namespace';
  namespace: string;
}

export interface ClusterScopedExecutionTarget {
  kind: 'cluster_scoped';
}

export type ExecutionTarget =
  | NamespaceExecutionTarget
  | ClusterScopedExecutionTarget;

export type ResultOutcome = 'success' | 'partial_failure' | 'failure';

export interface ActionResultSummary {
  outcome: ResultOutcome;
  affectedObjects?: string[];
  failedObjects?: string[];
}

export interface ActionContext {
  actionType?: ActionType;
  originDomain?: Extract<WorkflowDomainId, 'workloads'>;
  status?: ActionStatus;
  executionTarget?: ExecutionTarget;
  resultSummary?: ActionResultSummary;
  needsRevalidation?: boolean;
}

export interface SharedWorkingContext {
  activeCluster: ActiveCluster;
  namespaceScope: NamespaceScope;
  currentWorkflowDomain: WorkflowDomain;
  listContext: ListContext;
  actionContext: ActionContext;
}
