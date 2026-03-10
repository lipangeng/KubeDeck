import type { ActionContribution } from '../contracts/actionContribution';

export function resolveWorkflowActions(
  workflowDomainId: string,
  actions: ActionContribution[],
): ActionContribution[] {
  return actions
    .filter((action) => action.workflowDomainId === workflowDomainId)
    .filter((action) => action.visible ?? true)
    .sort((left, right) => (left.order ?? 0) - (right.order ?? 0));
}
