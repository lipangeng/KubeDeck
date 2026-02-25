import React, { useEffect, useMemo, useState } from 'react';
import ReactDOM from 'react-dom/client';
import CssBaseline from '@mui/material/CssBaseline';
import { ThemeProvider } from '@mui/material/styles';
import App from './App';
import { createAppTheme } from './theme';
import {
  getSystemPrefersDark,
  readStoredThemePreference,
  resolvePaletteMode,
  type ThemePreference,
  writeStoredThemePreference,
} from './themeMode';

function Root() {
  const [themePreference, setThemePreference] = useState<ThemePreference>(() =>
    readStoredThemePreference(),
  );
  const [systemPrefersDark, setSystemPrefersDark] = useState<boolean>(() =>
    getSystemPrefersDark(),
  );

  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
      return;
    }

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const onChange = (event: MediaQueryListEvent) => {
      setSystemPrefersDark(event.matches);
    };
    mediaQuery.addEventListener('change', onChange);
    return () => {
      mediaQuery.removeEventListener('change', onChange);
    };
  }, []);

  const paletteMode = resolvePaletteMode(themePreference, systemPrefersDark);
  const theme = useMemo(() => createAppTheme(paletteMode), [paletteMode]);

  const handleThemePreferenceChange = (next: ThemePreference) => {
    setThemePreference(next);
    writeStoredThemePreference(next);
  };

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <App
        themePreference={themePreference}
        onThemePreferenceChange={handleThemePreferenceChange}
      />
    </ThemeProvider>
  );
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Root />
  </React.StrictMode>,
);
