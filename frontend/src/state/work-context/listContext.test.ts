import { describe, expect, it } from 'vitest';
import {
  createListContext,
  resetListContext,
  updateListContext,
} from './listContext';

describe('listContext', () => {
  it('starts empty', () => {
    expect(createListContext()).toEqual({});
  });

  it('updates continuity-relevant list state', () => {
    const current = createListContext();
    expect(
      updateListContext(current, {
        searchText: 'api',
        statusFilters: ['Running'],
      }),
    ).toEqual({
      searchText: 'api',
      statusFilters: ['Running'],
    });
  });

  it('resets list context', () => {
    expect(resetListContext()).toEqual({});
  });
});
