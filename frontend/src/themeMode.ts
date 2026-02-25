export type ThemePreference = 'light' | 'dark' | 'system';

export const THEME_PREFERENCE_KEY = 'kubedeck.theme.preference';

export function sanitizeThemePreference(value: string | null | undefined): ThemePreference {
  if (value === 'light' || value === 'dark' || value === 'system') {
    return value;
  }
  return 'system';
}

export function resolvePaletteMode(
  preference: ThemePreference,
  systemPrefersDark: boolean,
): 'light' | 'dark' {
  if (preference === 'system') {
    return systemPrefersDark ? 'dark' : 'light';
  }
  return preference;
}

export function readStoredThemePreference(): ThemePreference {
  if (typeof window === 'undefined') {
    return 'system';
  }
  return sanitizeThemePreference(window.localStorage.getItem(THEME_PREFERENCE_KEY));
}

export function writeStoredThemePreference(preference: ThemePreference): void {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.setItem(THEME_PREFERENCE_KEY, preference);
}

export function getSystemPrefersDark(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
    return false;
  }
  return window.matchMedia('(prefers-color-scheme: dark)').matches;
}
