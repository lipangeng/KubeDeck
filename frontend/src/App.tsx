import { useEffect, useState } from 'react';
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
import { discoverFrontendPluginModules } from './kernel/runtime/discoverFrontendPluginModules';
import { fetchKernelMetadata } from './kernel/runtime/fetchKernelMetadata';
import { hydrateRemoteMenuGroups } from './kernel/runtime/hydrateKernelSnapshot';
import {
  fetchMenuPreferences,
  isClusterScopedMenuPreference,
  saveMenuPreferences,
  type MenuPreferenceScope,
  upsertMenuOverride,
} from './kernel/runtime/menuPreferences';
import { KernelRuntimeProvider, useKernelRuntime } from './kernel/runtime/KernelRuntimeContext';
import type { KernelNavigationGroup } from './kernel/runtime/menu/types';
import type { RemoteMenuPreferences } from './kernel/runtime/transport';
import { type ThemePreference } from './themeMode';

interface AppProps {
  themePreference: ThemePreference;
  onThemePreferenceChange: (next: ThemePreference) => void;
  pluginModules?: FrontendCapabilityModule[];
}

type MenuSurface = 'work' | 'system' | 'cluster';

function App({ themePreference, onThemePreferenceChange, pluginModules = [] }: AppProps) {
  const resolvedPluginModules =
    pluginModules.length > 0 ? pluginModules : discoverFrontendPluginModules();

  return (
    <KernelRuntimeProvider pluginModules={resolvedPluginModules}>
      <AppShell
        themePreference={themePreference}
        onThemePreferenceChange={onThemePreferenceChange}
      />
    </KernelRuntimeProvider>
  );
}

