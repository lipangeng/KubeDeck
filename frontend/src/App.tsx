import { useEffect, useMemo, useState, type ReactNode } from 'react';
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
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import IconButton from '@mui/material/IconButton';
import Paper from '@mui/material/Paper';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import ChevronRightRoundedIcon from '@mui/icons-material/ChevronRightRounded';
import ExpandMoreRoundedIcon from '@mui/icons-material/ExpandMoreRounded';
import FavoriteRoundedIcon from '@mui/icons-material/FavoriteRounded';
import FavoriteBorderRoundedIcon from '@mui/icons-material/FavoriteBorderRounded';
import GridViewRoundedIcon from '@mui/icons-material/GridViewRounded';
import Inventory2OutlinedIcon from '@mui/icons-material/Inventory2Outlined';
import WorkspacesOutlineIcon from '@mui/icons-material/WorkspacesOutline';
import HistoryRoundedIcon from '@mui/icons-material/HistoryRounded';
import { composeMenus } from './core/menuComposer';
import { groupMenusByGroup } from './core/menuGrouping';
import { translate, type Locale } from './i18n';
import {
  acceptInvite,
  clearAuthToken,
  login,
  me,
  oauthCallback,
  readAuthToken,
  switchTenant,
  writeAuthToken,
  type AuthLoginResponse,
  type AuthTenant,
} from './sdk/authApi';
import {
  parseAuditEventsResponse,
  type AuditEvent,
} from './sdk/auditApi';
import {
  parseGroup,
  parseGroupsResponse,
  parseInvitesResponse,
  parseMembershipsResponse,
  parsePermissionsResponse,
  parseTenantMembersResponse,
  parseTenantsResponse,
  parseUsersResponse,
  type IAMGroup,
  type IAMInvite,
  type IAMMembership,
  type IAMPermission,
  type IAMTenant,
  type IAMUser,
} from './sdk/iamApi';
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
const USER_FAVORITES_KEY = 'kubedeck.user.favorite.menu.ids';
const MENU_OVERRIDES_KEY = 'kubedeck.menu.overrides';

interface AppProps {
  locale: Locale;
  onLocaleChange: (next: Locale) => void;
  themePreference: ThemePreference;
  onThemePreferenceChange: (next: ThemePreference) => void;
}

interface AuthUserState {
  id: string;
  username: string;
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

function readStoredFavoriteMenuIDs(): string[] {
  if (typeof window === 'undefined') {
    return [];
  }
  const raw = window.localStorage.getItem(USER_FAVORITES_KEY);
  if (!raw) {
    return [];
  }
  try {
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) {
      return [];
    }
    return parsed.filter((item): item is string => typeof item === 'string');
  } catch {
    return [];
  }
}

interface MenuOverride {
  group?: string;
  order?: number;
  visible?: boolean;
}

