import { describe, expect, it } from 'vitest';
import { resolveDefaultTabs } from './tabs';

describe('resource page tabs', () => {
  it('resolves overview and yaml as the default tabs', () => {
    expect(resolveDefaultTabs().map((tab) => tab.id)).toEqual(['overview', 'yaml']);
  });
});
