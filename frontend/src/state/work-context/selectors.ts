import type { SharedWorkingContext } from './types';

export function selectClusterContext(state: SharedWorkingContext) {
  return state.activeCluster;
}

export function selectNamespaceScope(state: SharedWorkingContext) {
  return state.namespaceScope;
}

export function selectWorkflowDomain(state: SharedWorkingContext) {
  return state.currentWorkflowDomain;
}

export function selectListContext(state: SharedWorkingContext) {
  return state.listContext;
}

export function selectActionContext(state: SharedWorkingContext) {
  return state.actionContext;
}

export function selectHomepageContextSummary(state: SharedWorkingContext) {
  return {
    activeCluster: state.activeCluster,
    namespaceScope: state.namespaceScope,
    workflowDomain: state.currentWorkflowDomain,
  };
}

export function selectWorkloadsContext(state: SharedWorkingContext) {
  return {
    activeCluster: state.activeCluster,
    namespaceScope: state.namespaceScope,
    workflowDomain: state.currentWorkflowDomain,
    listContext: state.listContext,
    actionContext: state.actionContext,
  };
}

export function selectCreateApplyContext(state: SharedWorkingContext) {
  return {
    activeCluster: state.activeCluster,
    namespaceScope: state.namespaceScope,
    workflowDomain: state.currentWorkflowDomain,
    actionContext: state.actionContext,
  };
}

export function selectActionResultContext(state: SharedWorkingContext) {
  return {
    activeCluster: state.activeCluster,
    namespaceScope: state.namespaceScope,
    actionContext: state.actionContext,
  };
}
