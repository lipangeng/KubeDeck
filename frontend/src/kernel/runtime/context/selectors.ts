import type { WorkingContextState } from './types';

export function selectActiveCluster(state: WorkingContextState): string {
  return state.activeCluster;
}

export function selectNamespaceScope(state: WorkingContextState) {
  return state.namespaceScope;
}

export function selectCurrentWorkflowDomain(state: WorkingContextState): string | null {
  return state.currentWorkflowDomainId;
}

export function selectCurrentResource(state: WorkingContextState) {
  return state.currentResource;
}
