import type {
  ApplyResponse,
  ClustersResponse,
  MenuItem,
  MenusResponse,
  MenuSource,
  RegistryResponse,
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

export function parseApplyResponse(value: unknown): ApplyResponse {
  if (!isObject(value)) {
    throw new Error('invalid apply response: not an object');
  }

  const {
    status,
    cluster,
    defaultNamespace,
    total,
    succeeded,
    failed,
    results,
  } = value;

  if (status !== 'success' && status !== 'partial' && status !== 'failed') {
    throw new Error('invalid apply response: status is invalid');
  }
  if (typeof cluster !== 'string' || cluster === '') {
    throw new Error('invalid apply response: cluster is required');
  }
  if (typeof defaultNamespace !== 'string' || defaultNamespace === '') {
    throw new Error('invalid apply response: defaultNamespace is required');
  }
  if (
    typeof total !== 'number' ||
    typeof succeeded !== 'number' ||
    typeof failed !== 'number'
  ) {
    throw new Error('invalid apply response: counters must be numbers');
  }
  if (!Array.isArray(results)) {
    throw new Error('invalid apply response: results must be array');
  }

  for (const result of results) {
    if (!isObject(result)) {
      throw new Error('invalid apply response: result item must be object');
    }
    if (
      typeof result.index !== 'number' ||
      typeof result.kind !== 'string' ||
      typeof result.name !== 'string' ||
      typeof result.namespace !== 'string' ||
      (result.status !== 'succeeded' && result.status !== 'failed')
    ) {
      throw new Error('invalid apply response: malformed result item');
    }
    if (
      result.reason !== undefined &&
      result.reason !== null &&
      typeof result.reason !== 'string'
    ) {
      throw new Error('invalid apply response: reason must be string when present');
    }
  }

  return {
    status,
    cluster,
    defaultNamespace,
    total,
    succeeded,
    failed,
    results: results.map((result) => ({
      index: result.index as number,
      kind: result.kind as string,
      name: result.name as string,
      namespace: result.namespace as string,
      status: result.status as 'succeeded' | 'failed',
      reason: result.reason as string | undefined,
    })),
  };
}