function readStoredMenuOverrides(): Record<string, MenuOverride> {
  if (typeof window === 'undefined') {
    return {};
  }
  const raw = window.localStorage.getItem(MENU_OVERRIDES_KEY);
  if (!raw) {
    return {};
  }
  try {
    const parsed = JSON.parse(raw) as Record<string, unknown>;
    const next: Record<string, MenuOverride> = {};
    for (const [id, value] of Object.entries(parsed)) {
      if (typeof value !== 'object' || value === null) {
        continue;
      }
      const candidate = value as Record<string, unknown>;
      next[id] = {
        group: typeof candidate.group === 'string' ? candidate.group : undefined,
        order: typeof candidate.order === 'number' ? candidate.order : undefined,
        visible: typeof candidate.visible === 'boolean' ? candidate.visible : undefined,
      };
    }
    return next;
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

function routePath(route: string): string {
  return normalizeRoute(route.split('?')[0] ?? '/');
}

function parseInviteToken(route: string): string {
  const [path, query = ''] = route.split('?');
  if (routePath(path) !== '/accept-invite') {
    return '';
  }
  const params = new URLSearchParams(query);
  return params.get('token')?.trim() ?? '';
}

function menuIcon(group: string, targetType: 'page' | 'resource'): ReactNode {
  const normalizedGroup = group.trim().toUpperCase();
  if (normalizedGroup === 'WORKLOAD') {
    return <WorkspacesOutlineIcon fontSize="small" />;
  }
  if (normalizedGroup === 'FAVORITES') {
    return <FavoriteBorderRoundedIcon fontSize="small" />;
  }
  if (targetType === 'resource') {
    return <Inventory2OutlinedIcon fontSize="small" />;
  }
  return <GridViewRoundedIcon fontSize="small" />;
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
  const [manageMenusOpen, setManageMenusOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [healthStatus, setHealthStatus] = useState<ProbeStatus>('checking');
  const [readyStatus, setReadyStatus] = useState<ProbeStatus>('checking');
  const [healthError, setHealthError] = useState<string | null>(null);
  const [readyError, setReadyError] = useState<string | null>(null);
  const [lastCheckedAt, setLastCheckedAt] = useState<string | null>(null);
  const [selectedMenuID, setSelectedMenuID] = useState<string>('');
  const [currentRoute, setCurrentRoute] = useState<string>(() => readHashRoute());
  const [favoriteMenuIDs, setFavoriteMenuIDs] = useState<string[]>(() =>
    readStoredFavoriteMenuIDs(),
  );
  const [menuOverrides, setMenuOverrides] = useState<Record<string, MenuOverride>>(
    () => readStoredMenuOverrides(),
  );
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>(
    () => readStoredExpandedGroups(),
  );
  const [authToken, setAuthToken] = useState<string | null>(() => readAuthToken());
  const [authUser, setAuthUser] = useState<AuthUserState | null>(null);
  const [authRoles, setAuthRoles] = useState<string[]>([]);
  const [authTenants, setAuthTenants] = useState<AuthTenant[]>([]);
  const [activeTenantID, setActiveTenantID] = useState<string>('');
  const [authDialogOpen, setAuthDialogOpen] = useState(false);
  const [loginUsername, setLoginUsername] = useState('admin');
  const [loginPassword, setLoginPassword] = useState('admin');
  const [loginTenantCode, setLoginTenantCode] = useState('dev');
  const [authBusy, setAuthBusy] = useState(false);
  const [authError, setAuthError] = useState<string | null>(null);
  const [tenantBusy, setTenantBusy] = useState(false);
  const [iamDialogOpen, setIamDialogOpen] = useState(false);
  const [iamLoading, setIamLoading] = useState(false);
  const [iamError, setIamError] = useState<string | null>(null);
  const [iamPermissions, setIamPermissions] = useState<IAMPermission[]>([]);
  const [iamGroups, setIamGroups] = useState<IAMGroup[]>([]);
  const [iamMemberships, setIamMemberships] = useState<IAMMembership[]>([]);
  const [iamInvites, setIamInvites] = useState<IAMInvite[]>([]);
  const [iamUsers, setIamUsers] = useState<IAMUser[]>([]);
  const [iamTenants, setIamTenants] = useState<IAMTenant[]>([]);
  const [selectedTenantForMembers, setSelectedTenantForMembers] = useState<string>('');
  const [tenantMembers, setTenantMembers] = useState<IAMMembership[]>([]);
  const [newTenantMemberUserID, setNewTenantMemberUserID] = useState('');
  const [newTenantMemberUserLabel, setNewTenantMemberUserLabel] = useState('');
  const [newTenantMemberEffectiveFrom, setNewTenantMemberEffectiveFrom] = useState('');
  const [newGroupName, setNewGroupName] = useState('');
  const [newGroupDescription, setNewGroupDescription] = useState('');
  const [groupPermissionDrafts, setGroupPermissionDrafts] = useState<Record<string, string>>({});
  const [groupNameDrafts, setGroupNameDrafts] = useState<Record<string, string>>({});
  const [groupDescriptionDrafts, setGroupDescriptionDrafts] = useState<Record<string, string>>({});
  const [membershipGroupDrafts, setMembershipGroupDrafts] = useState<Record<string, string>>({});
  const [membershipEffectiveFromDrafts, setMembershipEffectiveFromDrafts] = useState<
    Record<string, string>
  >({});
  const [membershipEffectiveUntilDrafts, setMembershipEffectiveUntilDrafts] = useState<
    Record<string, string>
  >({});
  const [inviteEmail, setInviteEmail] = useState('');
  const [invitePhone, setInvitePhone] = useState('');
  const [inviteRoleHint, setInviteRoleHint] = useState('member');
  const [inviteExpiresInHours, setInviteExpiresInHours] = useState('72');
  const [inviteCreateBusy, setInviteCreateBusy] = useState(false);
  const [inviteLinkCopied, setInviteLinkCopied] = useState<string | null>(null);
  const [auditDialogOpen, setAuditDialogOpen] = useState(false);
  const [auditLoading, setAuditLoading] = useState(false);
  const [auditError, setAuditError] = useState<string | null>(null);
  const [auditEvents, setAuditEvents] = useState<AuditEvent[]>([]);
  const [auditActionFilter, setAuditActionFilter] = useState('');
  const [auditResultFilter, setAuditResultFilter] = useState('');
  const [auditLimit, setAuditLimit] = useState('50');
  const [inviteUsername, setInviteUsername] = useState('');
  const [invitePassword, setInvitePassword] = useState('');
  const [inviteBusy, setInviteBusy] = useState(false);
  const [inviteStatus, setInviteStatus] = useState<string | null>(null);
  const [inviteError, setInviteError] = useState<string | null>(null);

  function applyMenuOverride(menu: MenuItem): MenuItem {
    const override = menuOverrides[menu.id];
    if (!override) {
      return menu;
    }
    return {
      ...menu,
      group: override.group ?? menu.group,
      order: override.order ?? menu.order,
      visible: override.visible ?? menu.visible,
    };
  }

  const configuredMenus = useMemo(() => {
    const systemMenus = menus
      .filter((menu) => menu.source === 'system')
      .map(applyMenuOverride);
    const userMenus = menus
      .filter((menu) => menu.source === 'user')
      .map(applyMenuOverride);
    return composeMenus(systemMenus, userMenus, []);
  }, [menus, menuOverrides]);
  const manageableMenus = useMemo(() => {
    const systemMenus = menus
      .filter((menu) => menu.source === 'system')
      .map(applyMenuOverride);
    const userMenus = menus
      .filter((menu) => menu.source === 'user')
      .map(applyMenuOverride);
    return [...systemMenus, ...userMenus].sort((a, b) => {
      if (a.order !== b.order) {
        return a.order - b.order;
      }
      return a.id.localeCompare(b.id);
    });
  }, [menus, menuOverrides]);
  const configuredMenuIDSet = useMemo(
    () => new Set(configuredMenus.map((menu) => menu.id)),
    [configuredMenus],
  );
  const isFixedFavorite = (menu: MenuItem): boolean =>
    menu.source === 'user' || menu.group.trim().toUpperCase() === 'FAVORITES';
  const favoritesMenus = useMemo(
    () =>
      configuredMenus.filter(
        (menu) =>
          isFixedFavorite(menu) || favoriteMenuIDs.includes(menu.id),
      ),
    [configuredMenus, favoriteMenuIDs],
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
    setFavoriteMenuIDs((previous) =>
      previous.filter((menuID) => configuredMenuIDSet.has(menuID)),
    );
  }, [configuredMenuIDSet]);

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
    if (typeof window === 'undefined') {
      return;
    }
    window.localStorage.setItem(USER_FAVORITES_KEY, JSON.stringify(favoriteMenuIDs));
  }, [favoriteMenuIDs]);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    window.localStorage.setItem(MENU_OVERRIDES_KEY, JSON.stringify(menuOverrides));
  }, [menuOverrides]);

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
      (menu) => routePath(menu.targetRef) === routePath(currentRoute),
    );
    if (matched && matched.id !== selectedMenuID) {
      setSelectedMenuID(matched.id);
    }
  }, [configuredMenus, currentRoute, selectedMenuID]);

  function navigateMenu(menu: MenuItem) {
    const nextRoute = routePath(menu.targetRef);
    setSelectedMenuID(menu.id);
    setCurrentRoute(nextRoute);
    if (typeof window !== 'undefined') {
      window.location.hash = nextRoute;
    }
  }

  function toggleFavoriteMenu(menu: MenuItem) {
    if (isFixedFavorite(menu)) {
      return;
    }
    setFavoriteMenuIDs((previous) =>
      previous.includes(menu.id)
        ? previous.filter((menuID) => menuID !== menu.id)
        : [...previous, menu.id],
    );
  }

  function updateMenuOverride(menu: MenuItem, patch: MenuOverride) {
    setMenuOverrides((previous) => ({
      ...previous,
      [menu.id]: {
        ...previous[menu.id],
        ...patch,
      },
    }));
  }

  async function applyResources() {
    setApplyStatus(null);
    setApplyResults([]);
    try {
      const response = await apiFetch(
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
        const response = await apiFetch('/api/meta/clusters');
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
          apiFetch(`/api/meta/menus?cluster=${encodeURIComponent(activeCluster)}`),
          apiFetch(`/api/meta/registry?cluster=${encodeURIComponent(activeCluster)}`),
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

  useEffect(() => {
    let active = true;
    if (!authToken) {
      setAuthUser(null);
      setAuthRoles([]);
      setAuthTenants([]);
      setActiveTenantID('');
      return () => {
        active = false;
      };
    }
    const token = authToken;
    async function refreshSession() {
      try {
        const payload = await me(token);
        if (!active) {
          return;
        }
        setAuthUser({
          id: payload.user.id,
          username: payload.user.username,
        });
        setAuthRoles(payload.user.roles);
        setAuthTenants(payload.tenants);
        setActiveTenantID(payload.active_tenant_id);
        setAuthError(null);
      } catch (e) {
        if (!active) {
          return;
        }
        setAuthUser(null);
        setAuthRoles([]);
        setAuthTenants([]);
        setActiveTenantID('');
        clearAuthToken();
        setAuthToken(null);
        setAuthError(e instanceof Error ? e.message : 'auth me failed');
      }
    }
    void refreshSession();
    return () => {
      active = false;
    };
  }, [authToken]);

  function buildHeaders(base?: HeadersInit): HeadersInit {
    if (!authToken) {
      return base ?? {};
    }
    return {
      ...(base ?? {}),
      Authorization: `Bearer ${authToken}`,
    };
  }

  function handleUnauthorized() {
    clearAuthToken();
    setAuthToken(null);
    setAuthUser(null);
    setAuthRoles([]);
    setAuthTenants([]);
    setActiveTenantID('');
    setAuthError('unauthorized');
    setAuthDialogOpen(true);
  }

  async function apiFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
    const response = await fetch(input, {
      ...init,
      headers: buildHeaders(init?.headers),
    });
    if (response.status === 401) {
      handleUnauthorized();
      throw new Error('unauthorized');
    }
    return response;
  }

  async function submitLogin() {
    setAuthBusy(true);
    setAuthError(null);
    try {
      const payload = await login(loginUsername, loginPassword, loginTenantCode);
      applyAuthPayload(payload);
      setAuthDialogOpen(false);
    } catch (e) {
      setAuthError(e instanceof Error ? e.message : 'auth login failed');
    } finally {
      setAuthBusy(false);
    }
  }

  async function submitOAuthDemoLogin() {
    setAuthBusy(true);
    setAuthError(null);
    try {
      const payload = await oauthCallback('oauth-admin', loginTenantCode);
      applyAuthPayload(payload);
      setAuthDialogOpen(false);
    } catch (e) {
      setAuthError(e instanceof Error ? e.message : 'oauth callback failed');
    } finally {
      setAuthBusy(false);
    }
  }

  function applyAuthPayload(payload: AuthLoginResponse) {
    writeAuthToken(payload.token);
    setAuthToken(payload.token);
    setAuthUser({
      id: payload.user.id,
      username: payload.user.username,
    });
    setAuthRoles(payload.user.roles);
    setAuthTenants(payload.tenants);
    setActiveTenantID(payload.active_tenant_id);
  }

  async function logout() {
    if (authToken) {
      try {
        await fetch('/api/auth/logout', {
          method: 'POST',
          headers: buildHeaders(),
        });
      } catch {
        // best effort logout
      }
    }
    clearAuthToken();
    setAuthToken(null);
    setAuthUser(null);
    setAuthRoles([]);
    setAuthTenants([]);
    setActiveTenantID('');
    setAuthError(null);
  }

  async function onTenantCodeChange(nextTenantCode: string) {
    if (!authToken || !authUser) {
      return;
    }
    setTenantBusy(true);
    setAuthError(null);
    try {
      const payload = await switchTenant(authToken, nextTenantCode);
      setActiveTenantID(payload.active_tenant_id);
    } catch (e) {
      setAuthError(e instanceof Error ? e.message : 'switch tenant failed');
    } finally {
      setTenantBusy(false);
    }
  }

  function normalizePermissionsDraft(value: string): string[] {
    return Array.from(
      new Set(
        value
          .split(',')
          .map((item) => item.trim())
          .filter((item) => item !== ''),
      ),
    );
  }

  async function loadIAMData() {
    if (!authUser) {
      setIamError('auth_required');
      setIamPermissions([]);
      setIamGroups([]);
      setIamMemberships([]);
      setIamInvites([]);
      setIamUsers([]);
      setIamTenants([]);
      setTenantMembers([]);
      setSelectedTenantForMembers('');
      return;
    }
    setIamLoading(true);
    setIamError(null);
    try {
      const [permissionsResponse, groupsResponse, membershipsResponse, invitesResponse, usersResponse, tenantsResponse] = await Promise.all([
        apiFetch('/api/iam/permissions'),
        apiFetch('/api/iam/groups'),
        apiFetch('/api/iam/memberships'),
        apiFetch('/api/iam/invites'),
        apiFetch('/api/iam/users'),
        apiFetch('/api/iam/tenants'),
      ]);
      if (!permissionsResponse.ok) {
        throw new Error(`permissions request failed: ${permissionsResponse.status}`);
      }
      if (!groupsResponse.ok) {
        throw new Error(`groups request failed: ${groupsResponse.status}`);
      }
      if (!membershipsResponse.ok) {
        throw new Error(`memberships request failed: ${membershipsResponse.status}`);
      }
      if (!invitesResponse.ok) {
        throw new Error(`invites request failed: ${invitesResponse.status}`);
      }
      if (!usersResponse.ok) {
        throw new Error(`users request failed: ${usersResponse.status}`);
      }
      if (!tenantsResponse.ok) {
        throw new Error(`tenants request failed: ${tenantsResponse.status}`);
      }
      const permissions = parsePermissionsResponse(await permissionsResponse.json());
      const groups = parseGroupsResponse(await groupsResponse.json());
      const memberships = parseMembershipsResponse(await membershipsResponse.json());
      const invites = parseInvitesResponse(await invitesResponse.json());
      const users = parseUsersResponse(await usersResponse.json());
      const tenants = parseTenantsResponse(await tenantsResponse.json());
      setIamPermissions(permissions);
      setIamGroups(groups);
      setIamMemberships(memberships);
      setIamInvites(invites);
      setIamUsers(users);
      setIamTenants(tenants);
      const defaultTenantID = selectedTenantForMembers !== '' ? selectedTenantForMembers : (tenants[0]?.id ?? '');
      setSelectedTenantForMembers(defaultTenantID);
      if (defaultTenantID !== '') {
        const tenantMembersResponse = await apiFetch(`/api/iam/tenants/${encodeURIComponent(defaultTenantID)}/members`);
        if (tenantMembersResponse.ok) {
          setTenantMembers(parseTenantMembersResponse(await tenantMembersResponse.json()));
        } else {
          setTenantMembers([]);
        }
      } else {
        setTenantMembers([]);
      }
      setGroupPermissionDrafts(
        groups.reduce<Record<string, string>>((acc, group) => {
          acc[group.id] = group.permissions.join(', ');
          return acc;
        }, {}),
      );
      setGroupNameDrafts(
        groups.reduce<Record<string, string>>((acc, group) => {
          acc[group.id] = group.name;
          return acc;
        }, {}),
      );
      setGroupDescriptionDrafts(
        groups.reduce<Record<string, string>>((acc, group) => {
          acc[group.id] = group.description;
          return acc;
        }, {}),
      );
      setMembershipGroupDrafts(
        memberships.reduce<Record<string, string>>((acc, membership) => {
          acc[membership.id] = membership.groupIDs.join(', ');
          return acc;
        }, {}),
      );
      setMembershipEffectiveFromDrafts(
        memberships.reduce<Record<string, string>>((acc, membership) => {
          acc[membership.id] = membership.effectiveFrom;
          return acc;
        }, {}),
      );
      setMembershipEffectiveUntilDrafts(
        memberships.reduce<Record<string, string>>((acc, membership) => {
          acc[membership.id] = membership.effectiveUntil;
          return acc;
        }, {}),
      );
    } catch (e) {
      setIamError(e instanceof Error ? e.message : 'load iam failed');
    } finally {
      setIamLoading(false);
    }
  }

  async function createIAMGroup() {
    if (!authUser) {
      setIamError('auth_required');
      return;
    }
    if (newGroupName.trim() === '') {
      setIamError('name_required');
      return;
    }
    setIamError(null);
    try {
      const response = await apiFetch('/api/iam/groups', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: newGroupName.trim(),
          description: newGroupDescription.trim(),
        }),
      });
      if (!response.ok) {
        throw new Error(`create group failed: ${response.status}`);
      }
      const group = parseGroup((await response.json()) as Record<string, unknown>);
      setIamGroups((previous) => [...previous, group]);
      setGroupPermissionDrafts((previous) => ({
        ...previous,
        [group.id]: group.permissions.join(', '),
      }));
      setGroupNameDrafts((previous) => ({
        ...previous,
        [group.id]: group.name,
      }));
      setGroupDescriptionDrafts((previous) => ({
        ...previous,
        [group.id]: group.description,
      }));
      setNewGroupName('');
      setNewGroupDescription('');
    } catch (e) {
      setIamError(e instanceof Error ? e.message : 'create group failed');
    }
  }

  async function saveGroupPermissions(group: IAMGroup) {
    if (!authUser) {
      setIamError('auth_required');
      return;
    }
    setIamError(null);
    try {
      const permissions = normalizePermissionsDraft(groupPermissionDrafts[group.id] ?? '');
      const response = await apiFetch(`/api/iam/groups/${encodeURIComponent(group.id)}/permissions`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ permissions }),
      });
      if (!response.ok) {
        throw new Error(`save permissions failed: ${response.status}`);
      }
      const updated = parseGroup((await response.json()) as Record<string, unknown>);
      setIamGroups((previous) => previous.map((item) => (item.id === updated.id ? updated : item)));
      setGroupPermissionDrafts((previous) => ({
        ...previous,
        [updated.id]: updated.permissions.join(', '),
      }));
    } catch (e) {
      setIamError(e instanceof Error ? e.message : 'save permissions failed');
    }
  }

  async function saveGroupProfile(group: IAMGroup) {
    if (!authUser) {
      setIamError('auth_required');
      return;
    }
    setIamError(null);
    try {
      const response = await apiFetch(`/api/iam/groups/${encodeURIComponent(group.id)}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: groupNameDrafts[group.id] ?? group.name,
          description: groupDescriptionDrafts[group.id] ?? group.description,
        }),
      });
      if (!response.ok) {
        throw new Error(`save group profile failed: ${response.status}`);
      }
      const updated = parseGroup((await response.json()) as Record<string, unknown>);
      setIamGroups((previous) => previous.map((item) => (item.id === updated.id ? updated : item)));
      setGroupNameDrafts((previous) => ({
        ...previous,
        [updated.id]: updated.name,
      }));
      setGroupDescriptionDrafts((previous) => ({
        ...previous,
        [updated.id]: updated.description,
      }));
    } catch (e) {
      setIamError(e instanceof Error ? e.message : 'save group profile failed');
    }
  }

  async function deleteGroup(group: IAMGroup) {
    if (!authUser) {
      setIamError('auth_required');
      return;
    }
    setIamError(null);
    try {
      const response = await apiFetch(`/api/iam/groups/${encodeURIComponent(group.id)}`, {
        method: 'DELETE',
      });
      if (!response.ok) {
        throw new Error(`delete group failed: ${response.status}`);
      }
      setIamGroups((previous) => previous.filter((item) => item.id !== group.id));
      setGroupPermissionDrafts((previous) => {
        const next = { ...previous };
        delete next[group.id];
        return next;
      });
      setGroupNameDrafts((previous) => {
        const next = { ...previous };
        delete next[group.id];
        return next;
      });
      setGroupDescriptionDrafts((previous) => {
        const next = { ...previous };
        delete next[group.id];
        return next;
      });
    } catch (e) {
      setIamError(e instanceof Error ? e.message : 'delete group failed');
    }
  }

  async function saveMembershipGroups(membership: IAMMembership) {
    if (!authUser) {
      setIamError('auth_required');
      return;
    }
    setIamError(null);
    try {
      const groupIDs = normalizePermissionsDraft(membershipGroupDrafts[membership.id] ?? '');
      const response = await apiFetch(
        `/api/iam/memberships/${encodeURIComponent(membership.id)}/groups`,
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ group_ids: groupIDs }),
        },
      );
      if (!response.ok) {
        throw new Error(`save membership groups failed: ${response.status}`);
      }
      const updated = parseMembershipsResponse({
        memberships: [await response.json()],
      })[0];
      if (!updated) {
        throw new Error('invalid membership update response');
      }
      setIamMemberships((previous) =>
        previous.map((item) => (item.id === updated.id ? updated : item)),
      );
      setMembershipGroupDrafts((previous) => ({
        ...previous,
        [updated.id]: updated.groupIDs.join(', '),
      }));
    } catch (e) {
      setIamError(e instanceof Error ? e.message : 'save membership groups failed');
    }
  }

  async function saveMembershipValidity(membership: IAMMembership) {
    if (!authUser) {
      setIamError('auth_required');
      return;
    }
    setIamError(null);
    try {
      const effectiveFrom = membershipEffectiveFromDrafts[membership.id] ?? '';
      const effectiveUntil = membershipEffectiveUntilDrafts[membership.id] ?? '';
      const response = await apiFetch(
        `/api/iam/memberships/${encodeURIComponent(membership.id)}/validity`,
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            effective_from: effectiveFrom,
            effective_until: effectiveUntil,
          }),
        },
      );
      if (!response.ok) {
        throw new Error(`save membership validity failed: ${response.status}`);
      }
      const updated = parseMembershipsResponse({
        memberships: [await response.json()],
      })[0];
      if (!updated) {
        throw new Error('invalid membership validity response');
      }
      setIamMemberships((previous) =>
        previous.map((item) => (item.id === updated.id ? updated : item)),
      );
      setMembershipEffectiveFromDrafts((previous) => ({
        ...previous,
        [updated.id]: updated.effectiveFrom,
      }));
      setMembershipEffectiveUntilDrafts((previous) => ({
        ...previous,
        [updated.id]: updated.effectiveUntil,
      }));
    } catch (e) {
      setIamError(e instanceof Error ? e.message : 'save membership validity failed');
    }
  }

  async function createInvite() {
    if (!authUser) {
      setIamError('auth_required');
      return;
    }
    if (inviteEmail.trim() === '' && invitePhone.trim() === '') {
      setIamError('email_or_phone_required');
      return;
    }
    setIamError(null);
    setInviteCreateBusy(true);
    try {
      const response = await apiFetch('/api/iam/invites', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: inviteEmail.trim(),
          phone: invitePhone.trim(),
          role_hint: inviteRoleHint.trim(),
          expires_in_hours: Number(inviteExpiresInHours) || 72,
        }),
      });
      if (!response.ok) {
        throw new Error(`create invite failed: ${response.status}`);
      }
      const created = parseInvitesResponse({
        invites: [await response.json()],
      })[0];
      if (!created) {
        throw new Error('invalid invite response');
      }
      setIamInvites((previous) => [created, ...previous]);
      setInviteEmail('');
      setInvitePhone('');
      setInviteRoleHint('member');
      setInviteExpiresInHours('72');
    } catch (e) {
      setIamError(e instanceof Error ? e.message : 'create invite failed');
    } finally {
      setInviteCreateBusy(false);
    }
  }

  async function copyInviteLink(link: string) {
    if (typeof navigator === 'undefined' || !navigator.clipboard) {
      setInviteLinkCopied('clipboard_unavailable');
      return;
    }
    try {
      await navigator.clipboard.writeText(link);
      setInviteLinkCopied(link);
    } catch {
      setInviteLinkCopied('clipboard_failed');
    }
  }

  async function revokeInvite(invite: IAMInvite) {
    if (!authUser) {
      setIamError('auth_required');
      return;
    }
    setIamError(null);
    try {
      const response = await apiFetch(`/api/iam/invites/${encodeURIComponent(invite.id)}`, {
        method: 'DELETE',
      });
      if (!response.ok) {
        throw new Error(`revoke invite failed: ${response.status}`);
      }
      const updated = parseInvitesResponse({
        invites: [await response.json()],
      })[0];
      if (!updated) {
        throw new Error('invalid revoke invite response');
      }
      setIamInvites((previous) =>
        previous.map((item) => (item.id === updated.id ? updated : item)),
      );
    } catch (e) {
      setIamError(e instanceof Error ? e.message : 'revoke invite failed');
    }
  }

  async function loadAuditData() {
    if (!authUser) {
      setAuditError('auth_required');
      setAuditEvents([]);
      return;
    }
    setAuditLoading(true);
    setAuditError(null);
    try {
      const params = new URLSearchParams();
      if (auditActionFilter.trim() !== '') {
        params.set('action', auditActionFilter.trim());
      }
      if (auditResultFilter.trim() !== '') {
        params.set('result', auditResultFilter.trim());
      }
      if (auditLimit.trim() !== '') {
        params.set('limit', auditLimit.trim());
      }
      const query = params.toString();
      const response = await apiFetch(`/api/audit/events${query ? `?${query}` : ''}`);
      if (!response.ok) {
        throw new Error(`audit request failed: ${response.status}`);
      }
      const events = parseAuditEventsResponse(await response.json());
      setAuditEvents(events);
    } catch (e) {
      setAuditError(e instanceof Error ? e.message : 'load audit failed');
    } finally {
      setAuditLoading(false);
    }
  }

  async function loadTenantMembers(tenantID: string) {
    setSelectedTenantForMembers(tenantID);
    if (tenantID.trim() === '') {
      setTenantMembers([]);
      return;
    }
    try {
      const response = await apiFetch(`/api/iam/tenants/${encodeURIComponent(tenantID)}/members`);
      if (!response.ok) {
        throw new Error(`tenant members request failed: ${response.status}`);
      }
      setTenantMembers(parseTenantMembersResponse(await response.json()));
    } catch (e) {
      setIamError(e instanceof Error ? e.message : 'load tenant members failed');
      setTenantMembers([]);
    }
  }

  async function createTenantMember() {
    if (!authUser || selectedTenantForMembers.trim() === '') {
      setIamError('tenant_required');
      return;
    }
    if (newTenantMemberUserID.trim() === '') {
      setIamError('user_id_required');
      return;
    }
    setIamError(null);
    try {
      const response = await apiFetch(`/api/iam/tenants/${encodeURIComponent(selectedTenantForMembers)}/members`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: newTenantMemberUserID.trim(),
          user_label: newTenantMemberUserLabel.trim(),
          effective_from:
            newTenantMemberEffectiveFrom.trim() === ''
              ? new Date().toISOString()
              : newTenantMemberEffectiveFrom.trim(),
        }),
      });
      if (!response.ok) {
        throw new Error(`create tenant member failed: ${response.status}`);
      }
      setNewTenantMemberUserID('');
      setNewTenantMemberUserLabel('');
      setNewTenantMemberEffectiveFrom('');
      await loadTenantMembers(selectedTenantForMembers);
    } catch (e) {
      setIamError(e instanceof Error ? e.message : 'create tenant member failed');
    }
  }

  async function deleteTenantMember(member: IAMMembership) {
    if (!authUser || selectedTenantForMembers.trim() === '') {
      return;
    }
    setIamError(null);
    try {
      const response = await apiFetch(
        `/api/iam/tenants/${encodeURIComponent(selectedTenantForMembers)}/members?membership_id=${encodeURIComponent(member.id)}`,
        {
          method: 'DELETE',
        },
      );
      if (!response.ok) {
        throw new Error(`delete tenant member failed: ${response.status}`);
      }
      setTenantMembers((previous) => previous.filter((item) => item.id !== member.id));
    } catch (e) {
      setIamError(e instanceof Error ? e.message : 'delete tenant member failed');
    }
  }

  function membershipStatus(membership: IAMMembership): { label: string; color: 'success' | 'warning' | 'error' } {
    const now = Date.now();
    const until = membership.effectiveUntil ? Date.parse(membership.effectiveUntil) : Number.NaN;
    if (!Number.isNaN(until)) {
      if (until <= now) {
        return { label: t('membershipExpired'), color: 'error' };
      }
      const diffHours = (until - now) / (1000 * 60 * 60);
      if (diffHours <= 24 * 7) {
        return { label: t('membershipExpiringSoon'), color: 'warning' };
      }
    }
    return { label: t('membershipActive'), color: 'success' };
  }

  async function submitAcceptInvite(inviteToken: string) {
    if (inviteToken === '') {
      return;
    }
    setInviteBusy(true);
    setInviteError(null);
    setInviteStatus(null);
    try {
      const payload = await acceptInvite(inviteToken, inviteUsername, invitePassword);
      setInviteStatus(payload.status);
    } catch (e) {
      setInviteError(e instanceof Error ? e.message : 'accept invite failed');
    } finally {
      setInviteBusy(false);
    }
  }

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
  const activeTenantCode =
    authTenants.find((tenant) => tenant.id === activeTenantID)?.code || authTenants[0]?.code || '';
  const inviteToken = parseInviteToken(currentRoute);
  const acceptInviteOpen = inviteToken !== '';
  const authCanWrite =
    authRoles.length === 0 || authRoles.includes('admin') || authRoles.includes('owner');

  useEffect(() => {
    if (!iamDialogOpen) {
      return;
    }
    void loadIAMData();
  }, [iamDialogOpen, authUser?.id, activeTenantID]);

  useEffect(() => {
    if (!auditDialogOpen) {
      return;
    }
    void loadAuditData();
  }, [auditDialogOpen, authUser?.id, activeTenantID]);

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
          <Button variant="outlined" onClick={() => setManageMenusOpen(true)}>
            {t('manageMenus')}
          </Button>
          <Button variant="outlined" onClick={() => setIamDialogOpen(true)}>
            {t('accessControl')}
          </Button>
          <Button
            variant="outlined"
            startIcon={<HistoryRoundedIcon />}
            onClick={() => setAuditDialogOpen(true)}
          >
            {t('auditEvents')}
          </Button>
          {authUser ? (
            <>
              <Chip size="small" variant="outlined" label={authUser.username} sx={{ borderRadius: 999, bgcolor: 'background.paper' }} />
              <FormControl size="small" sx={{ minWidth: 130 }}>
                <InputLabel htmlFor="tenant-select">{t('tenant')}</InputLabel>
                <Select
                  native
                  value={activeTenantCode}
                  disabled={tenantBusy}
                  onChange={(event) => void onTenantCodeChange(event.target.value)}
                  label={t('tenant')}
                  inputProps={{ id: 'tenant-select' }}
                >
                  {authTenants.map((tenant) => (
                    <option key={tenant.id} value={tenant.code}>
                      {tenant.code}
                    </option>
                  ))}
                </Select>
              </FormControl>
              <Button color="inherit" onClick={() => void logout()}>
                {t('logout')}
              </Button>
            </>
          ) : (
            <Button variant="outlined" onClick={() => setAuthDialogOpen(true)}>
              {t('login')}
            </Button>
          )}
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
                        routePath(menu.targetRef) === routePath(currentRoute)
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
                      <ListItemIcon sx={{ minWidth: 30, color: 'inherit' }}>
                        {menuIcon(menu.group, menu.targetType)}
                      </ListItemIcon>
                      <ListItemText primary={menu.title} />
                      <Tooltip title={isFixedFavorite(menu) ? t('fixedFavorite') : t('favoriteRemove')}>
                        <span>
                          <IconButton
                            size="small"
                            onClick={(event) => {
                              event.stopPropagation();
                              toggleFavoriteMenu(menu);
                            }}
                            disabled={isFixedFavorite(menu)}
                            aria-label={isFixedFavorite(menu) ? t('fixedFavorite') : t('favoriteRemove')}
                          >
                            <FavoriteRoundedIcon fontSize="small" />
                          </IconButton>
                        </span>
                      </Tooltip>
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
                    <ListItemIcon sx={{ minWidth: 26, color: 'text.secondary' }}>
                      {expandedGroups[group.name] ? (
                        <ExpandMoreRoundedIcon fontSize="small" />
                      ) : (
                        <ChevronRightRoundedIcon fontSize="small" />
                      )}
                    </ListItemIcon>
                    <ListItemText
                      primary={group.name}
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
                              routePath(item.targetRef) === routePath(currentRoute)
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
                            <ListItemIcon sx={{ minWidth: 30, color: 'inherit' }}>
                              {menuIcon(item.group, item.targetType)}
                            </ListItemIcon>
                            <ListItemText primary={item.title} />
                            <Tooltip
                              title={
                                favoriteMenuIDs.includes(item.id)
                                  ? t('favoriteRemove')
                                  : t('favoriteAdd')
                              }
                            >
                              <IconButton
                                size="small"
                                onClick={(event) => {
                                  event.stopPropagation();
                                  toggleFavoriteMenu(item);
                                }}
                                aria-label={
                                  favoriteMenuIDs.includes(item.id)
                                    ? t('favoriteRemove')
                                    : t('favoriteAdd')
                                }
                              >
                                {favoriteMenuIDs.includes(item.id) ? (
                                  <FavoriteRoundedIcon fontSize="small" />
                                ) : (
                                  <FavoriteBorderRoundedIcon fontSize="small" />
                                )}
                              </IconButton>
                            </Tooltip>
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
        open={iamDialogOpen}
        onClose={() => setIamDialogOpen(false)}
        fullWidth
        maxWidth="md"
      >
        <DialogTitle>{t('accessControl')}</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={1.2}>
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
              <TextField
                size="small"
                label={t('groupName')}
                value={newGroupName}
                onChange={(event) => setNewGroupName(event.target.value)}
              />
              <TextField
                size="small"
                label={t('groupDescription')}
                value={newGroupDescription}
                onChange={(event) => setNewGroupDescription(event.target.value)}
              />
              <Button variant="contained" onClick={() => void createIAMGroup()} disabled={!authCanWrite}>
                {t('createGroup')}
              </Button>
              <Button onClick={() => void loadIAMData()}>{t('refresh')}</Button>
            </Stack>
            {iamLoading ? <Typography>{t('loading')}</Typography> : null}
            {iamError ? (
              <Typography color="error">{t('iamLoadError', { error: iamError })}</Typography>
            ) : null}

            <Paper variant="outlined" sx={{ p: 1 }}>
              <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 0.8 }}>
                {t('iamPermissionsCatalog')}
              </Typography>
              <Stack direction="row" spacing={0.8} useFlexGap flexWrap="wrap">
                {iamPermissions.map((permission) => (
                  <Chip
                    key={permission.code}
                    size="small"
                    variant="outlined"
                    label={`${permission.code} (${permission.scope})`}
                  />
                ))}
              </Stack>
            </Paper>

            <Stack spacing={1}>
              <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                {t('iamGroups')}
              </Typography>
              {iamGroups.map((group) => (
                <Paper key={group.id} variant="outlined" sx={{ p: 1 }}>
                  <TextField
                    size="small"
                    fullWidth
                    label={t('groupName')}
                    value={groupNameDrafts[group.id] ?? group.name}
                    onChange={(event) =>
                      setGroupNameDrafts((previous) => ({
                        ...previous,
                        [group.id]: event.target.value,
                      }))
                    }
                  />
                  <TextField
                    size="small"
                    fullWidth
                    sx={{ mt: 1 }}
                    label={t('groupDescription')}
                    value={groupDescriptionDrafts[group.id] ?? group.description}
                    onChange={(event) =>
                      setGroupDescriptionDrafts((previous) => ({
                        ...previous,
                        [group.id]: event.target.value,
                      }))
                    }
                  />
                  <TextField
                    size="small"
                    fullWidth
                    sx={{ mt: 1 }}
                    label={t('groupPermissions')}
                    value={groupPermissionDrafts[group.id] ?? ''}
                    onChange={(event) =>
                      setGroupPermissionDrafts((previous) => ({
                        ...previous,
                        [group.id]: event.target.value,
                      }))
                    }
                  />
                  <Stack direction="row" justifyContent="flex-end" sx={{ mt: 1 }}>
                    <Button
                      size="small"
                      variant="outlined"
                      onClick={() => void saveGroupProfile(group)}
                      disabled={!authCanWrite}
                    >
                      {t('saveGroup')}
                    </Button>
                    <Button
                      size="small"
                      variant="outlined"
                      onClick={() => void saveGroupPermissions(group)}
                      disabled={!authCanWrite}
                    >
                      {t('savePermissions')}
                    </Button>
                    <Button
                      size="small"
                      color="error"
                      variant="outlined"
                      onClick={() => void deleteGroup(group)}
                      disabled={!authCanWrite}
                    >
                      {t('deleteGroup')}
                    </Button>
                  </Stack>
                </Paper>
              ))}
              {iamGroups.length === 0 && !iamLoading ? (
                <Typography color="text.secondary">{t('noEntries')}</Typography>
              ) : null}
            </Stack>

            <Stack spacing={1}>
              <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                {t('iamUsers')}
              </Typography>
              {iamUsers.map((user) => (
                <Paper key={`${user.id}-${user.membershipID}`} variant="outlined" sx={{ p: 1 }}>
                  <Typography variant="subtitle2">{user.username || user.id}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {user.roles.join(', ') || 'viewer'} | {user.membershipID}
                  </Typography>
                  <Typography variant="caption" color="text.secondary" display="block">
                    {user.effectiveFrom}
                    {user.effectiveUntil ? ` ~ ${user.effectiveUntil}` : ''}
                  </Typography>
                </Paper>
              ))}
              {iamUsers.length === 0 && !iamLoading ? (
                <Typography color="text.secondary">{t('noEntries')}</Typography>
              ) : null}
            </Stack>

            <Stack spacing={1}>
              <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                {t('iamTenantMembers')}
              </Typography>
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
                <FormControl size="small" sx={{ minWidth: 180 }}>
                  <InputLabel htmlFor="tenant-members-select">{t('tenant')}</InputLabel>
                  <Select
                    native
                    label={t('tenant')}
                    value={selectedTenantForMembers}
                    onChange={(event) => void loadTenantMembers(event.target.value)}
                    inputProps={{ id: 'tenant-members-select' }}
                  >
                    <option value=""></option>
                    {iamTenants.map((tenant) => (
                      <option key={tenant.id} value={tenant.id}>
                        {tenant.code} ({tenant.name})
                      </option>
                    ))}
                  </Select>
                </FormControl>
                <TextField
                  size="small"
                  label={t('userId')}
                  value={newTenantMemberUserID}
                  onChange={(event) => setNewTenantMemberUserID(event.target.value)}
                />
                <TextField
                  size="small"
                  label={t('username')}
                  value={newTenantMemberUserLabel}
                  onChange={(event) => setNewTenantMemberUserLabel(event.target.value)}
                />
                <TextField
                  size="small"
                  label={t('membershipEffectiveFrom')}
                  value={newTenantMemberEffectiveFrom}
                  onChange={(event) => setNewTenantMemberEffectiveFrom(event.target.value)}
                  placeholder="2026-02-25T00:00:00Z"
                />
                <Button
                  variant="contained"
                  disabled={!authCanWrite || selectedTenantForMembers.trim() === ''}
                  onClick={() => void createTenantMember()}
                >
                  {t('addTenantMember')}
                </Button>
              </Stack>
              {tenantMembers.map((member) => (
                <Paper key={member.id} variant="outlined" sx={{ p: 1 }}>
                  <Typography variant="subtitle2">{member.userLabel || member.userID}</Typography>
                  <Typography variant="caption" color="text.secondary" display="block">
                    {member.id} | {member.effectiveFrom}
                    {member.effectiveUntil ? ` ~ ${member.effectiveUntil}` : ''}
                  </Typography>
                  <Stack direction="row" justifyContent="flex-end" sx={{ mt: 1 }}>
                    <Button
                      size="small"
                      color="error"
                      variant="outlined"
                      disabled={!authCanWrite}
                      onClick={() => void deleteTenantMember(member)}
                    >
                      {t('removeTenantMember')}
                    </Button>
                  </Stack>
                </Paper>
              ))}
              {tenantMembers.length === 0 && !iamLoading ? (
                <Typography color="text.secondary">{t('noEntries')}</Typography>
              ) : null}
            </Stack>

            <Stack spacing={1}>
              <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                {t('iamMemberships')}
              </Typography>
              {iamMemberships.map((membership) => (
                <Paper key={membership.id} variant="outlined" sx={{ p: 1 }}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Typography variant="subtitle2">{membership.userLabel || membership.userID}</Typography>
                    <Chip
                      size="small"
                      color={membershipStatus(membership).color}
                      label={membershipStatus(membership).label}
                    />
                  </Stack>
                  <Typography variant="caption" color="text.secondary">
                    {membership.id}
                  </Typography>
                  <TextField
                    size="small"
                    fullWidth
                    sx={{ mt: 1 }}
                    label={t('membershipGroups')}
                    value={membershipGroupDrafts[membership.id] ?? ''}
                    onChange={(event) =>
                      setMembershipGroupDrafts((previous) => ({
                        ...previous,
                        [membership.id]: event.target.value,
                      }))
                    }
                  />
                  <Stack direction="row" justifyContent="flex-end" sx={{ mt: 1 }}>
                    <Button
                      size="small"
                      variant="outlined"
                      onClick={() => void saveMembershipGroups(membership)}
                      disabled={!authCanWrite}
                    >
                      {t('saveMembershipGroups')}
                    </Button>
                  </Stack>
                  <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ mt: 1 }}>
                    <TextField
                      size="small"
                      fullWidth
                      label={t('membershipEffectiveFrom')}
                      value={membershipEffectiveFromDrafts[membership.id] ?? ''}
                      onChange={(event) =>
                        setMembershipEffectiveFromDrafts((previous) => ({
                          ...previous,
                          [membership.id]: event.target.value,
                        }))
                      }
                    />
                    <TextField
                      size="small"
                      fullWidth
                      label={t('membershipEffectiveUntil')}
                      value={membershipEffectiveUntilDrafts[membership.id] ?? ''}
                      onChange={(event) =>
                        setMembershipEffectiveUntilDrafts((previous) => ({
                          ...previous,
                          [membership.id]: event.target.value,
                        }))
                      }
                      placeholder={t('membershipEffectiveUntilPlaceholder')}
                    />
                  </Stack>
                  <Stack direction="row" justifyContent="flex-end" sx={{ mt: 1 }}>
                    <Button
                      size="small"
                      variant="outlined"
                      onClick={() => void saveMembershipValidity(membership)}
                      disabled={!authCanWrite}
                    >
                      {t('saveMembershipValidity')}
                    </Button>
                  </Stack>
                </Paper>
              ))}
              {iamMemberships.length === 0 && !iamLoading ? (
                <Typography color="text.secondary">{t('noEntries')}</Typography>
              ) : null}
            </Stack>

            <Stack spacing={1}>
              <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                {t('iamInvites')}
              </Typography>
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
                <TextField
                  size="small"
                  label={t('inviteEmail')}
                  value={inviteEmail}
                  onChange={(event) => setInviteEmail(event.target.value)}
                />
                <TextField
                  size="small"
                  label={t('invitePhone')}
                  value={invitePhone}
                  onChange={(event) => setInvitePhone(event.target.value)}
                />
                <TextField
                  size="small"
                  label={t('inviteRoleHint')}
                  value={inviteRoleHint}
                  onChange={(event) => setInviteRoleHint(event.target.value)}
                />
                <TextField
                  size="small"
                  type="number"
                  label={t('inviteExpiresInHours')}
                  value={inviteExpiresInHours}
                  onChange={(event) => setInviteExpiresInHours(event.target.value)}
                />
                <Button
                  variant="contained"
                  onClick={() => void createInvite()}
                  disabled={inviteCreateBusy || !authCanWrite}
                >
                  {t('createInvite')}
                </Button>
              </Stack>
              {!authCanWrite ? (
                <Typography variant="caption" color="warning.main">
                  {t('iamReadOnlyMode')}
                </Typography>
              ) : null}
              {inviteLinkCopied ? (
                <Typography variant="caption" color="text.secondary">
                  {t('inviteCopied', { value: inviteLinkCopied })}
                </Typography>
              ) : null}
              {iamInvites.map((invite) => (
                <Paper key={invite.id} variant="outlined" sx={{ p: 1 }}>
                  <Typography variant="subtitle2">
                    {invite.inviteeEmail || invite.inviteePhone || invite.id}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {invite.status} | {t('inviteCreatedAt', { value: invite.createdAt || '-' })} | {t('inviteExpiresAt', { value: invite.expiresAt || '-' })}
                  </Typography>
                  <Stack direction="row" spacing={1} sx={{ mt: 1 }} alignItems="center">
                    <Typography variant="body2" sx={{ wordBreak: 'break-all', flex: 1 }}>
                      {invite.inviteLink}
                    </Typography>
                    <Button size="small" variant="outlined" onClick={() => void copyInviteLink(invite.inviteLink)}>
                      {t('copyInviteLink')}
                    </Button>
                    <Button
                      size="small"
                      color="warning"
                      variant="outlined"
                      onClick={() => void revokeInvite(invite)}
                      disabled={invite.status !== 'pending' || !authCanWrite}
                    >
                      {t('revokeInvite')}
                    </Button>
                  </Stack>
                </Paper>
              ))}
              {iamInvites.length === 0 && !iamLoading ? (
                <Typography color="text.secondary">{t('noEntries')}</Typography>
              ) : null}
            </Stack>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setIamDialogOpen(false)}>{t('close')}</Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={acceptInviteOpen}
        onClose={() => {
          setCurrentRoute('/');
          if (typeof window !== 'undefined') {
            window.location.hash = '/';
          }
        }}
        fullWidth
        maxWidth="xs"
      >
        <DialogTitle>{t('acceptInvite')}</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={1.2} sx={{ pt: 0.4 }}>
            <Typography variant="caption" color="text.secondary">
              {t('inviteTokenDetected')}
            </Typography>
            <TextField
              label={t('username')}
              value={inviteUsername}
              onChange={(event) => setInviteUsername(event.target.value)}
            />
            <TextField
              label={t('password')}
              type="password"
              value={invitePassword}
              onChange={(event) => setInvitePassword(event.target.value)}
            />
            {inviteStatus ? (
              <Typography variant="body2" color="success.main">
                {t('inviteAccepted', { status: inviteStatus })}
              </Typography>
            ) : null}
            {inviteError ? (
              <Typography variant="body2" color="error">
                {t('inviteAcceptError', { error: inviteError })}
              </Typography>
            ) : null}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={() => {
              setCurrentRoute('/');
              if (typeof window !== 'undefined') {
                window.location.hash = '/';
              }
            }}
          >
            {t('close')}
          </Button>
          <Button
            variant="contained"
            onClick={() => void submitAcceptInvite(inviteToken)}
            disabled={inviteBusy || inviteUsername.trim() === '' || invitePassword.trim() === ''}
          >
            {t('acceptInvite')}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={auditDialogOpen}
        onClose={() => setAuditDialogOpen(false)}
        fullWidth
        maxWidth="md"
      >
        <DialogTitle>{t('auditEvents')}</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={1.2}>
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
              <TextField
                size="small"
                label={t('auditFilterAction')}
                value={auditActionFilter}
                onChange={(event) => setAuditActionFilter(event.target.value)}
              />
              <TextField
                size="small"
                label={t('auditFilterResult')}
                value={auditResultFilter}
                onChange={(event) => setAuditResultFilter(event.target.value)}
                placeholder={t('auditFilterResultPlaceholder')}
              />
              <TextField
                size="small"
                type="number"
                label={t('auditFilterLimit')}
                value={auditLimit}
                onChange={(event) => setAuditLimit(event.target.value)}
              />
              <Button variant="contained" onClick={() => void loadAuditData()}>
                {t('refresh')}
              </Button>
            </Stack>
            {auditLoading ? <Typography>{t('loading')}</Typography> : null}
            {auditError ? (
              <Typography color="error">{t('auditLoadError', { error: auditError })}</Typography>
            ) : null}
            <Stack spacing={1}>
              {auditEvents.map((event, index) => (
                <Paper key={`${event.createdAt}-${event.action}-${index}`} variant="outlined" sx={{ p: 1 }}>
                  <Stack
                    direction={{ xs: 'column', sm: 'row' }}
                    spacing={1}
                    alignItems={{ sm: 'center' }}
                  >
                    <Typography variant="body2" sx={{ fontWeight: 700 }}>
                      {event.action}
                    </Typography>
                    <Chip size="small" label={event.result} color={event.result === 'allowed' ? 'success' : 'error'} />
                    <Typography variant="caption" color="text.secondary">
                      {event.createdAt}
                    </Typography>
                  </Stack>
                  <Typography variant="caption" color="text.secondary">
                    {event.targetType}/{event.targetID} | actor: {event.actorID || '-'}
                  </Typography>
                  {event.reason ? (
                    <Typography variant="caption" color="error">
                      {event.reason}
                    </Typography>
                  ) : null}
                </Paper>
              ))}
              {auditEvents.length === 0 && !auditLoading ? (
                <Typography color="text.secondary">{t('noEntries')}</Typography>
              ) : null}
            </Stack>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setAuditDialogOpen(false)}>{t('close')}</Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={authDialogOpen}
        onClose={() => setAuthDialogOpen(false)}
        fullWidth
        maxWidth="xs"
      >
        <DialogTitle>{t('login')}</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={1.2} sx={{ pt: 0.4 }}>
            <TextField
              label={t('username')}
              value={loginUsername}
              onChange={(event) => setLoginUsername(event.target.value)}
              autoComplete="username"
            />
            <TextField
              label={t('password')}
              type="password"
              value={loginPassword}
              onChange={(event) => setLoginPassword(event.target.value)}
              autoComplete="current-password"
            />
            <TextField
              label={t('tenantCode')}
              value={loginTenantCode}
              onChange={(event) => setLoginTenantCode(event.target.value)}
            />
            {authError ? (
              <Typography variant="body2" color="error">
                {t('authError', { error: authError })}
              </Typography>
            ) : null}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setAuthDialogOpen(false)}>{t('close')}</Button>
          <Button variant="outlined" onClick={() => void submitOAuthDemoLogin()} disabled={authBusy}>
            {t('oauthDemoLogin')}
          </Button>
          <Button variant="contained" onClick={() => void submitLogin()} disabled={authBusy}>
            {t('login')}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={manageMenusOpen}
        onClose={() => setManageMenusOpen(false)}
        fullWidth
        maxWidth="md"
      >
        <DialogTitle>{t('menuManagement')}</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={1}>
            {manageableMenus.map((menu) => (
              <Paper
                key={menu.id}
                variant="outlined"
                sx={{
                  p: 1,
                  display: 'grid',
                  gridTemplateColumns: { xs: '1fr', md: 'minmax(0,1fr) 120px 100px 150px 100px' },
                  gap: 1,
                  alignItems: 'center',
                }}
              >
                <Stack direction="row" spacing={1} alignItems="center">
                  {menuIcon(menu.group, menu.targetType)}
                  <Typography variant="body2">{menu.title}</Typography>
                </Stack>
                <Typography variant="caption" color="text.secondary">
                  {menu.source === 'system' ? t('menuSourceSystem') : t('menuSourceUser')}
                </Typography>
                <Stack direction="row" spacing={0.5} alignItems="center">
                  <Switch
                    size="small"
                    checked={menu.visible}
                    onChange={(event) =>
                      updateMenuOverride(menu, { visible: event.target.checked })
                    }
                    inputProps={{ 'aria-label': `${t('menuVisible')}-${menu.id}` }}
                  />
                  <Typography variant="caption">{t('menuVisible')}</Typography>
                </Stack>
                <TextField
                  size="small"
                  label={t('menuGroup')}
                  value={menu.group}
                  onChange={(event) =>
                    updateMenuOverride(menu, { group: event.target.value })
                  }
                />
                <TextField
                  size="small"
                  type="number"
                  label={t('menuOrder')}
                  value={menu.order}
                  onChange={(event) =>
                    updateMenuOverride(menu, { order: Number(event.target.value) || 0 })
                  }
                />
              </Paper>
            ))}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setMenuOverrides({})}>{t('resetMenuSettings')}</Button>
          <Button onClick={() => setManageMenusOpen(false)}>{t('close')}</Button>
        </DialogActions>
      </Dialog>

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
