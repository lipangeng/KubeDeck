import type {
  ActionContext,
  ActionResultSummary,
  ActionType,
  ExecutionTarget,
} from './types';

export function createActionContext(): ActionContext {
  return {};
}

export function startAction(
  _current: ActionContext,
  actionType: ActionType,
): ActionContext {
  return {
    actionType,
    originDomain: 'workloads',
    status: 'editing',
  };
}

export function validateAction(current: ActionContext): ActionContext {
  if (!current.actionType) {
    throw new Error('cannot validate an action before it starts');
  }

  return {
    ...current,
    status: 'validating',
  };
}

export function failActionValidation(current: ActionContext): ActionContext {
  return {
    ...current,
    status: 'editing',
  };
}

export function resolveExecutionTarget(
  current: ActionContext,
  target: ExecutionTarget,
): ActionContext {
  if (!current.actionType) {
    throw new Error('cannot resolve execution target before action start');
  }

  return {
    ...current,
    executionTarget: target,
    needsRevalidation: false,
  };
}

export function markActionForRevalidation(
  current: ActionContext,
): ActionContext {
  if (!current.status) {
    return current;
  }

  return {
    ...current,
    needsRevalidation: true,
  };
}

export function submitAction(current: ActionContext): ActionContext {
  if (!current.actionType) {
    throw new Error('cannot submit an action before it starts');
  }

  if (!current.executionTarget) {
    throw new Error('cannot submit an action without an execution target');
  }

  if (current.needsRevalidation) {
    throw new Error('cannot submit an action that requires revalidation');
  }

  return {
    ...current,
    status: 'submitting',
  };
}

function completeAction(
  current: ActionContext,
  outcome: ActionResultSummary['outcome'],
  summary: Omit<ActionResultSummary, 'outcome'>,
): ActionContext {
  return {
    ...current,
    status: outcome,
    resultSummary: {
      outcome,
      ...summary,
    },
    needsRevalidation: false,
  };
}

export function completeActionSuccess(
  current: ActionContext,
  summary: Omit<ActionResultSummary, 'outcome'> = {},
): ActionContext {
  return completeAction(current, 'success', summary);
}

export function completeActionPartialFailure(
  current: ActionContext,
  summary: Omit<ActionResultSummary, 'outcome'> = {},
): ActionContext {
  return completeAction(current, 'partial_failure', summary);
}

export function completeActionFailure(
  current: ActionContext,
  summary: Omit<ActionResultSummary, 'outcome'> = {},
): ActionContext {
  return completeAction(current, 'failure', summary);
}

export function acknowledgeActionResult(): ActionContext {
  return {};
}
