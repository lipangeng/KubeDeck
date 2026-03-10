import { useEffect, useMemo, useReducer, useState } from 'react';
import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import { PrimarySidebar } from './components/app-shell/PrimarySidebar';
import { composeMenus } from './core/menuComposer';
import { ActionDrawer } from './features/actions/ActionDrawer';
import { HomepageView } from './pages/homepage/HomepageView';
import { WorkloadsPage } from './pages/workloads/WorkloadsPage';
import {
  parseClustersResponse,
  parseMenusResponse,
  parseWorkloadsResponse,
} from './sdk/metaApi';
import type { MenuItem, WorkloadItem } from './sdk/types';
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

function App({ themePreference, onThemePreferenceChange }: AppProps) {
  const apiTargetHint = resolveApiTargetHint();
  const [workContext, dispatchWorkContext] = useReducer(
    applyWorkContextEvent,
    undefined,
    () => createSharedWorkingContext('default'),
  );
  const [clusters, setClusters] = useState<string[]>(['default']);
  const [menus, setMenus] = useState<MenuItem[]>([]);
  const [workloadItems, setWorkloadItems] = useState<WorkloadItem[]>([]);
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
  const requestedNamespaceScope =
    namespaceScope.mode === 'all' ? 'all' : resolveSingleNamespace(namespaceScope);
  const namespaceScopeLabel = describeNamespaceScope(namespaceScope);

  const visibleWorkloads = useMemo(
    () => workloadItems.filter((workload) => {
      const searchText = workloadsContext.listContext.searchText?.trim().toLowerCase();
      if (!searchText) {
        return true;
      }

      return (
        workload.name.toLowerCase().includes(searchText) ||
        workload.kind.toLowerCase().includes(searchText) ||
        workload.namespace.toLowerCase().includes(searchText)
      );
    }),
    [workloadItems, workloadsContext.listContext.searchText],
  );

  const blockingSummary = useMemo(() => {
    const parts = [
      metadataError ? `metadata: ${metadataError}` : null,
      healthError ? `healthz: ${healthError}` : null,
      readyError ? `readyz: ${readyError}` : null,
    ].filter(Boolean);

    return parts.length > 0 ? parts.join('; ') : null;
  }, [healthError, metadataError, readyError]);

  const workloadsResultBanner = actionResultContext.actionContext.resultSummary
    ? actionResultContext.actionContext.resultSummary.outcome === 'failure'
      ? `${formatActionLabel(actionResultContext.actionContext.actionType)} failed: ${actionResultContext.actionContext.resultSummary.failedObjects?.join(', ')}`
      : `${formatActionLabel(actionResultContext.actionContext.actionType)} accepted: ${actionResultContext.actionContext.resultSummary.affectedObjects?.join(', ')}`
    : null;

  const workloadsResultSeverity =
    actionResultContext.actionContext.resultSummary?.outcome === 'failure'
      ? 'error'
      : 'success';

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
        const [menusResponse, workloadsResponse] = await Promise.all([
          fetch(`/api/meta/menus?cluster=${encodeURIComponent(metadataClusterId)}`),
          fetch(
            `/api/resources/workloads?cluster=${encodeURIComponent(metadataClusterId)}&namespace=${encodeURIComponent(requestedNamespaceScope)}`,
          ),
        ]);
        if (!menusResponse.ok) {
          throw new Error(`menus request failed: ${menusResponse.status}`);
        }
        if (!workloadsResponse.ok) {
          throw new Error(`workloads request failed: ${workloadsResponse.status}`);
        }

        const menusPayload = parseMenusResponse(await menusResponse.json());
        const workloadPayload = parseWorkloadsResponse(await workloadsResponse.json());
        if (!active) {
          return;
        }

        setMenus(menusPayload.menus);
        setWorkloadItems(workloadPayload.items);

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
        setWorkloadItems([]);
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
  }, [metadataClusterId, reloadToken, requestedClusterId, requestedNamespaceScope]);

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
        <PrimarySidebar
          clusters={clusters}
          selectedCluster={visibleClusterValue}
          namespaceScopeLabel={namespaceScopeLabel}
          clusterStatus={clusterContext.status}
          isWorkloadsPage={isWorkloadsPage}
          secondaryEntries={secondaryEntries}
          onClusterChange={handleClusterChange}
          onEnterWorkloads={handleEnterWorkloads}
        />

        {isWorkloadsPage ? (
          <WorkloadsPage
            activeClusterId={workloadsContext.activeCluster.id}
            namespaceScope={workloadsContext.namespaceScope}
            namespaceScopeLabel={namespaceScopeLabel}
            searchText={workloadsContext.listContext.searchText ?? ''}
            onSearchTextChange={(next) =>
              dispatchWorkContext({
                type: 'update_list_context',
                next: { searchText: next },
              })
            }
            onReturnHomepage={handleReturnHomepage}
            onRefresh={() => setReloadToken((current) => current + 1)}
            onNamespaceScopeChange={handleNamespaceScopeChange}
            onOpenAction={handleOpenAction}
            loading={loadingMetadata}
            metadataError={metadataError}
            workloads={visibleWorkloads}
            resultBanner={workloadsResultBanner}
            resultBannerSeverity={workloadsResultSeverity}
            onDismissResult={() =>
              dispatchWorkContext({ type: 'acknowledge_action_result' })
            }
          />
        ) : (
          <HomepageView
            activeClusterId={homepageContext.activeCluster.id}
            namespaceScopeLabel={namespaceScopeLabel}
            primaryEntryTitle={primaryWorkloadsEntry?.title ?? 'Workloads'}
            onEnterWorkloads={handleEnterWorkloads}
            blockingSummary={blockingSummary}
            healthStatus={healthStatus}
            readyStatus={readyStatus}
            apiTargetHint={apiTargetHint}
            lastCheckedAt={lastCheckedAt}
            additionalEntries={secondaryEntries}
            statusColor={statusColor}
          />
        )}
      </Box>

      <ActionDrawer
        open={actionDrawerOpen}
        actionLabel={actionLabel}
        activeClusterId={createApplyContext.activeCluster.id}
        namespaceScopeLabel={namespaceScopeLabel}
        namespaceScope={createApplyContext.namespaceScope}
        actionContext={createApplyContext.actionContext}
        actionNamespace={actionNamespace}
        actionManifest={actionManifest}
        actionFormError={actionFormError}
        onClose={handleCloseAction}
        onSubmit={handleSubmitAction}
        onBackToEdit={handleBackToEdit}
        onReturnToWorkloads={handleReturnToWorkloads}
        onNamespaceChange={setActionNamespace}
        onManifestChange={setActionManifest}
        resultSeverity={resolveResultSeverity}
      />
    </Box>
  );
}

export default App;
