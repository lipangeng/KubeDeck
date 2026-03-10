import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import { copy } from './i18n/copy';
import type { FrontendCapabilityModule } from './kernel/sdk';
import { KernelRuntimeProvider, useKernelRuntime } from './kernel/runtime/KernelRuntimeContext';
import { type ThemePreference } from './themeMode';

interface AppProps {
  themePreference: ThemePreference;
  onThemePreferenceChange: (next: ThemePreference) => void;
  pluginModules?: FrontendCapabilityModule[];
}

function App({ themePreference, onThemePreferenceChange, pluginModules = [] }: AppProps) {
  return (
    <KernelRuntimeProvider pluginModules={pluginModules}>
      <AppShell
        themePreference={themePreference}
        onThemePreferenceChange={onThemePreferenceChange}
      />
    </KernelRuntimeProvider>
  );
}

function AppShell({ themePreference, onThemePreferenceChange }: AppProps) {
  const { activeActions, activePage, actionSummary, executeAction, kernelSource, navigate, navigation } =
    useKernelRuntime();
  const ActiveComponent = activePage?.component;

  const cycleThemePreference = () => {
    const nextPreference: ThemePreference =
      themePreference === 'system'
        ? 'light'
        : themePreference === 'light'
          ? 'dark'
          : 'system';
    onThemePreferenceChange(nextPreference);
  };

  const handleExecuteAction = async (actionId: string) => {
    await executeAction({
      actionId,
      workflowDomainId: activePage?.workflowDomainId ?? 'homepage',
      target: {
        cluster: 'default',
        namespace: 'default',
        scope: 'namespace',
      },
      input: {
        name: 'sample',
      },
    });
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

      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: { xs: '1fr', md: '280px minmax(0, 1fr)' },
          gap: 2,
          px: { xs: 2, md: 3 },
          py: 3,
        }}
      >
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Stack spacing={1.5}>
            <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
              {copy('app.kernelNavigation')}
            </Typography>
            {navigation.map((entry) => (
              <Button
                key={entry.identity.contributionId}
                variant={entry.route === activePage?.route ? 'contained' : 'outlined'}
                onClick={() => navigate(entry.route ?? '/')}
              >
                {entry.title.fallback}
              </Button>
            ))}
            <Divider />
            <Typography variant="caption" color="text.secondary">
              {copy('app.registeredActions')}:{' '}
              {activeActions.map((action) => action.title.fallback).join(', ') || copy('app.none')}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {copy('app.kernelMetadataSource')}: {kernelSource}
            </Typography>
            {activeActions.map((action) => (
              <Button
                key={action.identity.contributionId}
                variant="text"
                onClick={() => {
                  void handleExecuteAction(action.actionId);
                }}
              >
                {copy('app.runAction')} {action.title.fallback}
              </Button>
            ))}
            {actionSummary ? (
              <Typography variant="caption" color="text.secondary">
                {copy('app.lastActionResult')}: {actionSummary}
              </Typography>
            ) : null}
          </Stack>
        </Paper>

        <Box>{ActiveComponent ? <ActiveComponent /> : null}</Box>
      </Box>
    </Box>
  );
}

export default App;
