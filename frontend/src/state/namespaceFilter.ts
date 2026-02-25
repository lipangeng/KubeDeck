export const ALL_NAMESPACES = 'All';

export interface NamespaceSelectionInput {
  listFilterNamespace: string;
  lastUsedNamespace?: string;
}

export function resolveCreateDefaultNamespace(
  input: NamespaceSelectionInput,
): string {
  if (input.listFilterNamespace && input.listFilterNamespace !== ALL_NAMESPACES) {
    return input.listFilterNamespace;
  }

  if (input.lastUsedNamespace) {
    return input.lastUsedNamespace;
  }

  return 'default';
}
