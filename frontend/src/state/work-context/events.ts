import {
  completeActionFailure,
  completeActionPartialFailure,
  completeActionSuccess,
  createActionContext,
  failActionValidation,
  markActionForRevalidation,
  resolveExecutionTarget,
  startAction,
  submitAction,
  validateAction,
  acknowledgeActionResult,
} from './actionContext';
import {
  completeClusterSwitch,
  createClusterContext,
  failClusterSwitch,
  requestClusterSwitch,
} from './clusterContext';
import { createListContext, updateListContext } from './listContext';
import {
  createNamespaceScope,
  resetNamespaceScopeForCluster,
  restoreNamespaceScope,
  updateNamespaceScope,
} from './namespaceContext';
import {
  createWorkflowContext,
  enterHomepage,
  enterWorkloads,
  returnToWorkloads,
} from './workflowContext';
import type {
  ActionContext,
  ActionType,
  ExecutionTarget,
  ListContext,
  NamespaceScope,
  SharedWorkingContext,
} from './types';

export type WorkContextEvent =
  | { type: 'enter_homepage' }
  | { type: 'request_cluster_switch' }
  | {
      type: 'complete_cluster_switch';
      clusterId: string;
      restoredNamespaceScope?: NamespaceScope;
    }
  | { type: 'fail_cluster_switch' }
  | { type: 'enter_workloads' }
  | { type: 'update_namespace_scope'; scope: NamespaceScope }
  | { type: 'update_list_context'; next: Partial<ListContext> }
  | { type: 'start_action'; actionType: ActionType }
  | { type: 'validate_action' }
  | { type: 'fail_action_validation' }
  | { type: 'resolve_execution_target'; target: ExecutionTarget }
  | { type: 'submit_action' }
  | {
      type: 'complete_action_success';
      summary?: NonNullable<ActionContext['resultSummary']> extends infer R
        ? R extends { outcome: string }
          ? Omit<R, 'outcome'>
          : never
        : never;
    }
  | {
      type: 'complete_action_partial_failure';
      summary?: NonNullable<ActionContext['resultSummary']> extends infer R
        ? R extends { outcome: string }
          ? Omit<R, 'outcome'>
          : never
        : never;
    }
  | {
      type: 'complete_action_failure';
      summary?: NonNullable<ActionContext['resultSummary']> extends infer R
        ? R extends { outcome: string }
          ? Omit<R, 'outcome'>
          : never
        : never;
    }
  | { type: 'acknowledge_action_result' }
  | { type: 'return_to_workloads' };

export function createSharedWorkingContext(
  clusterId = 'default',
): SharedWorkingContext {
  return {
    activeCluster: createClusterContext(clusterId),
    namespaceScope: createNamespaceScope(),
    currentWorkflowDomain: createWorkflowContext(),
    listContext: createListContext(),
    actionContext: createActionContext(),
  };
}

export function applyWorkContextEvent(
  state: SharedWorkingContext,
  event: WorkContextEvent,
): SharedWorkingContext {
  switch (event.type) {
    case 'enter_homepage':
      return {
        ...state,
        currentWorkflowDomain: enterHomepage(state.currentWorkflowDomain),
      };
    case 'request_cluster_switch':
      return {
        ...state,
        activeCluster: requestClusterSwitch(state.activeCluster),
      };
    case 'complete_cluster_switch':
      return {
        ...state,
        activeCluster: completeClusterSwitch(state.activeCluster, event.clusterId),
        namespaceScope: event.restoredNamespaceScope
          ? restoreNamespaceScope(event.restoredNamespaceScope)
          : resetNamespaceScopeForCluster(),
        actionContext: createActionContext(),
      };
    case 'fail_cluster_switch':
      return {
        ...state,
        activeCluster: failClusterSwitch(state.activeCluster),
      };
    case 'enter_workloads':
      return {
        ...state,
        currentWorkflowDomain: enterWorkloads(),
      };
    case 'update_namespace_scope':
      return {
        ...state,
        namespaceScope: updateNamespaceScope(state.namespaceScope, event.scope),
        actionContext: markActionForRevalidation(state.actionContext),
      };
    case 'update_list_context':
      return {
        ...state,
        listContext: updateListContext(state.listContext, event.next),
      };
    case 'start_action':
      return {
        ...state,
        actionContext: startAction(state.actionContext, event.actionType),
      };
    case 'validate_action':
      return {
        ...state,
        actionContext: validateAction(state.actionContext),
      };
    case 'fail_action_validation':
      return {
        ...state,
        actionContext: failActionValidation(state.actionContext),
      };
    case 'resolve_execution_target':
      return {
        ...state,
        actionContext: resolveExecutionTarget(state.actionContext, event.target),
      };
    case 'submit_action':
      return {
        ...state,
        actionContext: submitAction(state.actionContext),
      };
    case 'complete_action_success':
      return {
        ...state,
        actionContext: completeActionSuccess(state.actionContext, event.summary),
      };
    case 'complete_action_partial_failure':
      return {
        ...state,
        actionContext: completeActionPartialFailure(
          state.actionContext,
          event.summary,
        ),
      };
    case 'complete_action_failure':
      return {
        ...state,
        actionContext: completeActionFailure(state.actionContext, event.summary),
      };
    case 'acknowledge_action_result':
      return {
        ...state,
        actionContext: acknowledgeActionResult(),
      };
    case 'return_to_workloads':
      return {
        ...state,
        currentWorkflowDomain: returnToWorkloads(),
      };
  }
}
