import type {
  NamespaceScope,
  NamespaceScopeMode,
  NamespaceScopeSource,
} from './types';

function normalizeValues(values: string[]): string[] {
  return [...new Set(values.filter((value) => value.trim() !== ''))];
}

export function createNamespaceScope(
  mode: NamespaceScopeMode = 'single',
  values: string[] = ['default'],
  source: NamespaceScopeSource = 'default',
): NamespaceScope {
  if (mode === 'all') {
    return { mode, values: [], source };
  }

  const normalizedValues = normalizeValues(values);
  if (mode === 'single') {
    if (normalizedValues.length > 1) {
      throw new Error('single namespace scope must contain exactly one value');
    }

    const namespace = normalizedValues[0] ?? 'default';
    return { mode, values: [namespace], source };
  }

  if (normalizedValues.length <= 1) {
    throw new Error('multiple namespace scope must contain more than one value');
  }

  return {
    mode,
    values: normalizedValues,
    source,
  };
}

export function restoreNamespaceScope(
  scope: NamespaceScope | undefined,
): NamespaceScope {
  if (!scope) {
    return createNamespaceScope('single', ['default'], 'default');
  }

  return createNamespaceScope(scope.mode, scope.values, 'restored');
}

export function updateNamespaceScope(
  _current: NamespaceScope,
  next: NamespaceScope,
): NamespaceScope {
  return createNamespaceScope(next.mode, next.values, 'user_selected');
}

export function resetNamespaceScopeForCluster(): NamespaceScope {
  return createNamespaceScope('single', ['default'], 'default');
}
