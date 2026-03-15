import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
  useState,
} from 'react';
import type { FrontendCapabilityModule } from '../sdk';
import { composeKernelNavigation } from './composeKernelNavigation';
import {
  createInitialWorkingContextState,
  reduceWorkingContext,
} from './context/reducer';
import {
  selectActiveCluster,
  selectCurrentResource,
  selectCurrentWorkflowDomain,
  selectNamespaceScope,
} from './context/selectors';
import type { NamespaceScope, ResourceIdentity } from './context/types';
import type { KernelNavigationGroup } from './menu/types';
import { createLocalKernelSnapshot } from './createLocalKernelSnapshot';
import type { ResourcePageExtension } from '../resource-pages/types';
import {
  executeKernelAction as executeKernelActionRequest,
  type KernelActionExecutionRequest,
  type KernelActionExecutionResult,
} from './executeKernelAction';
import { fetchKernelMetadata } from './fetchKernelMetadata';
import { fetchWorkloads, type WorkloadItem } from './fetchWorkloads';
import { hydrateKernelSnapshot } from './hydrateKernelSnapshot';
import { resolveWorkflowActions } from './resolveWorkflowActions';
import { resolveWorkflowSlots } from './resolveWorkflowSlots';
import type { KernelRegistrySnapshot } from './types';

type KernelSource = 'loading' | 'backend' | 'local-fallback';

interface KernelRuntimeContextValue {
  activeRoute: string;
  activePage: KernelRegistrySnapshot['pages'][number] | null;
  activeActions: KernelRegistrySnapshot['actions'];
  activeSummarySlots: KernelRegistrySnapshot['slots'];
  actionSummary: string | null;
  activeCluster: string;
  namespaceScope: NamespaceScope;
  currentWorkflowDomainId: string | null;
  currentResource: ResourceIdentity | null;
  kernelSource: KernelSource;
  navigation: KernelNavigationGroup[];
  registrySnapshot: KernelRegistrySnapshot;
  resourcePageExtensions: ResourcePageExtension[];
  navigate: (route: string) => void;
  reloadKernelMetadata: () => void;
  switchCluster: (cluster: string) => void;
  enterResource: (resource: ResourceIdentity) => void;
  exitResource: () => void;
  fetchWorkloadsForDomain: (workflowDomainId: string, cluster?: string) => Promise<WorkloadItem[]>;
  executeAction: (request: KernelActionExecutionRequest) => Promise<KernelActionExecutionResult>;
}

const KernelRuntimeContext = createContext<KernelRuntimeContextValue | null>(null);

interface KernelRuntimeProviderProps extends PropsWithChildren {
  pluginModules?: FrontendCapabilityModule[];
}

export function KernelRuntimeProvider({
  children,
  pluginModules = [],
}: KernelRuntimeProviderProps) {
  const [activeRoute, setActiveRoute] = useState('/');
  const [runtimeSnapshot, setRuntimeSnapshot] = useState<KernelRegistrySnapshot | null>(null);
  const [kernelSource, setKernelSource] = useState<KernelSource>('loading');
  const [actionSummary, setActionSummary] = useState<string | null>(null);
  const [reloadGeneration, setReloadGeneration] = useState(0);
  const [workingContext, dispatchWorkingContext] = useReducer(
    reduceWorkingContext,
    undefined,
    createInitialWorkingContextState,
  );

  const localSnapshot = useMemo(() => createLocalKernelSnapshot(pluginModules), [pluginModules]);
  const activeCluster = selectActiveCluster(workingContext);

  useEffect(() => {
    let active = true;

    async function loadKernelMetadata() {
      try {
        const remoteMetadata = await fetchKernelMetadata(activeCluster);
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
  }, [activeCluster, localSnapshot, reloadGeneration]);

  const registrySnapshot = runtimeSnapshot ?? localSnapshot;
  const navigation = useMemo(
    () =>
      registrySnapshot.menuGroups.length > 0
        ? registrySnapshot.menuGroups
        : composeKernelNavigation(registrySnapshot.menus),
    [registrySnapshot.menuGroups, registrySnapshot.menus],
  );
  const activePage =
    registrySnapshot.pages.find((page) => page.route === activeRoute) ??
    registrySnapshot.pages[0] ??
    null;
  const activeActions = activePage
    ? resolveWorkflowActions(activePage.workflowDomainId, registrySnapshot.actions)
    : [];
  const activeSummarySlots = activePage
    ? resolveWorkflowSlots(activePage.workflowDomainId, registrySnapshot.slots, 'summary')
    : [];

  const navigate = useCallback(
    (route: string) => {
      setActiveRoute(route);
      const nextPage = registrySnapshot.pages.find((page) => page.route === route);
      if (nextPage) {
        dispatchWorkingContext({
          type: 'enter_workflow_domain',
          workflowDomainId: nextPage.workflowDomainId,
          route,
        });
      }
    },
    [registrySnapshot.pages],
  );

  const fetchWorkloadsForDomain = useCallback(
    async (workflowDomainId: string, cluster = activeCluster) => {
      return fetchWorkloads(workflowDomainId, cluster);
    },
    [activeCluster],
  );

  const enterResource = useCallback((resource: ResourceIdentity) => {
    dispatchWorkingContext({ type: 'enter_resource', resource });
  }, []);

  const exitResource = useCallback(() => {
    dispatchWorkingContext({ type: 'exit_resource' });
  }, []);

  const executeAction = useCallback(
    async (request: KernelActionExecutionRequest) => {
      const result = await executeKernelActionRequest(request);
      setActionSummary(result.Summary);
      return result;
    },
    [],
  );

  const switchCluster = useCallback((cluster: string) => {
    dispatchWorkingContext({ type: 'request_cluster_switch', cluster });
  }, []);

  const reloadKernelMetadata = useCallback(() => {
    setReloadGeneration((current) => current + 1);
  }, []);

  const value = useMemo<KernelRuntimeContextValue>(
    () => ({
      activeRoute,
      activePage,
      activeActions,
      activeSummarySlots,
      actionSummary,
      activeCluster,
      namespaceScope: selectNamespaceScope(workingContext),
      currentWorkflowDomainId: selectCurrentWorkflowDomain(workingContext),
      currentResource: selectCurrentResource(workingContext),
      kernelSource,
      navigation,
      registrySnapshot,
      resourcePageExtensions: registrySnapshot.resourcePageExtensions,
      navigate,
      reloadKernelMetadata,
      switchCluster,
      enterResource,
      exitResource,
      fetchWorkloadsForDomain,
      executeAction,
    }),
    [
      activeActions,
      activeSummarySlots,
      activePage,
      activeRoute,
      activeCluster,
      actionSummary,
      enterResource,
      exitResource,
      executeAction,
      fetchWorkloadsForDomain,
      kernelSource,
      navigate,
      navigation,
      reloadKernelMetadata,
      registrySnapshot.resourcePageExtensions,
      registrySnapshot,
      switchCluster,
      workingContext,
    ],
  );

  return <KernelRuntimeContext.Provider value={value}>{children}</KernelRuntimeContext.Provider>;
}

export function useKernelRuntime(): KernelRuntimeContextValue {
  const context = useContext(KernelRuntimeContext);
  if (!context) {
    throw new Error('useKernelRuntime must be used within KernelRuntimeProvider');
  }
  return context;
}
