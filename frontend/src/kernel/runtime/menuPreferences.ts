import type { RemoteMenuOverride, RemoteMenuPreferences } from './transport';

export type MenuPreferenceScope =
  | 'work-global'
  | 'work-cluster'
  | 'system'
  | 'cluster';

export async function fetchMenuPreferences(cluster: string): Promise<RemoteMenuPreferences> {
  const response = await fetch(`/api/preferences/menu?cluster=${encodeURIComponent(cluster)}`);
  if (!response.ok) {
    throw new Error(`menu preferences request failed: ${response.status}`);
  }
  return (await response.json()) as RemoteMenuPreferences;
}

export async function saveMenuPreferences(
  cluster: string,
  preferences: RemoteMenuPreferences,
): Promise<void> {
  const response = await fetch(`/api/preferences/menu?cluster=${encodeURIComponent(cluster)}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(preferences),
  });
  if (!response.ok) {
    throw new Error(`menu preferences update failed: ${response.status}`);
  }
}

export function isClusterScopedMenuPreference(scope: MenuPreferenceScope): boolean {
  return scope === 'work-cluster' || scope === 'cluster';
}

export function upsertMenuOverride(
  overrides: RemoteMenuOverride[],
  scope: MenuPreferenceScope,
  update: (current: RemoteMenuOverride) => RemoteMenuOverride,
): RemoteMenuOverride[] {
  const existing = overrides.find((override) => override.scope === scope) ?? { scope };
  const next = update(existing);
  const remaining = overrides.filter((override) => override.scope !== scope);
  if (
    !next.hiddenEntryKeys?.length &&
    !next.pinEntryKeys?.length &&
    !next.groupOrderOverrides?.length &&
    !next.itemOrderOverrides &&
    !next.moveEntryKeys
  ) {
    return remaining;
  }
  return [...remaining, next];
}
