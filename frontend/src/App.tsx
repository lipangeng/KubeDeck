import { useEffect, useMemo, useState } from 'react';
import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import Paper from '@mui/material/Paper';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import { ListPageShell } from './components/page-shell/ResourcePageShell';
import { composeMenus } from './core/menuComposer';
import { groupMenusBySource } from './core/menuGrouping';
import { parseMenusResponse } from './sdk/metaApi';
import type { MenuItem } from './sdk/types';
import type { ThemePreference } from './themeMode';

type ProbeStatus = 'checking' | 'ok' | 'error';

interface AppProps {
  themePreference: ThemePreference;
  onThemePreferenceChange: (next: ThemePreference) => void;
}

function resolveApiTarget(): string {
  const configured = import.meta.env.VITE_BACKEND_TARGET as string | undefined;
  if (configured && configured.trim() !== '') {
    return configured.replace(/\/$/, '');
  }

  return 'http://127.0.0.1:8080';
}

function resolveApiTargetHint(): string {
  const mode = import.meta.env.MODE;
  if (mode === 'development' || mode === 'test') {
    return `${mode}: ${resolveApiTarget()}`;
  }

  return mode;
}

function resolveProbePath(path: '/healthz' | '/readyz'): string {
  return `/api${path}`;
}

function MenuSection({
  title,
  items,
}: {
  title: string;
  items: MenuItem[];
}) {
  return (
    <Box>
      <Typography variant="subtitle2" color="text.secondary" sx={{ mb: 0.5 }}>
        {title}
      </Typography>
      <List dense sx={{ pt: 0 }}>
        {items.length === 0 ? (
          <ListItem disablePadding>
            <ListItemText primary="No entries" primaryTypographyProps={{ color: 'text.disabled' }} />
          </ListItem>
        ) : (
          items.map((menu) => (
            <ListItem key={menu.id} disablePadding>
              <ListItemText primary={menu.title} />
            </ListItem>
          ))
        )}
      </List>
    </Box>
  );
}

