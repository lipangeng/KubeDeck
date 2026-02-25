import { useEffect, useMemo, useState } from 'react';
import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Collapse from '@mui/material/Collapse';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import Paper from '@mui/material/Paper';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import { composeMenus } from './core/menuComposer';
import { groupMenusByGroup } from './core/menuGrouping';
import { translate, type Locale } from './i18n';
import {
  parseApplyResponse,
  parseClustersResponse,
  parseMenusResponse,
  parseRegistryResponse,
} from './sdk/metaApi';
import type { ApplyResultItem, MenuItem, RegistryResourceType } from './sdk/types';
import { ALL_NAMESPACES, resolveCreateDefaultNamespace } from './state/namespaceFilter';
import type { ThemePreference } from './themeMode';

type ProbeStatus = 'checking' | 'ok' | 'error';
const MENU_GROUPS_STORAGE_KEY = 'kubedeck.menu.groups.expanded';

interface AppProps {
  locale: Locale;
  onLocaleChange: (next: Locale) => void;
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

function resolveRuntimeStatus(healthStatus: ProbeStatus, readyStatus: ProbeStatus): ProbeStatus {
  if (healthStatus === 'error' || readyStatus === 'error') {
    return 'error';
  }
  if (healthStatus === 'checking' || readyStatus === 'checking') {
    return 'checking';
  }
  return 'ok';
}

function readStoredExpandedGroups(): Record<string, boolean> {
  if (typeof window === 'undefined') {
    return {};
  }
  const raw = window.localStorage.getItem(MENU_GROUPS_STORAGE_KEY);
  if (!raw) {
    return {};
  }
  try {
    const parsed = JSON.parse(raw) as Record<string, unknown>;
    const sanitized: Record<string, boolean> = {};
    for (const [key, value] of Object.entries(parsed)) {
      if (typeof value === 'boolean') {
        sanitized[key] = value;
      }
    }
    return sanitized;
  } catch {
    return {};
  }
}

function readHashRoute(): string {
  if (typeof window === 'undefined') {
    return '/';
  }
  const route = window.location.hash.replace(/^#/, '').trim();
  return route === '' ? '/' : route;
}

function normalizeRoute(route: string): string {
  const trimmed = route.trim();
  if (trimmed === '') {
    return '/';
  }
  return trimmed.startsWith('/') ? trimmed : `/${trimmed}`;
}

function menuIcon(group: string, targetType: 'page' | 'resource'): string {
  const normalizedGroup = group.trim().toUpperCase();
  if (normalizedGroup === 'WORKLOAD') {
    return '◈';
  }
  if (normalizedGroup === 'FAVORITES') {
    return '★';
  }
  if (targetType === 'resource') {
    return '◉';
  }
  return '•';
}

function countYamlDocuments(yaml: string): number {
  return yaml
    .split(/^---\s*$/m)
    .map((doc) => doc.trim())
    .filter((doc) => doc.length > 0).length;
}

function normalizeResourceGroupName(name: string): string {
  const trimmed = name.trim();
  return trimmed === '' ? 'core' : trimmed;
}

function App({
  locale,
  onLocaleChange,
  themePreference,
  onThemePreferenceChange,
}: AppProps) {
  const apiTargetHint = resolveApiTargetHint();
  const t = (key: Parameters<typeof translate>[1], vars?: Record<string, string>) =>
    translate(locale, key, vars);
  const [activeCluster, setActiveCluster] = useState('default');
  const [listFilterNamespace, setListFilterNamespace] = useState(ALL_NAMESPACES);
  const [lastUsedNamespace, setLastUsedNamespace] = useState<string | undefined>();
  const [createNamespace, setCreateNamespace] = useState('default');
  const [yamlInput, setYamlInput] = useState(
    'apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: example-config\n',
  );
  const [applyStatus, setApplyStatus] = useState<string | null>(null);
  const [applyResults, setApplyResults] = useState<ApplyResultItem[]>([]);
  const [clusters, setClusters] = useState<string[]>(['default']);
  const [menus, setMenus] = useState<MenuItem[]>([]);
  const [resourceTypes, setResourceTypes] = useState<RegistryResourceType[]>([]);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [healthStatus, setHealthStatus] = useState<ProbeStatus>('checking');
  const [readyStatus, setReadyStatus] = useState<ProbeStatus>('checking');
  const [healthError, setHealthError] = useState<string | null>(null);
  const [readyError, setReadyError] = useState<string | null>(null);
  const [lastCheckedAt, setLastCheckedAt] = useState<string | null>(null);
  const [selectedMenuID, setSelectedMenuID] = useState<string>('');
  const [currentRoute, setCurrentRoute] = useState<string>(() => readHashRoute());
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>(
    () => readStoredExpandedGroups(),
  );

  const configuredMenus = useMemo(() => {
    const systemMenus = menus.filter((menu) => menu.source === 'system');
    const userMenus = menus.filter((menu) => menu.source === 'user');
    return composeMenus(systemMenus, userMenus, []);
  }, [menus]);
  const favoritesMenus = useMemo(
    () =>
      configuredMenus.filter(
        (menu) =>
          menu.source === 'user' || menu.group.trim().toUpperCase() === 'FAVORITES',
      ),
    [configuredMenus],
  );
  const groupedMenus = useMemo(() => {
    const favoriteIDs = new Set(favoritesMenus.map((menu) => menu.id));
    return groupMenusByGroup(configuredMenus.filter((menu) => !favoriteIDs.has(menu.id)));
  }, [configuredMenus, favoritesMenus]);
  const resourceCatalogGroups = useMemo(() => {
    const grouped = new Map<string, RegistryResourceType[]>();
    for (const resource of resourceTypes) {
      const groupName = normalizeResourceGroupName(resource.group).toUpperCase();
      if (!grouped.has(groupName)) {
        grouped.set(groupName, []);
      }
      grouped.get(groupName)?.push(resource);
    }
    return Array.from(grouped.entries()).map(([groupName, items]) => ({
      groupName,
      items,
    }));
  }, [resourceTypes]);
  const menuIDs = useMemo(
    () => [...favoritesMenus.map((menu) => menu.id), ...groupedMenus.flatMap((group) => group.items.map((menu) => menu.id))],
    [favoritesMenus, groupedMenus],
  );

  useEffect(() => {
    setCreateNamespace(
      resolveCreateDefaultNamespace({
        listFilterNamespace,
        lastUsedNamespace,
      }),
    );
  }, [listFilterNamespace, lastUsedNamespace]);

  useEffect(() => {
    setExpandedGroups((previous) => {
      const next = { ...previous };
      let changed = false;
      for (const group of groupedMenus) {
        if (next[group.name] === undefined) {
          next[group.name] = true;
          changed = true;
        }
      }
      return changed ? next : previous;
    });
  }, [groupedMenus]);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    window.localStorage.setItem(MENU_GROUPS_STORAGE_KEY, JSON.stringify(expandedGroups));
  }, [expandedGroups]);

