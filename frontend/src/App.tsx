import { useEffect, useMemo, useReducer, useState } from 'react';
import Alert from '@mui/material/Alert';
import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardActionArea from '@mui/material/CardActionArea';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import Drawer from '@mui/material/Drawer';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import Paper from '@mui/material/Paper';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TextField from '@mui/material/TextField';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import { ListPageShell } from './components/page-shell/ResourcePageShell';
import { composeMenus } from './core/menuComposer';
import { parseClustersResponse, parseMenusResponse, parseRegistryResponse } from './sdk/metaApi';
import type { MenuItem, RegistryResourceType } from './sdk/types';
import { type ThemePreference } from './themeMode';
import {
  applyWorkContextEvent,
  createSharedWorkingContext,
} from './state/work-context/events';
import {
  selectActionResultContext,
  selectClusterContext,
  selectCreateApplyContext,
  selectHomepageContextSummary,
  selectNamespaceScope,
  selectWorkloadsContext,
} from './state/work-context/selectors';
import type { ActionType, NamespaceScope } from './state/work-context/types';

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

function describeNamespaceScope(scope: NamespaceScope): string {
  if (scope.mode === 'all') {
    return 'All namespaces';
  }

  if (scope.mode === 'multiple') {
    return scope.values.join(', ');
  }

  return scope.values[0] ?? 'default';
}

function resolveSingleNamespace(scope: NamespaceScope): string {
  return scope.mode === 'single' ? (scope.values[0] ?? 'default') : '';
}

function resolveWorkloadKinds(resourceTypes: RegistryResourceType[]): RegistryResourceType[] {
  return resourceTypes.filter((resourceType) =>
    ['Deployment', 'Service'].includes(resourceType.kind),
  );
}

function resolveResultSeverity(outcome: string): 'success' | 'warning' | 'error' {
  if (outcome === 'success') {
    return 'success';
  }
  if (outcome === 'partial_failure') {
    return 'warning';
  }
  return 'error';
}

function formatActionLabel(actionType: ActionType | undefined): string {
  return actionType === 'create' ? 'Create' : 'Apply';
}

function PrimaryEntryCard({
  title,
  description,
  actionLabel,
  onClick,
}: {
  title: string;
  description: string;
  actionLabel: string;
  onClick: () => void;
}) {
  return (
    <Card variant="outlined">
      <CardActionArea onClick={onClick}>
        <CardContent>
          <Stack spacing={1}>
            <Typography variant="overline" color="primary.main">
              Primary Workflow
            </Typography>
            <Typography variant="h5" sx={{ fontWeight: 700 }}>
              {title}
            </Typography>
            <Typography color="text.secondary">{description}</Typography>
            <Typography variant="button" color="primary.main">
              {actionLabel}
            </Typography>
          </Stack>
        </CardContent>
      </CardActionArea>
    </Card>
  );
}

function AdditionalEntryList({
  items,
}: {
  items: MenuItem[];
}) {
  return (
    <Paper variant="outlined" sx={{ p: 2 }}>
      <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 1 }}>
        Additional Entries
      </Typography>
      <List dense sx={{ pt: 0 }}>
        {items.length === 0 ? (
          <ListItemText
            primary="No additional entries yet"
            primaryTypographyProps={{ color: 'text.disabled' }}
          />
        ) : (
          items.map((menu) => (
            <ListItemButton key={menu.id} disabled>
              <ListItemText
                primary={menu.title}
                secondary={`${menu.targetType} · available later`}
              />
            </ListItemButton>
          ))
        )}
      </List>
    </Paper>
  );
}