function App({ themePreference, onThemePreferenceChange }: AppProps) {
  const apiTargetHint = resolveApiTargetHint();
  const [activeCluster, setActiveCluster] = useState('default');
  const [menus, setMenus] = useState<MenuItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [healthStatus, setHealthStatus] = useState<ProbeStatus>('checking');
  const [readyStatus, setReadyStatus] = useState<ProbeStatus>('checking');
  const [healthError, setHealthError] = useState<string | null>(null);
  const [readyError, setReadyError] = useState<string | null>(null);
  const [lastCheckedAt, setLastCheckedAt] = useState<string | null>(null);

  const groupedMenus = useMemo(() => {
    const systemMenus = menus.filter((menu) => menu.source === 'system');
    const userMenus = menus.filter((menu) => menu.source === 'user');
    const dynamicMenus = menus.filter((menu) => menu.source === 'dynamic');
    const merged = composeMenus(systemMenus, userMenus, dynamicMenus);
    return groupMenusBySource(merged);
  }, [menus]);

  useEffect(() => {
    let active = true;

    async function loadMenus() {
      setLoading(true);
      setError(null);
      try {
        const response = await fetch(
          `/api/meta/menus?cluster=${encodeURIComponent(activeCluster)}`,
        );
        if (!response.ok) {
          throw new Error(`menus request failed: ${response.status}`);
        }

        const payload = parseMenusResponse(await response.json());
        if (active) {
          setMenus(payload.menus);
        }
      } catch (e) {
        if (active) {
          setError(e instanceof Error ? e.message : 'unknown error');
          setMenus([]);
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    }

    loadMenus();

    return () => {
      active = false;
    };
  }, [activeCluster]);

  useEffect(() => {
    let active = true;

    async function checkEndpoint(
      path: '/healthz' | '/readyz',
    ): Promise<{ status: ProbeStatus; error: string | null }> {
      try {
        const response = await fetch(resolveProbePath(path));
        if (!response.ok) {
          throw new Error(`status ${response.status}`);
        }
        return { status: 'ok', error: null };
      } catch (e) {
        if (e instanceof Error) {
          return { status: 'error', error: e.message };
        }
        return { status: 'error', error: 'network error' };
      }
    }

    async function probeRuntime() {
      if (active) {
        setHealthStatus('checking');
        setReadyStatus('checking');
        setHealthError(null);
        setReadyError(null);
      }
      const [healthResult, readyResult] = await Promise.all([
        checkEndpoint('/healthz'),
        checkEndpoint('/readyz'),
      ]);
      if (active) {
        setHealthStatus(healthResult.status);
        setReadyStatus(readyResult.status);
        setHealthError(healthResult.error);
        setReadyError(readyResult.error);
        setLastCheckedAt(new Date().toISOString());
      }
    }

    probeRuntime();
    const timer = setInterval(probeRuntime, 10_000);

    return () => {
      active = false;
      clearInterval(timer);
    };
  }, []);

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: 'background.default' }}>
      <AppBar position="static" color="transparent" elevation={0}>
        <Toolbar sx={{ borderBottom: 1, borderColor: 'divider', gap: 2 }}>
          <Typography variant="h5" component="h1" sx={{ fontWeight: 700, flexGrow: 1 }}>
            KubeDeck
          </Typography>
          <FormControl size="small" sx={{ minWidth: 160 }}>
            <InputLabel htmlFor="theme-select">Theme</InputLabel>
            <Select
              native
              value={themePreference}
              onChange={(event) =>
                onThemePreferenceChange(event.target.value as ThemePreference)
              }
              label="Theme"
              inputProps={{ id: 'theme-select' }}
            >
              <option value="system">System</option>
              <option value="light">Light</option>
              <option value="dark">Dark</option>
            </Select>
          </FormControl>
        </Toolbar>
      </AppBar>

      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: { xs: '1fr', md: '300px minmax(0, 1fr)' },
          gap: 2,
          p: 2,
        }}
      >
        <Paper
          component="nav"
          aria-label="Primary Sidebar"
          variant="outlined"
          sx={{ p: 2, minHeight: { md: 'calc(100vh - 110px)' } }}
        >
          <Stack spacing={2}>
            <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
              Navigation
            </Typography>
            <FormControl size="small" fullWidth>
              <InputLabel htmlFor="cluster-select">Cluster</InputLabel>
              <Select
                native
                value={activeCluster}
                onChange={(event) => setActiveCluster(event.target.value)}
                label="Cluster"
                inputProps={{ id: 'cluster-select' }}
              >
                <option value="default">default</option>
                <option value="dev">dev</option>
                <option value="staging">staging</option>
                <option value="prod">prod</option>
              </Select>
            </FormControl>

            <Divider />

            {loading ? <Typography>Loading menus...</Typography> : null}
            {error ? (
              <Typography color="error">Failed to load menus: {error}</Typography>
            ) : null}
            <MenuSection title="System Menus" items={groupedMenus.system} />
            <MenuSection title="User Menus" items={groupedMenus.user} />
            <MenuSection title="Dynamic Menus" items={groupedMenus.dynamic} />
          </Stack>
        </Paper>

        <Stack spacing={2}>
          <Box aria-label="Runtime Status">
            <ListPageShell
              title="Runtime Status"
              toolbar={
                <Stack direction="row" spacing={1}>
                  <Chip size="small" label={`healthz: ${healthStatus}`} />
                  <Chip size="small" label={`readyz: ${readyStatus}`} />
                </Stack>
              }
            >
              <Typography variant="body2" sx={{ mb: 0.5 }}>
                API target ({apiTargetHint})
              </Typography>
              <Typography variant="body2" sx={{ mb: 0.5 }}>
                Last checked: {lastCheckedAt ?? 'never'}
              </Typography>
              <Typography variant="body2">
                Failure summary:{' '}
                {healthError || readyError
                  ? [
                      healthError ? `healthz: ${healthError}` : null,
                      readyError ? `readyz: ${readyError}` : null,
                    ]
                      .filter(Boolean)
                      .join('; ')
                  : 'none'}
              </Typography>
            </ListPageShell>
          </Box>
        </Stack>
      </Box>
    </Box>
  );
}

export default App;
