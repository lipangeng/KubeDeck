import { describe, expect, it } from 'vitest';
import {
  resolvePaletteMode,
  sanitizeThemePreference,
  THEME_PREFERENCE_KEY,
} from './themeMode';

describe('themeMode', () => {
  it('sanitizes persisted preference with fallback', () => {
    expect(sanitizeThemePreference('light')).toBe('light');
    expect(sanitizeThemePreference('dark')).toBe('dark');
    expect(sanitizeThemePreference('system')).toBe('system');
    expect(sanitizeThemePreference('unknown')).toBe('system');
  });

  it('resolves palette mode for system preference', () => {
    expect(resolvePaletteMode('system', true)).toBe('dark');
    expect(resolvePaletteMode('system', false)).toBe('light');
  });

  it('exposes stable storage key', () => {
    expect(THEME_PREFERENCE_KEY).toBe('kubedeck.theme.preference');
  });
});
