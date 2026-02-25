import type { MenuItem, MenusResponse, MenuSource } from './types';

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
