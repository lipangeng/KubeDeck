import { useState } from 'react';
import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import { copy } from './i18n/copy';
import { useKernelRuntime } from './kernel/runtime/KernelRuntimeContext';
import { AIFloatingChat } from './features/ai/AIFloatingChat';
import { UserMenu } from './components/UserMenu';
import { type ThemePreference } from './themeMode';

interface AppShellProps {
  themePreference: ThemePreference;
  onThemePreferenceChange: (next: ThemePreference) => void;
}

type MenuSurface = 'work' | 'system' | 'cluster';

export function AppShell({ themePreference, onThemePreferenceChange }: AppShellProps) {
  const {
    activeActions,
    activeCluster,
    activePage,
    executeAction,
    navigate,
    navigation,
    switchCluster,
  } = useKernelRuntime();
  
  const ActiveComponent = activePage?.component;
  const [menuSurface, setMenuSurface] = useState<MenuSurface>('work');

  const cycleThemePreference = () => {
    onThemePreferenceChange(themePreference === 'system' ? 'light' : themePreference === 'light' ? 'dark' : 'system');
  };

  const cycleCluster = () => {
    const clusters = ['default', 'cluster-a', 'cluster-b'];
    const currentIndex = clusters.indexOf(activeCluster);
    switchCluster(clusters[(currentIndex + 1) % clusters.length]);
  };

  const handleExecuteAction = async (actionId: string) => {
    await executeAction({ actionId, workflowDomainId: activePage?.workflowDomainId || 'workloads', target: { cluster: activeCluster, namespace: 'default', scope: 'namespace' }, input: {} });
  };

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: 'background.default' }}>
      <AppBar position="sticky" color="transparent" elevation={0}>
        <Toolbar sx={{ gap: 2, borderBottom: 1, borderColor: 'divider' }}>
          <Typography variant="h6" sx={{ fontWeight: 700, flexGrow: 1 }}>KubeDeck</Typography>
          <Button variant="outlined" onClick={cycleThemePreference}>{copy('app.themeLabel')}: {themePreference}</Button>
          <Button variant="outlined" onClick={cycleCluster}>Cluster: {activeCluster}</Button>
          <Button variant="outlined" onClick={() => setMenuSurface('system')}>System Settings</Button>
          <UserMenu />
        </Toolbar>
      </AppBar>

      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '280px minmax(0, 1fr)' }, gap: 2, p: 3 }}>
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Stack spacing={1.5}>
            <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>{copy('app.kernelNavigation')}</Typography>
            {menuSurface === 'work' ? null : (
              <Button variant="outlined" onClick={() => setMenuSurface('work')}>Back to Work</Button>
            )}
            {navigation.map((group) => (
              <Stack key={group.key} spacing={1}>
                <Typography variant="caption" sx={{ fontWeight: 700, textTransform: 'uppercase' }}>{group.title.fallback}</Typography>
                {group.entries.map((entry) => (
                  <Button
                    key={entry.identity.contributionId}
                    variant={entry.route === activePage?.route && menuSurface === 'work' ? 'contained' : 'outlined'}
                    disabled={entry.availability !== 'enabled'}
                    onClick={() => { if (menuSurface === 'work') navigate(entry.route ?? '/'); }}
                  >
                    {entry.title.fallback}
                  </Button>
                ))}
              </Stack>
            ))}
            {menuSurface === 'work' && (
              <>
                <Divider />
                <Button variant="contained" color="secondary" onClick={() => setMenuSurface('cluster')}>Cluster Settings</Button>
              </>
            )}
            <Divider />
            {activeActions.map((action) => (
              <Button key={action.identity.contributionId} variant="text" onClick={() => void handleExecuteAction(action.actionId)}>
                {action.title.fallback}
              </Button>
            ))}
          </Stack>
        </Paper>

        <Box>
          {menuSurface === 'work' && ActiveComponent ? (
            <ActiveComponent />
          ) : (
            <Paper variant="outlined" sx={{ p: 3 }}>
              <Typography variant="h6">Settings</Typography>
              <Typography variant="body2" color="text.secondary">Menu surface: {menuSurface}</Typography>
            </Paper>
          )}
        </Box>
      </Box>

      <AIFloatingChat />
    </Box>
  );
}
