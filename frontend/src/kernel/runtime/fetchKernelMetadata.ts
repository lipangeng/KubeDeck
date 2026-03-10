import type { RemoteKernelMetadata } from './transport';

export async function fetchKernelMetadata(cluster: string): Promise<RemoteKernelMetadata> {
  const response = await fetch(`/api/meta/kernel?cluster=${encodeURIComponent(cluster)}`);
  if (!response.ok) {
    throw new Error(`kernel metadata request failed: ${response.status}`);
  }
  return (await response.json()) as RemoteKernelMetadata;
}
