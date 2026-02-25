import { createTheme } from '@mui/material/styles';
import type { PaletteMode } from '@mui/material/styles';

export function createAppTheme(mode: PaletteMode) {
  return createTheme({
    palette: {
      mode,
      background:
        mode === 'dark'
          ? {
              default: '#0f1720',
              paper: '#17212d',
            }
          : {
              default: '#f5f7fb',
              paper: '#ffffff',
            },
      primary: {
        main: '#0f4c81',
      },
    },
    shape: {
      borderRadius: 12,
    },
    typography: {
      fontFamily: '"IBM Plex Sans", "Noto Sans SC", sans-serif',
    },
  });
}
