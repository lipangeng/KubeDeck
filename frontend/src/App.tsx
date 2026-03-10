import { useEffect, useMemo, useState } from 'react';
import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import { copy } from './i18n/copy';
import { executeKernelAction } from './kernel/runtime/executeKernelAction';
import { registerBuiltInActions } from './kernel/builtins/registerBuiltInActions';
import { registerBuiltInMenus } from './kernel/builtins/registerBuiltInMenus';
import { registerBuiltInPages } from './kernel/builtins/registerBuiltInPages';
import { registerBuiltInSlots } from './kernel/builtins/registerBuiltInSlots';
import { composeKernelNavigation } from './kernel/runtime/composeKernelNavigation';
import { fetchKernelMetadata } from './kernel/runtime/fetchKernelMetadata';
import { hydrateKernelSnapshot } from './kernel/runtime/hydrateKernelSnapshot';
import { KernelRegistry } from './kernel/runtime/kernelRegistry';
import { resolveWorkflowActions } from './kernel/runtime/resolveWorkflowActions';
import type { KernelRegistrySnapshot } from './kernel/runtime/types';
import { type ThemePreference } from './themeMode';

interface AppProps {
  themePreference: ThemePreference;
  onThemePreferenceChange: (next: ThemePreference) => void;
}

function App({ themePreference, onThemePreferenceChange }: AppProps) {
  const [activeRoute, setActiveRoute] = useState('/');
  const [runtimeSnapshot, setRuntimeSnapshot] = useState<KernelRegistrySnapshot | null>(null);
  const [kernelSource, setKernelSource] = useState<'loading' | 'backend' | 'local-fallback'>(
    'loading',
  );
  const [actionSummary, setActionSummary] = useState<string | null>(null);

  const cycleThemePreference = () => {
    const nextPreference: ThemePreference =
      themePreference === 'system'
        ? 'light'
        : themePreference === 'light'
          ? 'dark'
          : 'system';
    onThemePreferenceChange(nextPreference);
  };

  const localSnapshot = useMemo(() => {
    const registry = new KernelRegistry();
    registry.register({
      pages: registerBuiltInPages(),
      menus: registerBuiltInMenus(),
      actions: registerBuiltInActions(),
      slots: registerBuiltInSlots(),
    });
    return registry.snapshot();
  }, []);

  useEffect(() => {
    let active = true;

    async function loadKernelMetadata() {
      try {
        const remoteMetadata = await fetchKernelMetadata();
        if (!active) {
          return;
        }
        setRuntimeSnapshot(hydrateKernelSnapshot(localSnapshot, remoteMetadata));
        setKernelSource('backend');
      } catch {
        if (!active) {
          return;
        }
        setRuntimeSnapshot(localSnapshot);
        setKernelSource('local-fallback');
      }
    }

    void loadKernelMetadata();
    return () => {
      active = false;
    };
  }, [localSnapshot]);

  const registrySnapshot = runtimeSnapshot ?? localSnapshot;

  const navigation = useMemo(
    () => composeKernelNavigation(registrySnapshot.menus),
    [registrySnapshot.menus],
  );

  const activePage =
    registrySnapshot.pages.find((page) => page.route === activeRoute) ?? registrySnapshot.pages[0];
  const activeActions = activePage
    ? resolveWorkflowActions(activePage.workflowDomainId, registrySnapshot.actions)
    : [];
  const ActiveComponent = activePage?.component;

  const handleExecuteAction = async (actionId: string) => {
    const result = await executeKernelAction({
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
    setActionSummary(result.Summary);
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
              Kernel Navigation
            </Typography>
            {navigation.map((entry) => (
              <Button
                key={entry.identity.contributionId}
                variant={entry.route === activeRoute ? 'contained' : 'outlined'}
                onClick={() => setActiveRoute(entry.route ?? '/')}
              >
                {entry.title.fallback}
              </Button>
            ))}
            <Divider />
            <Typography variant="caption" color="text.secondary">
              Registered actions: {activeActions.map((action) => action.title.fallback).join(', ') || 'None'}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              Kernel metadata source: {kernelSource}
            </Typography>
            {activeActions.map((action) => (
              <Button
                key={action.identity.contributionId}
                variant="text"
                onClick={() => {
                  void handleExecuteAction(action.actionId);
                }}
              >
                Run {action.title.fallback}
              </Button>
            ))}
            {actionSummary ? (
              <Typography variant="caption" color="text.secondary">
                Last action result: {actionSummary}
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
