import type { WorkflowDomain } from './types';

export function createWorkflowContext(): WorkflowDomain {
  return {
    id: 'homepage',
    source: 'system_default',
  };
}

export function enterHomepage(
  current: WorkflowDomain,
  source: WorkflowDomain['source'] = 'direct_navigation',
): WorkflowDomain {
  return {
    id: 'homepage',
    source,
  };
}

export function enterWorkloads(
  source: WorkflowDomain['source'] = 'homepage_entry',
): WorkflowDomain {
  return {
    id: 'workloads',
    source,
  };
}

export function returnToWorkloads(): WorkflowDomain {
  return {
    id: 'workloads',
    source: 'return_flow',
  };
}
