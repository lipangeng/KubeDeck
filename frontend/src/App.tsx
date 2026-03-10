import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import { copy } from './i18n/copy';
import { type ThemePreference } from './themeMode';

interface AppProps {
  themePreference: ThemePreference;
  onThemePreferenceChange: (next: ThemePreference) => void;
}

function App({ themePreference, onThemePreferenceChange }: AppProps) {
  const cycleThemePreference = () => {
    const nextPreference: ThemePreference =
      themePreference === 'system'
        ? 'light'
        : themePreference === 'light'
          ? 'dark'
          : 'system';
    onThemePreferenceChange(nextPreference);
  };

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: 'background.default', color: 'text.primary' }}>
      <AppBar position="sticky" color="transparent" elevation={0}>
        <Toolbar sx={{ gap: 2, borderBottom: 1, borderColor: 'divider' }}>
          <Typography variant="h6" sx={{ fontWeight: 700, flexGrow: 1 }}>
            {copy('app.title')}
          </Typography>
          <Button variant="outlined" onClick={cycleThemePreference}>
            {copy('app.themeLabel')}: {themePreference}
          </Button>
        </Toolbar>
      </AppBar>

      <Box sx={{ px: { xs: 2, md: 3 }, py: 3 }}>
        <Paper variant="outlined" sx={{ p: 3 }}>
          <Stack spacing={1.5}>
            <Typography variant="overline" color="primary.main">
              {copy('app.cleanup.badge')}
            </Typography>
            <Typography variant="h4" sx={{ fontWeight: 700 }}>
              {copy('app.cleanup.title')}
            </Typography>
            <Typography color="text.secondary">
              {copy('app.cleanup.body')}
            </Typography>
          </Stack>
        </Paper>
      </Box>
    </Box>
  );
}

export default App;
