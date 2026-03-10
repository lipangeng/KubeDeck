import { describe, expect, it } from 'vitest';
import {
  createInitialWorkingContextState,
  reduceWorkingContext,
} from './reducer';
import type { WorkingContextAction } from './types';

function reduce(actions: WorkingContextAction[]) {
  return actions.reduce(reduceWorkingContext, createInitialWorkingContextState());
}

describe('working context reducer', () => {
  it('updates the active cluster and clears current resource on cluster switch', () => {
    const state = reduce([
      {
        type: 'enter_resource',
        resource: { kind: 'Deployment', name: 'api', namespace: 'default' },
      },
      { type: 'request_cluster_switch', cluster: 'prod-eu1' },
    ]);

    expect(state.activeCluster).toBe('prod-eu1');
    expect(state.currentResource).toBeNull();
  });

  it('updates namespace scope without dropping workflow continuity', () => {
    const state = reduce([
      { type: 'enter_workflow_domain', workflowDomainId: 'workloads', route: '/workloads' },
      {
        type: 'update_namespace_scope',
        namespaceScope: { kind: 'all', namespaces: [] },
      },
    ]);

    expect(state.currentWorkflowDomainId).toBe('workloads');
    expect(state.namespaceScope).toEqual({ kind: 'all', namespaces: [] });
  });

  it('records workflow entry and route continuity', () => {
    const state = reduce([
      { type: 'enter_workflow_domain', workflowDomainId: 'workloads', route: '/workloads' },
    ]);

    expect(state.currentWorkflowDomainId).toBe('workloads');
    expect(state.currentRoute).toBe('/workloads');
  });

  it('records resource identity inside the current workflow domain', () => {
    const state = reduce([
      { type: 'enter_workflow_domain', workflowDomainId: 'workloads', route: '/workloads' },
      {
        type: 'enter_resource',
        resource: { kind: 'Deployment', name: 'api', namespace: 'default' },
      },
    ]);

    expect(state.currentWorkflowDomainId).toBe('workloads');
    expect(state.currentResource).toEqual({
      kind: 'Deployment',
      name: 'api',
      namespace: 'default',
    });
  });
});
