import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { composeKernelNavigation } from './composeKernelNavigation';
import { createLocalKernelSnapshot } from './createLocalKernelSnapshot';
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
  kernelSource: KernelSource;
  navigation: ReturnType<typeof composeKernelNavigation>;
  registrySnapshot: KernelRegistrySnapshot;
  navigate: (route: string) => void;
  fetchWorkloadsForDomain: (workflowDomainId: string, cluster?: string) => Promise<WorkloadItem[]>;
  executeAction: (request: KernelActionExecutionRequest) => Promise<KernelActionExecutionResult>;
}

const KernelRuntimeContext = createContext<KernelRuntimeContextValue | null>(null);

export function KernelRuntimeProvider({ children }: PropsWithChildren) {
  const [activeRoute, setActiveRoute] = useState('/');
  const [runtimeSnapshot, setRuntimeSnapshot] = useState<KernelRegistrySnapshot | null>(null);
  const [kernelSource, setKernelSource] = useState<KernelSource>('loading');
  const [actionSummary, setActionSummary] = useState<string | null>(null);

  const localSnapshot = useMemo(() => createLocalKernelSnapshot(), []);

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
    registrySnapshot.pages.find((page) => page.route === activeRoute) ??
    registrySnapshot.pages[0] ??
    null;
  const activeActions = activePage
    ? resolveWorkflowActions(activePage.workflowDomainId, registrySnapshot.actions)
    : [];
  const activeSummarySlots = activePage
    ? resolveWorkflowSlots(activePage.workflowDomainId, registrySnapshot.slots, 'summary')
    : [];

  const navigate = useCallback((route: string) => {
    setActiveRoute(route);
  }, []);

  const fetchWorkloadsForDomain = useCallback(
    async (workflowDomainId: string, cluster = 'default') => {
      return fetchWorkloads(workflowDomainId, cluster);
    },
    [],
  );

  const executeAction = useCallback(
    async (request: KernelActionExecutionRequest) => {
      const result = await executeKernelActionRequest(request);
      setActionSummary(result.Summary);
      return result;
    },
    [],
  );

  const value = useMemo<KernelRuntimeContextValue>(
    () => ({
      activeRoute,
      activePage,
      activeActions,
      activeSummarySlots,
      actionSummary,
      kernelSource,
      navigation,
      registrySnapshot,
      navigate,
      fetchWorkloadsForDomain,
      executeAction,
    }),
    [
      activeActions,
      activeSummarySlots,
      activePage,
      activeRoute,
      actionSummary,
      executeAction,
      fetchWorkloadsForDomain,
      kernelSource,
      navigate,
      navigation,
      registrySnapshot,
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
