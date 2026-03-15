import type { RemoteKernelMetadata } from './transport';

export async function fetchKernelMetadata(
  cluster: string,
  scope: 'work' | 'system' | 'cluster' = 'work',
): Promise<RemoteKernelMetadata> {
  const query = new URLSearchParams({ cluster });
  if (scope !== 'work') {
    query.set('scope', scope);
  }
  const response = await fetch(`/api/meta/kernel?${query.toString()}`);
  if (!response.ok) {
    throw new Error(`kernel metadata request failed: ${response.status}`);
  }
  return (await response.json()) as RemoteKernelMetadata;
}
