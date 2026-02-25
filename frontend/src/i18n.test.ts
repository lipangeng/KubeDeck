import { describe, expect, it } from 'vitest';
import { sanitizeLocale, translate } from './i18n';

describe('i18n', () => {
  it('sanitizes locale values', () => {
    expect(sanitizeLocale('zh')).toBe('zh');
    expect(sanitizeLocale('en')).toBe('en');
    expect(sanitizeLocale('other')).toBe('en');
  });

  it('renders translated messages with variables', () => {
    expect(translate('en', 'clusterOverview', { cluster: 'dev' })).toBe('Cluster dev Overview');
    expect(translate('zh', 'clusterOverview', { cluster: 'dev' })).toBe('集群 dev 总览');
  });
});