  useEffect(() => {
    if (menuIDs.length === 0) {
      if (selectedMenuID !== '') {
        setSelectedMenuID('');
      }
      return;
    }
    if (!menuIDs.includes(selectedMenuID)) {
      setSelectedMenuID(menuIDs[0]);
    }
  }, [menuIDs, selectedMenuID]);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    const onHashChange = () => {
      setCurrentRoute(readHashRoute());
    };
    window.addEventListener('hashchange', onHashChange);
    return () => {
      window.removeEventListener('hashchange', onHashChange);
    };
  }, []);

  useEffect(() => {
    const matched = configuredMenus.find(
      (menu) => normalizeRoute(menu.targetRef) === normalizeRoute(currentRoute),
    );
    if (matched && matched.id !== selectedMenuID) {
      setSelectedMenuID(matched.id);
    }
  }, [configuredMenus, currentRoute, selectedMenuID]);

  function navigateMenu(menu: MenuItem) {
    const nextRoute = normalizeRoute(menu.targetRef);
    setSelectedMenuID(menu.id);
    setCurrentRoute(nextRoute);
    if (typeof window !== 'undefined') {
      window.location.hash = nextRoute;
    }
  }

  async function applyResources() {
    setApplyStatus(null);
    setApplyResults([]);
    try {
      const response = await fetch(
        `/api/resources/apply?cluster=${encodeURIComponent(activeCluster)}&defaultNs=${encodeURIComponent(createNamespace)}`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/yaml' },
          body: yamlInput,
        },
      );
      if (!response.ok) {
        throw new Error(`apply request failed: ${response.status}`);
      }
      const payload = parseApplyResponse(await response.json());
      setApplyStatus(payload.status);
      setApplyResults(payload.results);
      setLastUsedNamespace(createNamespace);
    } catch (e) {
      setApplyStatus(e instanceof Error ? e.message : 'apply failed');
      setApplyResults([]);
    }
  }

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
      setResourceTypes([]);
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
          setResourceTypes(registryPayload.resourceTypes);
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
  const resourceTypeCount = resourceTypes.length;
  const namespacedResourceCount = resourceTypes.filter((item) => item.namespaced).length;
  const clusterScopedResourceCount = resourceTypeCount - namespacedResourceCount;
  const discoveredApiGroupCount = resourceCatalogGroups.length;
  const yamlDocumentCount = countYamlDocuments(yamlInput);
  const runtimeStatus = resolveRuntimeStatus(healthStatus, readyStatus);
  const checkedAtLabel = lastCheckedAt
    ? new Date(lastCheckedAt).toLocaleTimeString()
    : 'never';

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

          <Tooltip
            arrow
            title={`healthz: ${healthStatus}, readyz: ${readyStatus}, checked: ${lastCheckedAt ?? 'never'}`}
          >
            <Chip
              size="small"
              color={statusColor(runtimeStatus)}
              variant="outlined"
              label={`${t('runtime')}: ${runtimeStatus}`}
              sx={{
                borderRadius: 999,
                fontWeight: 600,
                bgcolor: 'background.paper',
              }}
            />
          </Tooltip>
          <Typography variant="caption" color="text.secondary" sx={{ whiteSpace: 'nowrap' }}>
            {t('checked', { time: checkedAtLabel })}
          </Typography>
          <FormControl size="small" sx={{ minWidth: 160 }}>
            <InputLabel htmlFor="cluster-select">{t('cluster')}</InputLabel>
            <Select
              native
              value={activeCluster}
              onChange={(event) => setActiveCluster(event.target.value)}
              label={t('cluster')}
              inputProps={{ id: 'cluster-select' }}
            >
              {clusters.map((clusterId) => (
                <option key={clusterId} value={clusterId}>
                  {clusterId}
                </option>
              ))}
            </Select>
          </FormControl>

          <Box sx={{ flexGrow: 1 }} />

          <FormControl size="small" sx={{ minWidth: 160 }}>
            <InputLabel htmlFor="theme-select">{t('theme')}</InputLabel>
            <Select
              native
              value={themePreference}
              onChange={(event) =>
                onThemePreferenceChange(event.target.value as ThemePreference)
              }
              label={t('theme')}
              inputProps={{ id: 'theme-select' }}
            >
              <option value="system">{t('system')}</option>
              <option value="light">{t('light')}</option>
              <option value="dark">{t('dark')}</option>
            </Select>
          </FormControl>
          <FormControl size="small" sx={{ minWidth: 140 }}>
            <InputLabel htmlFor="language-select">{t('language')}</InputLabel>
            <Select
              native
              value={locale}
              onChange={(event) => onLocaleChange(event.target.value as Locale)}
              label={t('language')}
              inputProps={{ id: 'language-select' }}
            >
              <option value="en">{t('english')}</option>
              <option value="zh">{t('chinese')}</option>
            </Select>
          </FormControl>
          <Button variant="contained" onClick={() => setCreateDialogOpen(true)}>
            {t('createResources')}
          </Button>
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
            {loading ? <Typography>{t('loadingMenus')}</Typography> : null}
            {error ? (
              <Typography color="error">{t('failedMenus', { error })}</Typography>
            ) : null}

            <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
              {t('favorites')}
            </Typography>
            {favoritesMenus.length === 0 ? (
              <Typography color="text.secondary">{t('noMenus')}</Typography>
            ) : (
              <List dense sx={{ pt: 0 }}>
                {favoritesMenus.map((menu) => (
                  <ListItem key={menu.id} disablePadding>
                    <ListItemButton
                      selected={
                        selectedMenuID === menu.id ||
                        normalizeRoute(menu.targetRef) === normalizeRoute(currentRoute)
                      }
                      onClick={() => navigateMenu(menu)}
                      sx={{
                        borderRadius: 1.2,
                        mb: 0.3,
                        transition: 'all 150ms ease',
                        '&.Mui-selected': {
                          bgcolor: 'action.selected',
                          boxShadow: 'inset 0 0 0 1px',
                          borderColor: 'divider',
                        },
                      }}
                    >
                      <ListItemText primary={`${menuIcon(menu.group, menu.targetType)} ${menu.title}`} />
                    </ListItemButton>
                  </ListItem>
                ))}
              </List>
            )}

            <Divider />

            {groupedMenus.length === 0 ? (
              <Typography color="text.secondary">{t('noMenus')}</Typography>
            ) : (
              groupedMenus.map((group) => (
                <Box key={group.name}>
                  <ListItemButton
                    onClick={() =>
                      setExpandedGroups((previous) => ({
                        ...previous,
                        [group.name]: !previous[group.name],
                      }))
                    }
                    sx={{
                      px: 1,
                      py: 0.4,
                      borderRadius: 1.2,
                      mb: 0.2,
                    }}
                  >
                    <ListItemText
                      primary={`${expandedGroups[group.name] ? '▾' : '▸'} ${group.name}`}
                      primaryTypographyProps={{
                        fontWeight: 700,
                        fontSize: 12,
                        color: 'text.secondary',
                        letterSpacing: 0.35,
                      }}
                    />
                  </ListItemButton>
                  <Collapse in={expandedGroups[group.name] ?? true} timeout={160}>
                    <List dense sx={{ pt: 0, pl: 0.8 }}>
                      {group.items.map((item) => (
                        <ListItem key={item.id} disablePadding>
                          <ListItemButton
                            selected={
                              selectedMenuID === item.id ||
                              normalizeRoute(item.targetRef) === normalizeRoute(currentRoute)
                            }
                            onClick={() => navigateMenu(item)}
                            sx={{
                              borderRadius: 1.2,
                              mb: 0.2,
                              '&.Mui-selected': {
                                bgcolor: 'action.selected',
                                boxShadow: 'inset 0 0 0 1px',
                                borderColor: 'divider',
                              },
                            }}
                          >
                            <ListItemText primary={item.title} />
                          </ListItemButton>
                        </ListItem>
                      ))}
                    </List>
                  </Collapse>
                </Box>
              ))
            )}
          </Stack>
        </Paper>

        <Stack spacing={2}>
          <Paper elevation={1} sx={{ p: 1.4, border: 1, borderColor: 'divider', borderRadius: 2 }}>
            <Stack
              direction={{ xs: 'column', sm: 'row' }}
              spacing={1.2}
              alignItems={{ sm: 'center' }}
            >
              <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 700 }}>
                {t('context')}
              </Typography>
              <Chip size="small" variant="outlined" label={`${t('cluster')}: ${activeCluster}`} />
              <FormControl size="small" sx={{ minWidth: 220 }}>
                <InputLabel htmlFor="namespace-filter-select">{t('namespaceFilter')}</InputLabel>
                <Select
                  native
                  value={listFilterNamespace}
                  onChange={(event) => setListFilterNamespace(event.target.value)}
                  label={t('namespaceFilter')}
                  inputProps={{ id: 'namespace-filter-select' }}
                >
                  <option value={ALL_NAMESPACES}>{ALL_NAMESPACES}</option>
                  <option value="default">default</option>
                  <option value="kube-system">kube-system</option>
                  <option value="dev">dev</option>
                </Select>
              </FormControl>
            </Stack>
          </Paper>
          <Paper
            elevation={3}
            sx={{
              p: 2.2,
              border: 1,
              borderColor: 'divider',
              background:
                'linear-gradient(120deg, rgba(25,118,210,0.08), rgba(46,125,50,0.07))',
            }}
          >
            <Typography variant="overline" color="primary.main" sx={{ letterSpacing: 1.1 }}>
              {t('controlPlane')}
            </Typography>
            <Typography variant="h5" sx={{ fontWeight: 700, mb: 0.8 }}>
              {t('clusterOverview', { cluster: activeCluster })}
            </Typography>
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems={{ sm: 'center' }}>
              <Typography color="text.secondary">{t('apiTarget', { target: apiTargetHint })}</Typography>
              <Chip size="small" label={`${t('health')}: ${healthStatus}`} color={statusColor(healthStatus)} />
              <Chip size="small" label={`${t('ready')}: ${readyStatus}`} color={statusColor(readyStatus)} />
            </Stack>
          </Paper>

          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, minmax(0, 1fr))', lg: 'repeat(4, minmax(0, 1fr))' },
              gap: 1.5,
            }}
          >
            <Paper elevation={1} sx={{ p: 1.6, border: 1, borderColor: 'divider', borderRadius: 2 }}>
              <Typography variant="caption" color="text.secondary">
                {t('registryResourceTypes')}
              </Typography>
              <Typography variant="h4" sx={{ fontWeight: 700 }} data-testid="registry-resource-type-count">
                {resourceTypeCount}
              </Typography>
            </Paper>
            <Paper elevation={1} sx={{ p: 1.6, border: 1, borderColor: 'divider', borderRadius: 2 }}>
              <Typography variant="caption" color="text.secondary">
                {t('namespacedTypes')}
              </Typography>
              <Typography variant="h4" sx={{ fontWeight: 700 }} data-testid="namespaced-resource-type-count">
                {namespacedResourceCount}
              </Typography>
            </Paper>
            <Paper elevation={1} sx={{ p: 1.6, border: 1, borderColor: 'divider', borderRadius: 2 }}>
              <Typography variant="caption" color="text.secondary">
                {t('clusterScopedTypes')}
              </Typography>
              <Typography variant="h4" sx={{ fontWeight: 700 }} data-testid="cluster-scoped-resource-type-count">
                {clusterScopedResourceCount}
              </Typography>
            </Paper>
            <Paper elevation={1} sx={{ p: 1.6, border: 1, borderColor: 'divider', borderRadius: 2 }}>
              <Typography variant="caption" color="text.secondary">
                {t('resourceCatalogGroups')}
              </Typography>
              <Typography variant="h4" sx={{ fontWeight: 700 }}>
                {discoveredApiGroupCount}
              </Typography>
            </Paper>
          </Box>

          <Paper elevation={1} sx={{ p: 1.6, border: 1, borderColor: 'divider', borderRadius: 2 }}>
            <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 0.4 }}>
              {t('failureSummary')}
            </Typography>
            <Typography variant="body2" color={failureSummary === 'none' ? 'text.secondary' : 'error'}>
              {t('failureSummaryText', { summary: failureSummary })}
            </Typography>
          </Paper>
        </Stack>
      </Box>

      <Dialog
        open={createDialogOpen}
        onClose={() => setCreateDialogOpen(false)}
        fullWidth
        maxWidth="md"
      >
        <DialogTitle>{t('createResources')}</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={1.2} sx={{ pt: 0.4 }}>
            <FormControl size="small" sx={{ maxWidth: 280 }}>
              <InputLabel htmlFor="create-namespace-select">{t('createNamespace')}</InputLabel>
              <Select
                native
                value={createNamespace}
                onChange={(event) => {
                  setCreateNamespace(event.target.value);
                  setLastUsedNamespace(event.target.value);
                }}
                label={t('createNamespace')}
                inputProps={{ id: 'create-namespace-select' }}
              >
                <option value="default">default</option>
                <option value="kube-system">kube-system</option>
                <option value="dev">dev</option>
              </Select>
            </FormControl>
            <Typography variant="caption" color="text.secondary">
              {t('yamlDocuments', { count: String(yamlDocumentCount) })}
            </Typography>
            <TextField
              label={t('yamlLabel')}
              multiline
              minRows={10}
              value={yamlInput}
              onChange={(event) => setYamlInput(event.target.value)}
            />

            {applyStatus ? (
              <Typography
                variant="body2"
                color={applyStatus === 'success' ? 'success.main' : applyStatus === 'partial' ? 'warning.main' : 'error'}
              >
                  {t('applyStatus', { status: applyStatus })}
                </Typography>
              ) : null}
            {applyResults.length > 0 ? (
              <List dense sx={{ border: 1, borderColor: 'divider', borderRadius: 1 }}>
                {applyResults.map((result) => (
                  <ListItem key={`${result.index}-${result.name}-${result.kind}`} disableGutters sx={{ px: 1 }}>
                    <ListItemText
                      primary={`#${result.index} ${result.kind || t('unknown')} ${result.name || '-'} (${result.namespace || '-'})`}
                      secondary={result.status === 'failed' ? `${t('failed')}: ${result.reason ?? t('unknown')}` : t('succeeded')}
                    />
                  </ListItem>
                ))}
              </List>
            ) : null}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateDialogOpen(false)}>{t('close')}</Button>
          <Button variant="contained" onClick={() => void applyResources()}>
            {t('applyYaml')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}

export default App;
