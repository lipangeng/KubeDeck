import { createTheme } from '@mui/material/styles';
import type { PaletteMode } from '@mui/material/styles';

export function createAppTheme(mode: PaletteMode) {
  const isDark = mode === 'dark';
  return createTheme({
    palette: {
      mode,
      primary: {
        main: isDark ? '#7cc8ff' : '#0067d6',
        light: isDark ? '#a6dcff' : '#3c8cff',
        dark: isDark ? '#4fa8f3' : '#004a9a',
      },
      secondary: {
        main: isDark ? '#6fe7ff' : '#00a2c7',
      },
      background: isDark
        ? {
            default: '#060c18',
            paper: 'rgba(15, 26, 48, 0.62)',
          }
        : {
            default: '#edf3ff',
            paper: 'rgba(255, 255, 255, 0.72)',
          },
      text: isDark
        ? {
            primary: '#eaf4ff',
            secondary: '#9fb4cc',
          }
        : {
            primary: '#0b1b34',
            secondary: '#446487',
          },
    },
    shape: {
      borderRadius: 16,
    },
    typography: {
      fontFamily: '"IBM Plex Sans", "Noto Sans SC", sans-serif',
      h5: {
        fontWeight: 700,
        letterSpacing: 0.2,
      },
      subtitle2: {
        fontWeight: 600,
      },
    },
    components: {
      MuiPaper: {
        styleOverrides: {
          root: {
            backgroundImage: 'none',
            backdropFilter: 'blur(14px)',
            border: `1px solid ${isDark ? 'rgba(124,200,255,0.2)' : 'rgba(33,106,196,0.18)'}`,
            boxShadow: isDark
              ? '0 16px 36px rgba(3, 10, 22, 0.5)'
              : '0 16px 30px rgba(38, 90, 168, 0.14)',
          },
        },
      },
      MuiAppBar: {
        styleOverrides: {
          root: {
            backgroundColor: isDark ? 'rgba(6, 14, 30, 0.62)' : 'rgba(242, 248, 255, 0.7)',
            borderBottom: `1px solid ${isDark ? 'rgba(124,200,255,0.22)' : 'rgba(33,106,196,0.18)'}`,
          },
        },
      },
      MuiButton: {
        styleOverrides: {
          contained: {
            boxShadow: 'none',
            borderRadius: 12,
            background: isDark
              ? 'linear-gradient(135deg, #5ab8ff 0%, #2d8ce8 100%)'
              : 'linear-gradient(135deg, #2f8dff 0%, #126ddf 100%)',
          },
        },
      },
      MuiChip: {
        styleOverrides: {
          root: {
            fontWeight: 600,
            borderRadius: 10,
          },
        },
      },
      MuiOutlinedInput: {
        styleOverrides: {
          root: {
            backgroundColor: isDark ? 'rgba(11, 20, 36, 0.52)' : 'rgba(255, 255, 255, 0.66)',
          },
        },
      },
    },
  });
}