function App({ themePreference, onThemePreferenceChange }: AppProps) {
  const apiTargetHint = resolveApiTargetHint();
  const [workContext, dispatchWorkContext] = useReducer(
    applyWorkContextEvent,
    undefined,
    () => createSharedWorkingContext('default'),
  );
  const [clusters, setClusters] = useState<string[]>(['default']);
  const [menus, setMenus] = useState<MenuItem[]>([]);
  const [resourceTypes, setResourceTypes] = useState<RegistryResourceType[]>([]);
  const [loadingMetadata, setLoadingMetadata] = useState(true);
  const [metadataError, setMetadataError] = useState<string | null>(null);
  const [healthStatus, setHealthStatus] = useState<ProbeStatus>('checking');
  const [readyStatus, setReadyStatus] = useState<ProbeStatus>('checking');
  const [healthError, setHealthError] = useState<string | null>(null);
  const [readyError, setReadyError] = useState<string | null>(null);
  const [lastCheckedAt, setLastCheckedAt] = useState<string | null>(null);
  const [requestedClusterId, setRequestedClusterId] = useState<string | null>(null);
  const [reloadToken, setReloadToken] = useState(0);
  const [actionDrawerOpen, setActionDrawerOpen] = useState(false);
  const [actionManifest, setActionManifest] = useState(
    [
      'apiVersion: apps/v1',
      'kind: Deployment',
      'metadata:',
      '  name: sample-api',
      'spec:',
      '  replicas: 1',
      '',
    ].join('\n'),
  );
  const [actionNamespace, setActionNamespace] = useState('default');
  const [actionFormError, setActionFormError] = useState<string | null>(null);

  const clusterContext = selectClusterContext(workContext);
  const namespaceScope = selectNamespaceScope(workContext);
  const homepageContext = selectHomepageContextSummary(workContext);
  const workloadsContext = selectWorkloadsContext(workContext);
  const createApplyContext = selectCreateApplyContext(workContext);
  const actionResultContext = selectActionResultContext(workContext);
  const actionType = createApplyContext.actionContext.actionType ?? 'apply';
  const actionLabel = formatActionLabel(actionType);
  const currentClusterId = clusterContext.id;
  const metadataClusterId = requestedClusterId ?? currentClusterId;
  const visibleClusterValue = requestedClusterId ?? currentClusterId;
  const isWorkloadsPage = workloadsContext.workflowDomain.id === 'workloads';

  const composedMenus = useMemo(() => {
    const systemMenus = menus.filter((menu) => menu.source === 'system');
    const userMenus = menus.filter((menu) => menu.source === 'user');
    const dynamicMenus = menus.filter((menu) => menu.source === 'dynamic');
    return composeMenus(systemMenus, userMenus, dynamicMenus);
  }, [menus]);

  const primaryWorkloadsEntry = useMemo(
    () => composedMenus.find((menu) => menu.targetRef === '/workloads') ?? null,
    [composedMenus],
  );
  const secondaryEntries = useMemo(
    () => composedMenus.filter((menu) => menu.targetRef !== '/workloads'),
    [composedMenus],
  );
  const workloadKinds = useMemo(
    () => resolveWorkloadKinds(resourceTypes).filter((resourceType) => {
      const searchText = workloadsContext.listContext.searchText?.trim().toLowerCase();
      if (!searchText) {
        return true;
      }

      return (
        resourceType.kind.toLowerCase().includes(searchText) ||
        resourceType.plural.toLowerCase().includes(searchText)
      );
    }),
    [resourceTypes, workloadsContext.listContext.searchText],
  );

  const blockingSummary = useMemo(() => {
    const parts = [
      metadataError ? `metadata: ${metadataError}` : null,
      healthError ? `healthz: ${healthError}` : null,
      readyError ? `readyz: ${readyError}` : null,
    ].filter(Boolean);

    return parts.length > 0 ? parts.join('; ') : null;
  }, [healthError, metadataError, readyError]);

  useEffect(() => {
    let active = true;

    async function loadClusters() {
      try {
        const response = await fetch('/api/meta/clusters');
        if (!response.ok) {
          throw new Error(`clusters request failed: ${response.status}`);
        }
        const payload = parseClustersResponse(await response.json());
        if (!active || payload.clusters.length === 0) {
          return;
        }

        setClusters(payload.clusters);
        if (!payload.clusters.includes(currentClusterId)) {
          dispatchWorkContext({ type: 'request_cluster_switch' });
          setRequestedClusterId(payload.clusters[0]);
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
  }, [currentClusterId]);

  useEffect(() => {
    let active = true;

    async function loadClusterMetadata() {
      setLoadingMetadata(true);
      setMetadataError(null);
      try {
        const [menusResponse, registryResponse] = await Promise.all([
          fetch(`/api/meta/menus?cluster=${encodeURIComponent(metadataClusterId)}`),
          fetch(`/api/meta/registry?cluster=${encodeURIComponent(metadataClusterId)}`),
        ]);
        if (!menusResponse.ok) {
          throw new Error(`menus request failed: ${menusResponse.status}`);
        }
        if (!registryResponse.ok) {
          throw new Error(`registry request failed: ${registryResponse.status}`);
        }

        const menusPayload = parseMenusResponse(await menusResponse.json());
        const registryPayload = parseRegistryResponse(await registryResponse.json());
        if (!active) {
          return;
        }

        setMenus(menusPayload.menus);
        setResourceTypes(registryPayload.resourceTypes);

        if (requestedClusterId === metadataClusterId) {
          dispatchWorkContext({
            type: 'complete_cluster_switch',
            clusterId: metadataClusterId,
          });
          setRequestedClusterId(null);
        }
      } catch (error) {
        if (!active) {
          return;
        }

        setMetadataError(error instanceof Error ? error.message : 'unknown error');
        setMenus([]);
        setResourceTypes([]);
        if (requestedClusterId === metadataClusterId) {
          dispatchWorkContext({ type: 'fail_cluster_switch' });
          setRequestedClusterId(null);
        }
      } finally {
        if (active) {
          setLoadingMetadata(false);
        }
      }
    }

    loadClusterMetadata();

    return () => {
      active = false;
    };
  }, [metadataClusterId, reloadToken, requestedClusterId]);

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
      } catch (error) {
        if (error instanceof Error) {
          return { status: 'error', error: error.message };
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

  useEffect(() => {
    if (!actionDrawerOpen) {
      return;
    }

    if (namespaceScope.mode === 'single') {
      setActionNamespace(resolveSingleNamespace(namespaceScope));
    }
  }, [actionDrawerOpen, namespaceScope]);

  function handleClusterChange(nextClusterId: string) {
    if (nextClusterId === visibleClusterValue) {
      return;
    }

    dispatchWorkContext({ type: 'request_cluster_switch' });
    setRequestedClusterId(nextClusterId);
  }

  function handleEnterWorkloads() {
    dispatchWorkContext({ type: 'enter_workloads' });
  }

  function handleReturnHomepage() {
    dispatchWorkContext({ type: 'enter_homepage' });
  }

  function handleNamespaceScopeChange(value: string) {
    if (value === 'all') {
      dispatchWorkContext({
        type: 'update_namespace_scope',
        scope: { mode: 'all', values: [], source: 'user_selected' },
      });
      return;
    }

    dispatchWorkContext({
      type: 'update_namespace_scope',
      scope: { mode: 'single', values: [value], source: 'user_selected' },
    });
  }

  function handleOpenAction(nextActionType: ActionType) {
    dispatchWorkContext({ type: 'start_action', actionType: nextActionType });
    if (namespaceScope.mode === 'single') {
      setActionNamespace(resolveSingleNamespace(namespaceScope));
    } else {
      setActionNamespace('');
    }
    setActionFormError(null);
    setActionDrawerOpen(true);
  }

  function handleCloseAction() {
    setActionDrawerOpen(false);
    setActionFormError(null);
    dispatchWorkContext({ type: 'acknowledge_action_result' });
  }

  function handleReturnToWorkloads() {
    setActionDrawerOpen(false);
    setActionFormError(null);
    dispatchWorkContext({ type: 'return_to_workloads' });
    setReloadToken((current) => current + 1);
  }

  function handleBackToEdit() {
    setActionFormError(null);
    dispatchWorkContext({ type: 'start_action', actionType });
  }

  async function handleSubmitAction() {
    dispatchWorkContext({ type: 'validate_action' });

    const manifest = actionManifest.trim();
    if (manifest === '') {
      setActionFormError('Manifest is required before submit.');
      dispatchWorkContext({ type: 'fail_action_validation' });
      return;
    }

    const namespaceTarget =
      namespaceScope.mode === 'single'
        ? resolveSingleNamespace(namespaceScope)
        : actionNamespace.trim();
    if (namespaceTarget === '') {
      setActionFormError('Namespace target is required when browsing all namespaces.');
      dispatchWorkContext({ type: 'fail_action_validation' });
      return;
    }

    setActionFormError(null);
    dispatchWorkContext({
      type: 'resolve_execution_target',
      target: { kind: 'namespace', namespace: namespaceTarget },
    });

    try {
      dispatchWorkContext({ type: 'submit_action' });
      const response = await fetch('/api/resources/apply', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          actionType,
          cluster: currentClusterId,
          namespace: namespaceTarget,
          manifest,
        }),
      });
      if (!response.ok) {
        throw new Error(`${actionLabel.toLowerCase()} request failed: ${response.status}`);
      }

      dispatchWorkContext({
        type: 'complete_action_success',
        summary: {
          affectedObjects: [`${actionLabel.toLowerCase()} accepted for ${namespaceTarget}`],
        },
      });
    } catch (error) {
      dispatchWorkContext({
        type: 'complete_action_failure',
        summary: {
          failedObjects: [
            error instanceof Error
              ? error.message
              : `${actionLabel.toLowerCase()} request failed`,
          ],
        },
      });
    }
  }

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
          gridTemplateColumns: { xs: '1fr', md: '300px minmax(0, 1fr)' },
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
              Working Context
            </Typography>

            <FormControl size="small" fullWidth>
              <InputLabel htmlFor="cluster-select">Cluster</InputLabel>
              <Select
                native
                value={visibleClusterValue}
                onChange={(event) => handleClusterChange(event.target.value)}
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

            <Paper variant="outlined" sx={{ p: 1.5 }}>
              <Typography variant="caption" color="text.secondary">
                Active namespace scope
              </Typography>
              <Typography sx={{ fontWeight: 700 }}>
                {describeNamespaceScope(homepageContext.namespaceScope)}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Cluster status: {clusterContext.status}
              </Typography>
            </Paper>

            <Divider />

            <Button
              variant={isWorkloadsPage ? 'contained' : 'outlined'}
              onClick={handleEnterWorkloads}
            >
              Enter Workloads
            </Button>

            <AdditionalEntryList items={secondaryEntries} />
          </Stack>
        </Paper>

        <Stack spacing={2}>
          {!isWorkloadsPage ? (
            <>
              <Paper elevation={3} sx={{ p: 2.2, border: 1, borderColor: 'divider' }}>
                <Typography variant="overline" color="primary.main" sx={{ letterSpacing: 1.1 }}>
                  Current Context
                </Typography>
                <Typography variant="h5" sx={{ fontWeight: 700, mb: 0.8 }}>
                  Cluster {homepageContext.activeCluster.id}
                </Typography>
                <Typography color="text.secondary">
                  Namespace scope: {describeNamespaceScope(homepageContext.namespaceScope)}
                </Typography>
              </Paper>

              <PrimaryEntryCard
                title={primaryWorkloadsEntry?.title ?? 'Workloads'}
                description="Browse workload-capable resources and run your first apply flow in the current cluster context."
                actionLabel="Enter Workloads"
                onClick={handleEnterWorkloads}
              />

              <Paper variant="outlined" sx={{ p: 2 }}>
                <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 1 }}>
                  Default Task
                </Typography>
                <Typography color="text.secondary" sx={{ mb: 1.5 }}>
                  Start in Workloads to inspect supported workload types and continue with apply in the same cluster and namespace context.
                </Typography>
                <Button variant="text" onClick={handleEnterWorkloads}>
                  Continue to Workloads
                </Button>
              </Paper>

              {blockingSummary ? (
                <Alert severity="warning">
                  Blocking summary: {blockingSummary}
                </Alert>
              ) : null}

              <Paper variant="outlined" sx={{ p: 2 }}>
                <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 1 }}>
                  Runtime diagnostics
                </Typography>
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ mb: 1 }}>
                  <Chip
                    size="small"
                    color={statusColor(healthStatus)}
                    label={`healthz: ${healthStatus}`}
                  />
                  <Chip
                    size="small"
                    color={statusColor(readyStatus)}
                    label={`readyz: ${readyStatus}`}
                  />
                </Stack>
                <Typography variant="body2" color="text.secondary">
                  API target ({apiTargetHint}) · Last checked: {lastCheckedAt ?? 'never'}
                </Typography>
              </Paper>
            </>
          ) : (
            <ListPageShell
              title="Workloads"
              toolbar={
                <Stack direction="row" spacing={1}>
                  <Button variant="text" onClick={handleReturnHomepage}>
                    Homepage
                  </Button>
                  <Button variant="outlined" onClick={() => setReloadToken((current) => current + 1)}>
                    Refresh
                  </Button>
                  <Button variant="contained" onClick={() => handleOpenAction('apply')}>
                    Apply
                  </Button>
                  <Button variant="outlined" onClick={() => handleOpenAction('create')}>
                    Create
                  </Button>
                </Stack>
              }
            >
              <Stack spacing={2}>
                <Paper variant="outlined" sx={{ p: 1.5 }}>
                  <Stack
                    direction={{ xs: 'column', md: 'row' }}
                    spacing={1.5}
                    alignItems={{ md: 'center' }}
                    justifyContent="space-between"
                  >
                    <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                      <Chip
                        color="primary"
                        label={`Cluster: ${workloadsContext.activeCluster.id}`}
                      />
                      <Chip
                        variant="outlined"
                        label={`Namespace scope: ${describeNamespaceScope(workloadsContext.namespaceScope)}`}
                      />
                    </Stack>
                    <FormControl size="small" sx={{ minWidth: 220 }}>
                      <InputLabel htmlFor="namespace-scope-select">Namespace Scope</InputLabel>
                      <Select
                        native
                        value={
                          workloadsContext.namespaceScope.mode === 'all'
                            ? 'all'
                            : resolveSingleNamespace(workloadsContext.namespaceScope)
                        }
                        onChange={(event) => handleNamespaceScopeChange(event.target.value)}
                        label="Namespace Scope"
                        inputProps={{ id: 'namespace-scope-select' }}
                      >
                        <option value="default">default</option>
                        <option value="all">All namespaces</option>
                      </Select>
                    </FormControl>
                  </Stack>
                </Paper>

                <Paper variant="outlined" sx={{ p: 1.5 }}>
                  <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
                    <TextField
                      label="Search workload types"
                      size="small"
                      fullWidth
                      value={workloadsContext.listContext.searchText ?? ''}
                      onChange={(event) =>
                        dispatchWorkContext({
                          type: 'update_list_context',
                          next: { searchText: event.target.value },
                        })
                      }
                    />
                    <Chip
                      variant="outlined"
                      label={`Available types: ${workloadKinds.length}`}
                      data-testid="workload-type-count"
                    />
                  </Stack>
                </Paper>

                {loadingMetadata ? <Alert severity="info">Loading workloads...</Alert> : null}
                {metadataError ? (
                  <Alert severity="error">Failed to load workloads: {metadataError}</Alert>
                ) : null}

                {actionResultContext.actionContext.resultSummary ? (
                  <Alert
                    severity={
                      actionResultContext.actionContext.resultSummary.outcome === 'failure'
                        ? 'error'
                        : 'success'
                    }
                    action={
                      <Button
                        color="inherit"
                        size="small"
                        onClick={() =>
                          dispatchWorkContext({ type: 'acknowledge_action_result' })
                        }
                      >
                        Dismiss
                      </Button>
                    }
                  >
                    {actionResultContext.actionContext.resultSummary.outcome === 'failure'
                      ? `${formatActionLabel(actionResultContext.actionContext.actionType)} failed: ${actionResultContext.actionContext.resultSummary.failedObjects?.join(', ')}`
                      : `${formatActionLabel(actionResultContext.actionContext.actionType)} accepted: ${actionResultContext.actionContext.resultSummary.affectedObjects?.join(', ')}`}
                  </Alert>
                ) : null}

                <Paper variant="outlined">
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell>Kind</TableCell>
                        <TableCell>Plural</TableCell>
                        <TableCell>Scope</TableCell>
                        <TableCell>Source</TableCell>
                        <TableCell>Version</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {workloadKinds.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={5}>
                            <Typography color="text.secondary">
                              No workload-capable resource types available in the current cluster context.
                            </Typography>
                          </TableCell>
                        </TableRow>
                      ) : (
                        workloadKinds.map((resourceType) => (
                          <TableRow key={resourceType.id}>
                            <TableCell sx={{ fontWeight: 700 }}>{resourceType.kind}</TableCell>
                            <TableCell>{resourceType.plural}</TableCell>
                            <TableCell>
                              {resourceType.namespaced ? 'Namespaced' : 'Cluster-scoped'}
                            </TableCell>
                            <TableCell>{resourceType.source}</TableCell>
                            <TableCell>{resourceType.preferredVersion}</TableCell>
                          </TableRow>
                        ))
                      )}
                    </TableBody>
                  </Table>
                </Paper>
              </Stack>
            </ListPageShell>
          )}
        </Stack>
      </Box>

      <Drawer
        anchor="right"
        open={actionDrawerOpen}
        onClose={handleCloseAction}
        PaperProps={{ sx: { width: { xs: '100%', sm: 480 }, p: 2 } }}
      >
        <Stack spacing={2}>
          <Typography variant="overline" color="primary.main">
            {actionLabel} Workflow
          </Typography>
          <Typography variant="h5" sx={{ fontWeight: 700 }}>
            {actionLabel} in {createApplyContext.activeCluster.id}
          </Typography>
          <Typography color="text.secondary">
            Browsing scope: {describeNamespaceScope(createApplyContext.namespaceScope)}
          </Typography>

          {createApplyContext.actionContext.resultSummary ? (
            <>
              <Alert
                severity={resolveResultSeverity(
                  createApplyContext.actionContext.resultSummary.outcome,
                )}
              >
                {createApplyContext.actionContext.resultSummary.outcome === 'success'
                  ? `${actionLabel} succeeded`
                  : createApplyContext.actionContext.resultSummary.outcome ===
                      'partial_failure'
                    ? `${actionLabel} partially failed`
                    : `${actionLabel} failed`}
              </Alert>

              <Paper variant="outlined" sx={{ p: 1.5 }}>
                <Stack spacing={1}>
                  <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                    Result summary
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Execution target:{' '}
                    {createApplyContext.actionContext.executionTarget?.kind === 'namespace'
                      ? createApplyContext.actionContext.executionTarget.namespace
                      : 'cluster-scoped'}
                  </Typography>
                  {createApplyContext.actionContext.resultSummary.affectedObjects?.length ? (
                    <Typography variant="body2">
                      Affected: {createApplyContext.actionContext.resultSummary.affectedObjects.join(', ')}
                    </Typography>
                  ) : null}
                  {createApplyContext.actionContext.resultSummary.failedObjects?.length ? (
                    <Typography variant="body2" color="error">
                      Failed: {createApplyContext.actionContext.resultSummary.failedObjects.join(', ')}
                    </Typography>
                  ) : null}
                </Stack>
              </Paper>

              <Stack direction="row" spacing={1} justifyContent="flex-end">
                {createApplyContext.actionContext.resultSummary.outcome === 'failure' ? (
                  <Button variant="outlined" onClick={handleBackToEdit}>
                    Back to Edit
                  </Button>
                ) : null}
                <Button variant="contained" onClick={handleReturnToWorkloads}>
                  Back to Workloads
                </Button>
              </Stack>
            </>
          ) : (
            <>
              {createApplyContext.actionContext.needsRevalidation ? (
                <Alert severity="warning">
                  Namespace browsing scope changed. Review the execution target before submit.
                </Alert>
              ) : null}

              {actionFormError ? <Alert severity="error">{actionFormError}</Alert> : null}

              <TextField
                label="Execution namespace"
                size="small"
                value={actionNamespace}
                onChange={(event) => setActionNamespace(event.target.value)}
                disabled={createApplyContext.namespaceScope.mode === 'single'}
                helperText={
                  createApplyContext.namespaceScope.mode === 'single'
                    ? 'Derived from the current single-namespace browsing scope.'
                    : 'Required because all namespaces is not a valid write target.'
                }
              />

              <TextField
                label="Manifest"
                multiline
                minRows={12}
                value={actionManifest}
                onChange={(event) => setActionManifest(event.target.value)}
              />

              <Stack direction="row" spacing={1} justifyContent="flex-end">
                <Button variant="text" onClick={handleCloseAction}>
                  Cancel
                </Button>
                <Button
                  variant="contained"
                  onClick={handleSubmitAction}
                  disabled={createApplyContext.actionContext.status === 'submitting'}
                >
                  {createApplyContext.actionContext.status === 'submitting'
                    ? 'Submitting...'
                    : `Submit ${actionLabel}`}
                </Button>
              </Stack>
            </>
          )}
        </Stack>
      </Drawer>
    </Box>
  );
}

export default App;
