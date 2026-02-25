import { describe, expect, it } from 'vitest';
import {
  ALL_NAMESPACES,
  resolveCreateDefaultNamespace,
} from './namespaceFilter';

describe('resolveCreateDefaultNamespace', () => {
  it('uses list filter namespace when a specific namespace is selected', () => {
    expect(
      resolveCreateDefaultNamespace({
        listFilterNamespace: 'kube-system',
        lastUsedNamespace: 'default',
      }),
    ).toBe('kube-system');
  });

  it('uses last used namespace when list filter is All', () => {
    expect(
      resolveCreateDefaultNamespace({
        listFilterNamespace: ALL_NAMESPACES,
        lastUsedNamespace: 'team-a',
      }),
    ).toBe('team-a');
  });

  it('falls back to default when list filter is All and no last used namespace', () => {
    expect(
      resolveCreateDefaultNamespace({
        listFilterNamespace: ALL_NAMESPACES,
      }),
    ).toBe('default');
  });
});
