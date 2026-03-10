import { describe, expect, it } from 'vitest';
import {
  createNamespaceScope,
  resetNamespaceScopeForCluster,
  restoreNamespaceScope,
  updateNamespaceScope,
} from './namespaceContext';

describe('namespaceContext', () => {
  it('creates a single namespace scope by default', () => {
    expect(createNamespaceScope()).toEqual({
      mode: 'single',
      values: ['default'],
      source: 'default',
    });
  });

  it('normalizes all-namespace scope to an empty values list', () => {
    expect(createNamespaceScope('all', ['default'])).toEqual({
      mode: 'all',
      values: [],
      source: 'default',
    });
  });

  it('restores scope with restored source', () => {
    expect(
      restoreNamespaceScope({
        mode: 'single',
        values: ['team-a'],
        source: 'user_selected',
      }),
    ).toEqual({
      mode: 'single',
      values: ['team-a'],
      source: 'restored',
    });
  });

  it('marks user-selected scope updates explicitly', () => {
    const current = createNamespaceScope();
    expect(
      updateNamespaceScope(current, {
        mode: 'single',
        values: ['kube-system'],
        source: 'default',
      }),
    ).toEqual({
      mode: 'single',
      values: ['kube-system'],
      source: 'user_selected',
    });
  });

  it('rejects invalid single and multiple namespace scope shapes', () => {
    expect(() =>
      createNamespaceScope('single', ['team-a', 'team-b']),
    ).toThrow('single namespace scope must contain exactly one value');
    expect(() => createNamespaceScope('multiple', ['team-a'])).toThrow(
      'multiple namespace scope must contain more than one value',
    );
  });

  it('resets cluster scope to default single namespace', () => {
    expect(resetNamespaceScopeForCluster()).toEqual({
      mode: 'single',
      values: ['default'],
      source: 'default',
    });
  });
});
