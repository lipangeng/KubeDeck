export interface WorkloadItem {
  id: string;
  name: string;
  kind: string;
  namespace: string;
  status: string;
  health: string;
  updatedAt: string;
}

export async function fetchWorkloads(
  workflowDomainId: string,
  cluster = 'default',
): Promise<WorkloadItem[]> {
  const params = new URLSearchParams({
    workflowDomainId,
    cluster,
  });
  const response = await fetch(`/api/workflows/workloads/items?${params.toString()}`);
  if (!response.ok) {
    throw new Error(`workloads request failed: ${response.status}`);
  }
  return (await response.json()) as WorkloadItem[];
}