function AppShell({ themePreference, onThemePreferenceChange }: AppProps) {
  const {
    activeActions,
    activeCluster,
    activePage,
    actionSummary,
    currentWorkflowDomainId,
    executeAction,
    kernelSource,
    namespaceScope,
    navigate,
    navigation,
    registrySnapshot,
    reloadKernelMetadata,
    switchCluster,
  } = useKernelRuntime();
  const ActiveComponent = activePage?.component;
  const [menuSurface, setMenuSurface] = useState<MenuSurface>('work');
  const [menuPreferences, setMenuPreferences] = useState<RemoteMenuPreferences>({
    globalOverrides: [],
    clusterOverrides: [],
  });
  const [scopedConfigNavigation, setScopedConfigNavigation] = useState<KernelNavigationGroup[]>([]);
  const [editableScope, setEditableScope] = useState<MenuPreferenceScope>('work-global');
  const [settingsMessage, setSettingsMessage] = useState<string | null>(null);

  useEffect(() => {
    setEditableScope(menuSurface === 'cluster' ? 'work-cluster' : 'work-global');
    setSettingsMessage(null);
  }, [menuSurface]);

  useEffect(() => {
    if (menuSurface === 'work') {
      return;
    }
    let active = true;

    async function loadPreferences() {
      try {
        const preferences = await fetchMenuPreferences(activeCluster);
        if (active) {
          setMenuPreferences(preferences);
        }
      } catch {
        if (active) {
          setMenuPreferences({ globalOverrides: [], clusterOverrides: [] });
        }
      }
    }

    void loadPreferences();
    return () => {
      active = false;
    };
  }, [activeCluster, menuSurface]);

  useEffect(() => {
    if (menuSurface === 'work') {
      setScopedConfigNavigation([]);
      return;
    }
    let active = true;

    async function loadScopedNavigation() {
      try {
        const metadata = await fetchKernelMetadata(activeCluster, menuSurface);
        if (active) {
          setScopedConfigNavigation(hydrateRemoteMenuGroups(metadata.menuGroups ?? []));
        }
      } catch {
        if (active) {
          setScopedConfigNavigation([]);
        }
      }
    }

    void loadScopedNavigation();
    return () => {
      active = false;
    };
  }, [activeCluster, menuPreferences, menuSurface]);

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
        cluster: activeCluster,
        namespace: namespaceScope.namespaces[0] ?? 'default',
        scope: 'namespace',
      },
      input: {
        name: 'sample',
      },
    });
  };

  const cycleCluster = () => {
    switchCluster(activeCluster === 'default' ? 'prod-eu1' : 'default');
  };

  const visibleNavigation = menuSurface === 'work' ? navigation : scopedConfigNavigation;
  const editableGroups =
    editableScope === 'work-global' || editableScope === 'work-cluster'
      ? registrySnapshot.menuGroups
      : scopedConfigNavigation;

  const saveMenuOverride = async (
    scope: MenuPreferenceScope,
    update: Parameters<typeof upsertMenuOverride>[2],
  ) => {
    const nextPreferences: RemoteMenuPreferences = isClusterScopedMenuPreference(scope)
      ? {
          ...menuPreferences,
          clusterOverrides: upsertMenuOverride(menuPreferences.clusterOverrides, scope, update),
        }
      : {
          ...menuPreferences,
          globalOverrides: upsertMenuOverride(menuPreferences.globalOverrides, scope, update),
        };
    await saveMenuPreferences(activeCluster, nextPreferences);
    setMenuPreferences(nextPreferences);
    setSettingsMessage(`Saved menu settings for ${scope}`);
    reloadKernelMetadata();
  };

  const handleHideEntry = async (entryKey: string) => {
    await saveMenuOverride(editableScope, (current) => ({
      ...current,
      hiddenEntryKeys: Array.from(new Set([...(current.hiddenEntryKeys ?? []), entryKey])),
    }));
  };

  const handlePinEntry = async (entryKey: string) => {
    await saveMenuOverride(editableScope, (current) => ({
      ...current,
      pinEntryKeys: Array.from(new Set([...(current.pinEntryKeys ?? []), entryKey])),
    }));
  };

  const handleResetScope = async () => {
    await saveMenuOverride(editableScope, () => ({ scope: editableScope }));
  };

  const handleMoveEntry = async (
    groupKey: string,
    entryKey: string,
    direction: 'up' | 'down',
  ) => {
    const group = editableGroups.find((item) => item.key === groupKey);
    if (!group) {
      return;
    }
    const currentOrder = group.entries.map((entry) => entry.entryKey);
    const currentIndex = currentOrder.indexOf(entryKey);
    if (currentIndex < 0) {
      return;
    }
    const targetIndex = direction === 'up' ? currentIndex - 1 : currentIndex + 1;
    if (targetIndex < 0 || targetIndex >= currentOrder.length) {
      return;
    }
    const nextOrder = [...currentOrder];
    const [moved] = nextOrder.splice(currentIndex, 1);
    nextOrder.splice(targetIndex, 0, moved);

    await saveMenuOverride(editableScope, (current) => ({
      ...current,
      itemOrderOverrides: {
        ...(current.itemOrderOverrides ?? {}),
        [groupKey]: nextOrder,
      },
    }));
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
          <Button variant="outlined" onClick={cycleCluster}>
            Cluster: {activeCluster}
          </Button>
          <Button variant="outlined" onClick={() => setMenuSurface('system')}>
            System Settings
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
            {menuSurface === 'work' ? null : (
              <Button variant="outlined" onClick={() => setMenuSurface('work')}>
                Back to Work
              </Button>
            )}
            {visibleNavigation.map((group) => (
              <Stack key={group.key} spacing={1}>
                <Typography variant="caption" sx={{ fontWeight: 700, textTransform: 'uppercase' }}>
                  {group.title.fallback}
                </Typography>
                {group.entries.map((entry) => (
                  <Button
                    key={entry.identity.contributionId}
                    variant={entry.route === activePage?.route && menuSurface === 'work' ? 'contained' : 'outlined'}
                    disabled={entry.availability !== 'enabled'}
                    onClick={() => {
                      if (menuSurface === 'work') {
                        navigate(entry.route ?? '/');
                      }
                    }}
                  >
                    {entry.title.fallback}
                  </Button>
                ))}
              </Stack>
            ))}
            {menuSurface === 'work' ? (
              <>
                <Divider />
                <Button variant="contained" color="secondary" onClick={() => setMenuSurface('cluster')}>
                  Cluster Settings
                </Button>
              </>
            ) : null}
            <Divider />
            <Typography variant="caption" color="text.secondary">
              {copy('app.registeredActions')}:{' '}
              {activeActions.map((action) => action.title.fallback).join(', ') || copy('app.none')}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {copy('app.kernelMetadataSource')}: {kernelSource}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {copy('app.activeCluster')}: {activeCluster}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {copy('app.namespaceScope')}:{' '}
              {namespaceScope.kind === 'single' ? namespaceScope.namespaces.join(', ') : 'all'}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {copy('app.activeWorkflow')}: {currentWorkflowDomainId ?? copy('app.none')}
            </Typography>
            {menuSurface === 'work'
              ? activeActions.map((action) => (
                  <Button
                    key={action.identity.contributionId}
                    variant="text"
                    onClick={() => {
                      void handleExecuteAction(action.actionId);
                    }}
                  >
                    {copy('app.runAction')} {action.title.fallback}
                  </Button>
                ))
              : null}
            {actionSummary ? (
              <Typography variant="caption" color="text.secondary">
                {copy('app.lastActionResult')}: {actionSummary}
              </Typography>
            ) : null}
          </Stack>
        </Paper>

        <Box>
          {menuSurface === 'work' ? (
            ActiveComponent ? <ActiveComponent /> : null
          ) : (
            <MenuSettingsPanel
              editableGroups={editableGroups}
              editableScope={editableScope}
              menuSurface={menuSurface}
              message={settingsMessage}
              onHideEntry={handleHideEntry}
              onMoveEntry={handleMoveEntry}
              onPinEntry={handlePinEntry}
              onResetScope={handleResetScope}
              onScopeChange={setEditableScope}
            />
          )}
        </Box>
      </Box>
    </Box>
  );
}

