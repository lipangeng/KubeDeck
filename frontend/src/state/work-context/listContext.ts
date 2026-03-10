import type { ListContext } from './types';

export function createListContext(): ListContext {
  return {};
}

export function updateListContext(
  current: ListContext,
  next: Partial<ListContext>,
): ListContext {
  return {
    ...current,
    ...next,
  };
}

export function resetListContext(): ListContext {
  return {};
}
