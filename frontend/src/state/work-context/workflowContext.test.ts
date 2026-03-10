import { describe, expect, it } from 'vitest';
import {
  createWorkflowContext,
  enterHomepage,
  enterWorkloads,
  returnToWorkloads,
} from './workflowContext';

describe('workflowContext', () => {
  it('creates a homepage default workflow context', () => {
    expect(createWorkflowContext()).toEqual({
      id: 'homepage',
      source: 'system_default',
    });
  });

  it('enters workloads from homepage entry', () => {
    expect(enterWorkloads()).toEqual({
      id: 'workloads',
      source: 'homepage_entry',
    });
  });

  it('can return to homepage with direct navigation semantics', () => {
    expect(
      enterHomepage({ id: 'workloads', source: 'homepage_entry' }),
    ).toEqual({
      id: 'homepage',
      source: 'direct_navigation',
    });
  });

  it('returns to workloads with return-flow semantics', () => {
    expect(returnToWorkloads()).toEqual({
      id: 'workloads',
      source: 'return_flow',
    });
  });
});
