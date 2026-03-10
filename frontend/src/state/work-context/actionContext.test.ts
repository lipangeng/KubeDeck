import { describe, expect, it } from 'vitest';
import {
  acknowledgeActionResult,
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
} from './actionContext';

describe('actionContext', () => {
  it('starts empty', () => {
    expect(createActionContext()).toEqual({});
  });

  it('starts an action in editing state', () => {
    expect(startAction(createActionContext(), 'apply')).toEqual({
      actionType: 'apply',
      originDomain: 'workloads',
      status: 'editing',
    });
  });

  it('moves through validation and submission states', () => {
    const started = startAction(createActionContext(), 'create');
    const validating = validateAction(started);
    const readyToSubmit = resolveExecutionTarget(validating, {
      kind: 'namespace',
      namespace: 'default',
    });
    const submitting = submitAction(readyToSubmit);

    expect(validating.status).toBe('validating');
    expect(submitting.status).toBe('submitting');
  });

  it('marks action for revalidation without dropping state', () => {
    const started = startAction(createActionContext(), 'apply');
    expect(markActionForRevalidation(started)).toEqual({
      actionType: 'apply',
      originDomain: 'workloads',
      status: 'editing',
      needsRevalidation: true,
    });
  });

  it('resolves a namespace execution target explicitly', () => {
    const started = startAction(createActionContext(), 'apply');
    expect(
      resolveExecutionTarget(started, {
        kind: 'namespace',
        namespace: 'team-a',
      }),
    ).toEqual({
      actionType: 'apply',
      originDomain: 'workloads',
      status: 'editing',
      executionTarget: {
        kind: 'namespace',
        namespace: 'team-a',
      },
      needsRevalidation: false,
    });
  });

  it('returns to editing after validation failure', () => {
    const validating = validateAction(startAction(createActionContext(), 'apply'));
    expect(failActionValidation(validating).status).toBe('editing');
  });

  it('rejects validation or submission before the action is ready', () => {
    expect(() => validateAction(createActionContext())).toThrow(
      'cannot validate an action before it starts',
    );
    expect(() => submitAction(startAction(createActionContext(), 'apply'))).toThrow(
      'cannot submit an action without an execution target',
    );
  });

  it('rejects submit when namespace changes require revalidation', () => {
    const started = startAction(createActionContext(), 'apply');
    const resolved = resolveExecutionTarget(started, {
      kind: 'namespace',
      namespace: 'team-a',
    });
    const revalidationNeeded = markActionForRevalidation(resolved);

    expect(() => submitAction(revalidationNeeded)).toThrow(
      'cannot submit an action that requires revalidation',
    );
  });

  it('records success, partial failure, and failure results', () => {
    const submitted = submitAction(
      resolveExecutionTarget(startAction(createActionContext(), 'apply'), {
        kind: 'namespace',
        namespace: 'default',
      }),
    );
    expect(
      completeActionSuccess(submitted, { affectedObjects: ['deployment/api'] }),
    ).toMatchObject({
      status: 'success',
      resultSummary: {
        outcome: 'success',
        affectedObjects: ['deployment/api'],
      },
    });
    expect(
      completeActionPartialFailure(submitted, {
        affectedObjects: ['deployment/api'],
        failedObjects: ['service/api'],
      }),
    ).toMatchObject({
      status: 'partial_failure',
      resultSummary: {
        outcome: 'partial_failure',
        failedObjects: ['service/api'],
      },
    });
    expect(
      completeActionFailure(submitted, { failedObjects: ['deployment/api'] }),
    ).toMatchObject({
      status: 'failure',
      resultSummary: {
        outcome: 'failure',
        failedObjects: ['deployment/api'],
      },
    });
  });

  it('clears action result state on acknowledgement', () => {
    expect(acknowledgeActionResult()).toEqual({});
  });
});
