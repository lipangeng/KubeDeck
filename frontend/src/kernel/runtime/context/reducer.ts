import type { WorkingContextAction, WorkingContextState } from './types';

export function createInitialWorkingContextState(): WorkingContextState {
  return {
    activeCluster: 'default',
    namespaceScope: {
      kind: 'single',
      namespaces: ['default'],
    },
    currentWorkflowDomainId: 'homepage',
    currentRoute: '/',
    currentResource: null,
  };
}

export function reduceWorkingContext(
  state: WorkingContextState,
  action: WorkingContextAction,
): WorkingContextState {
  switch (action.type) {
    case 'request_cluster_switch':
      return {
        ...state,
        activeCluster: action.cluster,
        currentResource: null,
      };
    case 'update_namespace_scope':
      return {
        ...state,
        namespaceScope: action.namespaceScope,
      };
    case 'enter_workflow_domain':
      return {
        ...state,
        currentWorkflowDomainId: action.workflowDomainId,
        currentRoute: action.route,
        currentResource: null,
      };
    case 'enter_resource':
      return {
        ...state,
        currentResource: action.resource,
      };
    default:
      return state;
  }
}
