export interface KernelActionExecutionRequest {
  actionId: string;
  workflowDomainId: string;
  target: {
    cluster: string;
    namespace: string;
    scope: string;
  };
  input: Record<string, unknown>;
}

export interface KernelActionExecutionResult {
  Accepted: boolean;
  Summary: string;
  AffectedObjects: string[];
  FailedObjects: string[];
}

export async function executeKernelAction(
  request: KernelActionExecutionRequest,
): Promise<KernelActionExecutionResult> {
  const response = await fetch('/api/actions/execute', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  });
  if (!response.ok) {
    throw new Error(`action execution failed: ${response.status}`);
  }
  return (await response.json()) as KernelActionExecutionResult;
}
