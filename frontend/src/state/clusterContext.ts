import {
  resolveCreateDefaultNamespace,
  type NamespaceSelectionInput,
} from './namespaceFilter';

export interface ClusterSwitchState {
  activeCluster: string;
  namespaceFilter: string;
  resourceCache: Record<string, unknown>;
  pageState: Record<string, unknown>;
}

export interface ClusterSwitchInput {
  targetCluster: string;
  namespace: NamespaceSelectionInput;
}

export function switchClusterContext(
  input: ClusterSwitchInput,
): ClusterSwitchState {
  return {
    activeCluster: input.targetCluster,
    namespaceFilter: resolveCreateDefaultNamespace(input.namespace),
    resourceCache: {},
    pageState: {},
  };
}
