import { describe, expect, it } from 'vitest';
import {
  applyWorkContextEvent,
  createSharedWorkingContext,
} from './events';

describe('applyWorkContextEvent', () => {
  it('enters workloads without mutating cluster or namespace scope', () => {
    const state = createSharedWorkingContext('prod');

    const next = applyWorkContextEvent(state, { type: 'enter_workloads' });

    expect(next.currentWorkflowDomain.id).toBe('workloads');
    expect(next.activeCluster).toEqual(state.activeCluster);
    expect(next.namespaceScope).toEqual(state.namespaceScope);
  });

  it('marks action context for revalidation when namespace scope changes', () => {
    const state = applyWorkContextEvent(createSharedWorkingContext(), {
      type: 'start_action',
      actionType: 'apply',
    });

    const next = applyWorkContextEvent(state, {
      type: 'update_namespace_scope',
      scope: {
        mode: 'single',
        values: ['team-a'],
        source: 'default',
      },
    });

    expect(next.namespaceScope).toEqual({
      mode: 'single',
      values: ['team-a'],
      source: 'user_selected',
    });
    expect(next.actionContext.needsRevalidation).toBe(true);
  });

  it('does not rewrite execution target when namespace browsing scope changes', () => {
    let state = createSharedWorkingContext();
    state = applyWorkContextEvent(state, {
      type: 'start_action',
      actionType: 'apply',
    });
    state = applyWorkContextEvent(state, {
      type: 'resolve_execution_target',
      target: { kind: 'namespace', namespace: 'team-a' },
    });

    const next = applyWorkContextEvent(state, {
      type: 'update_namespace_scope',
      scope: {
        mode: 'all',
        values: [],
        source: 'default',
      },
    });

    expect(next.namespaceScope).toEqual({
      mode: 'all',
      values: [],
      source: 'user_selected',
    });
    expect(next.actionContext.executionTarget).toEqual({
      kind: 'namespace',
      namespace: 'team-a',
    });
    expect(next.actionContext.needsRevalidation).toBe(true);
  });

  it('restores or resets namespace scope when cluster switch completes', () => {
    const state = applyWorkContextEvent(createSharedWorkingContext('prod'), {
      type: 'request_cluster_switch',
    });

    const next = applyWorkContextEvent(state, {
      type: 'complete_cluster_switch',
      clusterId: 'staging',
    });

    expect(next.activeCluster.id).toBe('staging');
    expect(next.activeCluster.status).toBe('ready');
    expect(next.namespaceScope).toEqual({
      mode: 'single',
      values: ['default'],
      source: 'default',
    });
    expect(next.actionContext).toEqual({});
  });

  it('keeps unrelated browsing continuity on action completion', () => {
    let state = createSharedWorkingContext('prod');
    state = applyWorkContextEvent(state, { type: 'enter_workloads' });
    state = applyWorkContextEvent(state, {
      type: 'update_list_context',
      next: { searchText: 'api' },
    });
    state = applyWorkContextEvent(state, {
      type: 'start_action',
      actionType: 'apply',
    });
    state = applyWorkContextEvent(state, {
      type: 'resolve_execution_target',
      target: { kind: 'namespace', namespace: 'default' },
    });
    state = applyWorkContextEvent(state, { type: 'submit_action' });

    const next = applyWorkContextEvent(state, {
      type: 'complete_action_success',
      summary: { affectedObjects: ['deployment/api'] },
    });

    expect(next.listContext).toEqual({ searchText: 'api' });
    expect(next.currentWorkflowDomain.id).toBe('workloads');
    expect(next.actionContext.resultSummary).toEqual({
      outcome: 'success',
      affectedObjects: ['deployment/api'],
    });
  });

  it('acknowledges result state without changing cluster or namespace', () => {
    let state = createSharedWorkingContext('prod');
    state = applyWorkContextEvent(state, {
      type: 'start_action',
      actionType: 'apply',
    });
    state = applyWorkContextEvent(state, {
      type: 'complete_action_failure',
      summary: { failedObjects: ['deployment/api'] },
    });

    const next = applyWorkContextEvent(state, {
      type: 'acknowledge_action_result',
    });

    expect(next.actionContext).toEqual({});
    expect(next.activeCluster).toEqual(state.activeCluster);
    expect(next.namespaceScope).toEqual(state.namespaceScope);
  });
});
