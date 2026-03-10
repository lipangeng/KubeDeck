import type { RemoteKernelMetadata } from './transport';

export async function fetchKernelMetadata(): Promise<RemoteKernelMetadata> {
  const response = await fetch('/api/meta/kernel');
  if (!response.ok) {
    throw new Error(`kernel metadata request failed: ${response.status}`);
  }
  return (await response.json()) as RemoteKernelMetadata;
}