interface MenuSettingsPanelProps {
  editableGroups: KernelNavigationGroup[];
  editableScope: MenuPreferenceScope;
  menuSurface: Exclude<MenuSurface, 'work'>;
  message: string | null;
  onHideEntry: (entryKey: string) => Promise<void>;
  onMoveEntry: (groupKey: string, entryKey: string, direction: 'up' | 'down') => Promise<void>;
  onPinEntry: (entryKey: string) => Promise<void>;
  onResetScope: () => Promise<void>;
  onScopeChange: (scope: MenuPreferenceScope) => void;
}

function MenuSettingsPanel({
  editableGroups,
  editableScope,
  menuSurface,
  message,
  onHideEntry,
  onMoveEntry,
  onPinEntry,
  onResetScope,
  onScopeChange,
}: MenuSettingsPanelProps) {
  const scopeOptions =
    menuSurface === 'system'
      ? (['work-global', 'system'] as const)
      : (['work-cluster', 'cluster'] as const);

  return (
    <Paper variant="outlined" sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h5">Menu Settings</Typography>
        <Typography variant="body2">Scope: {editableScope}</Typography>
        <Stack direction="row" spacing={1}>
          {scopeOptions.map((scope) => (
            <Button
              key={scope}
              variant={scope === editableScope ? 'contained' : 'outlined'}
              onClick={() => onScopeChange(scope)}
            >
              {scope}
            </Button>
          ))}
        </Stack>
        {editableGroups.map((group) => (
          <Stack key={group.key} spacing={1}>
            <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
              {group.title.fallback}
            </Typography>
            {group.entries.map((entry) => (
              <Stack key={entry.identity.contributionId} direction="row" spacing={1} alignItems="center">
                <Typography sx={{ minWidth: 160 }}>{entry.title.fallback}</Typography>
                <Button variant="outlined" size="small" onClick={() => void onPinEntry(entry.entryKey)}>
                  Pin {entry.title.fallback}
                </Button>
                <Button variant="outlined" size="small" onClick={() => void onHideEntry(entry.entryKey)}>
                  Hide {entry.title.fallback}
                </Button>
                <Button
                  variant="outlined"
                  size="small"
                  onClick={() => void onMoveEntry(group.key, entry.entryKey, 'up')}
                >
                  Move Up {entry.title.fallback}
                </Button>
                <Button
                  variant="outlined"
                  size="small"
                  onClick={() => void onMoveEntry(group.key, entry.entryKey, 'down')}
                >
                  Move Down {entry.title.fallback}
                </Button>
              </Stack>
            ))}
          </Stack>
        ))}
        <Button variant="text" onClick={() => void onResetScope()}>
          Reset Current Scope
        </Button>
        {message ? <Typography color="text.secondary">{message}</Typography> : null}
      </Stack>
    </Paper>
  );
}

export default App;
