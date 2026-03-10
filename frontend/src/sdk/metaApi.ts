import type {
  ClustersResponse,
  MenuItem,
  MenusResponse,
  MenuSource,
  RegistryResponse,
  WorkloadsResponse,
} from './types';

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function isMenuSource(value: unknown): value is MenuSource {
  return value === 'system' || value === 'user' || value === 'dynamic';
}

function parseMenuItem(value: unknown): MenuItem {
  if (!isObject(value)) {
    throw new Error('invalid menu payload: menu item is not an object');
  }

  const { id, group, title, targetType, targetRef, source, order, visible } = value;

  if (typeof id !== 'string' || id === '') {
    throw new Error('invalid menu payload: id is required');
  }
  if (typeof group !== 'string' || group === '') {
    throw new Error('invalid menu payload: group is required');
  }
  if (typeof title !== 'string' || title === '') {
    throw new Error('invalid menu payload: title is required');
  }
  if (targetType !== 'page' && targetType !== 'resource') {
    throw new Error('invalid menu payload: targetType must be page|resource');
  }
  if (typeof targetRef !== 'string' || targetRef === '') {
    throw new Error('invalid menu payload: targetRef is required');
  }
  if (!isMenuSource(source)) {
    throw new Error('invalid menu payload: source must be system|user|dynamic');
  }
  if (typeof order !== 'number') {
    throw new Error('invalid menu payload: order must be number');
  }
  if (typeof visible !== 'boolean') {
    throw new Error('invalid menu payload: visible must be boolean');
  }

  return {
    id,
    group,
    title,
    targetType,
    targetRef,
    source,
    order,
    visible,
  };
}

export function parseMenusResponse(value: unknown): MenusResponse {
  if (!isObject(value)) {
    throw new Error('invalid menus response: not an object');
  }

  const { cluster, menus } = value;

  if (typeof cluster !== 'string' || cluster === '') {
    throw new Error('invalid menus response: cluster is required');
  }
  if (!Array.isArray(menus)) {
    throw new Error('invalid menus response: menus must be array');
  }

  return {
    cluster,
    menus: menus.map(parseMenuItem),
  };
}

export function parseClustersResponse(value: unknown): ClustersResponse {
  if (!isObject(value)) {
    throw new Error('invalid clusters response: not an object');
  }

  const { clusters } = value;
  if (!Array.isArray(clusters) || !clusters.every((item) => typeof item === 'string')) {
    throw new Error('invalid clusters response: clusters must be string[]');
  }

  return { clusters };
}

export function parseRegistryResponse(value: unknown): RegistryResponse {
  if (!isObject(value)) {
    throw new Error('invalid registry response: not an object');
  }

  const { cluster, resourceTypes } = value;
  if (typeof cluster !== 'string' || cluster === '') {
    throw new Error('invalid registry response: cluster is required');
  }
  if (!Array.isArray(resourceTypes)) {
    throw new Error('invalid registry response: resourceTypes must be array');
  }

  for (const item of resourceTypes) {
    if (!isObject(item)) {
      throw new Error('invalid registry response: resourceType item must be object');
    }
    if (
      typeof item.id !== 'string' ||
      typeof item.group !== 'string' ||
      typeof item.version !== 'string' ||
      typeof item.kind !== 'string' ||
      typeof item.plural !== 'string' ||
      typeof item.namespaced !== 'boolean' ||
      typeof item.preferredVersion !== 'string' ||
      typeof item.source !== 'string'
    ) {
      throw new Error('invalid registry response: malformed resourceType');
    }
  }

  return {
    cluster,
    resourceTypes: resourceTypes as RegistryResponse['resourceTypes'],
  };
}

export function parseWorkloadsResponse(value: unknown): WorkloadsResponse {
  if (!isObject(value)) {
    throw new Error('invalid workloads response: not an object');
  }

  const { cluster, namespace, items } = value;
  if (typeof cluster !== 'string' || cluster === '') {
    throw new Error('invalid workloads response: cluster is required');
  }
  if (typeof namespace !== 'string') {
    throw new Error('invalid workloads response: namespace must be string');
  }
  if (!Array.isArray(items)) {
    throw new Error('invalid workloads response: items must be array');
  }

  for (const item of items) {
    if (!isObject(item)) {
      throw new Error('invalid workloads response: item must be object');
    }
    if (
      typeof item.id !== 'string' ||
      typeof item.name !== 'string' ||
      typeof item.kind !== 'string' ||
      typeof item.namespace !== 'string' ||
      typeof item.status !== 'string' ||
      typeof item.health !== 'string' ||
      typeof item.updatedAt !== 'string'
    ) {
      throw new Error('invalid workloads response: malformed item');
    }
  }

  return {
    cluster,
    namespace,
    items: items as WorkloadsResponse['items'],
  };
}
