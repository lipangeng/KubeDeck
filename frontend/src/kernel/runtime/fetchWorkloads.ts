export interface WorkloadItem {
  id: string;
  name: string;
  kind: string;
  namespace: string;
  status: string;
  health: string;
  updatedAt: string;
}

export async function fetchWorkloads(cluster = 'default'): Promise<WorkloadItem[]> {
  const response = await fetch(
    `/api/workflows/workloads/items?cluster=${encodeURIComponent(cluster)}`,
  );
  if (!response.ok) {
    throw new Error(`workloads request failed: ${response.status}`);
  }
  return (await response.json()) as WorkloadItem[];
}
