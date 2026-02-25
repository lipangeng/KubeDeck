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
import { composeMenus } from './core/menuComposer';
import { groupMenusBySource } from './core/menuGrouping';
import {
  parseClustersResponse,
  parseMenusResponse,
  parseRegistryResponse,
} from './sdk/metaApi';
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

function statusColor(status: ProbeStatus): 'success' | 'warning' | 'error' {
  if (status === 'ok') {
    return 'success';
  }
  if (status === 'checking') {
    return 'warning';
  }
  return 'error';
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
  const [clusters, setClusters] = useState<string[]>(['default']);
  const [menus, setMenus] = useState<MenuItem[]>([]);
  const [resourceTypeCount, setResourceTypeCount] = useState(0);
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

    async function loadClusters() {
      try {
        const response = await fetch('/api/meta/clusters');
        if (!response.ok) {
          throw new Error(`clusters request failed: ${response.status}`);
        }
        const payload = parseClustersResponse(await response.json());
        if (active && payload.clusters.length > 0) {
          setClusters(payload.clusters);
          if (!payload.clusters.includes(activeCluster)) {
            setActiveCluster(payload.clusters[0]);
          }
        }
      } catch {
        if (active) {
          setClusters(['default']);
        }
      }
    }

    loadClusters();
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    let active = true;

    async function loadClusterMetadata() {
      setLoading(true);
      setError(null);
      setMenus([]);
      setResourceTypeCount(0);
      try {
        const [menusResponse, registryResponse] = await Promise.all([
          fetch(`/api/meta/menus?cluster=${encodeURIComponent(activeCluster)}`),
          fetch(`/api/meta/registry?cluster=${encodeURIComponent(activeCluster)}`),
        ]);
        if (!menusResponse.ok) {
          throw new Error(`menus request failed: ${menusResponse.status}`);
        }
        if (!registryResponse.ok) {
          throw new Error(`registry request failed: ${registryResponse.status}`);
        }
        const menusPayload = parseMenusResponse(await menusResponse.json());
        const registryPayload = parseRegistryResponse(await registryResponse.json());
        if (active) {
          setMenus(menusPayload.menus);
          setResourceTypeCount(registryPayload.resourceTypes.length);
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

    loadClusterMetadata();

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

  const failureSummary =
    healthError || readyError
      ? [
          healthError ? `healthz: ${healthError}` : null,
          readyError ? `readyz: ${readyError}` : null,
        ]
          .filter(Boolean)
          .join('; ')
      : 'none';

  return (
    <Box
      sx={{
        minHeight: '100vh',
        background:
          'radial-gradient(1400px 500px at 10% -10%, rgba(25,118,210,0.18), transparent 65%), radial-gradient(1200px 460px at 95% -20%, rgba(56,142,60,0.14), transparent 60%)',
        bgcolor: 'background.default',
      }}
    >
      <AppBar
        position="sticky"
        color="transparent"
        elevation={0}
        sx={{
          backdropFilter: 'blur(10px)',
          borderBottom: 1,
          borderColor: 'divider',
          backgroundColor: 'rgba(255,255,255,0.72)',
        }}
      >
        <Toolbar sx={{ gap: 1.5, minHeight: 72 }}>
          <Typography variant="h5" component="h1" sx={{ fontWeight: 800, pr: 1.5 }}>
            KubeDeck
          </Typography>

          <Paper
            variant="outlined"
            sx={{
              px: 1.2,
              py: 0.9,
              display: 'flex',
              alignItems: 'center',
              gap: 1,
              flexWrap: 'wrap',
              borderRadius: 2,
              bgcolor: 'background.paper',
            }}
          >
            <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 600 }}>
              Runtime
            </Typography>
            <Chip
              size="small"
              color={statusColor(healthStatus)}
              variant="filled"
              label={`healthz: ${healthStatus}`}
            />
            <Chip
              size="small"
              color={statusColor(readyStatus)}
              variant="filled"
              label={`readyz: ${readyStatus}`}
            />
            <Typography variant="caption" color="text.secondary">
              Last checked: {lastCheckedAt ?? 'never'}
            </Typography>
          </Paper>

          <Box sx={{ flexGrow: 1 }} />

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
          gridTemplateColumns: { xs: '1fr', md: '320px minmax(0, 1fr)' },
          gap: 2,
          p: 2,
        }}
      >
        <Paper
          component="nav"
          aria-label="Primary Sidebar"
          elevation={2}
          sx={{
            p: 2,
            minHeight: { md: 'calc(100vh - 120px)' },
            border: 1,
            borderColor: 'divider',
          }}
        >
          <Stack spacing={2}>
            <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
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
                {clusters.map((clusterId) => (
                  <option key={clusterId} value={clusterId}>
                    {clusterId}
                  </option>
                ))}
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
          <Paper elevation={3} sx={{ p: 2.2, border: 1, borderColor: 'divider' }}>
            <Typography variant="overline" color="primary.main" sx={{ letterSpacing: 1.1 }}>
              Control Plane
            </Typography>
            <Typography variant="h5" sx={{ fontWeight: 700, mb: 0.8 }}>
              Cluster {activeCluster} Overview
            </Typography>
            <Typography color="text.secondary">
              API target ({apiTargetHint})
            </Typography>
          </Paper>

          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: { xs: '1fr', sm: 'repeat(3, minmax(0, 1fr))' },
              gap: 1.5,
            }}
          >
            <Paper elevation={1} sx={{ p: 1.6, border: 1, borderColor: 'divider' }}>
              <Typography variant="caption" color="text.secondary">
                Registry Resource Types
              </Typography>
              <Typography variant="h4" sx={{ fontWeight: 700 }} data-testid="registry-resource-type-count">
                {resourceTypeCount}
              </Typography>
            </Paper>
            <Paper elevation={1} sx={{ p: 1.6, border: 1, borderColor: 'divider' }}>
              <Typography variant="caption" color="text.secondary">
                Health Endpoint
              </Typography>
              <Typography variant="h6" sx={{ fontWeight: 700 }}>
                {healthStatus}
              </Typography>
            </Paper>
            <Paper elevation={1} sx={{ p: 1.6, border: 1, borderColor: 'divider' }}>
              <Typography variant="caption" color="text.secondary">
                Readiness Endpoint
              </Typography>
              <Typography variant="h6" sx={{ fontWeight: 700 }}>
                {readyStatus}
              </Typography>
            </Paper>
          </Box>

          <Paper elevation={1} sx={{ p: 1.6, border: 1, borderColor: 'divider' }}>
            <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 0.4 }}>
              Failure summary
            </Typography>
            <Typography variant="body2" color={failureSummary === 'none' ? 'text.secondary' : 'error'}>
              Failure summary: {failureSummary}
            </Typography>
          </Paper>
        </Stack>
      </Box>
    </Box>
  );
}

export default App;
