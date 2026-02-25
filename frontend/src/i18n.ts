export type Locale = 'en' | 'zh';

export const LOCALE_KEY = 'kubedeck.locale';

const messages = {
  en: {
    runtime: 'Runtime',
    checked: 'Updated {{time}}',
    theme: 'Theme',
    language: 'Language',
    manageMenus: 'Manage Menus',
    accessControl: 'Access Control',
    menuManagement: 'Menu Management',
    menuSource: 'Source',
    menuSourceSystem: 'System',
    menuSourceUser: 'User',
    menuVisible: 'Visible',
    menuGroup: 'Group',
    menuOrder: 'Order',
    resetMenuSettings: 'Reset Menu Settings',
    system: 'System',
    light: 'Light',
    dark: 'Dark',
    login: 'Login',
    logout: 'Logout',
    username: 'Username',
    password: 'Password',
    tenantCode: 'Tenant Code',
    authError: 'Authentication failed: {{error}}',
    acceptInvite: 'Accept Invite',
    inviteTokenDetected: 'Invite token detected from URL hash.',
    inviteAccepted: 'Invite accepted: {{status}}',
    inviteAcceptError: 'Invite acceptance failed: {{error}}',
    chinese: '中文',
    english: 'English',
    navigation: 'Navigation',
    favorites: 'Favorites',
    favoriteAdd: 'Add to favorites',
    favoriteRemove: 'Remove from favorites',
    fixedFavorite: 'Fixed favorite',
    menu: 'Menu',
    context: 'Context',
    cluster: 'Cluster',
    tenant: 'Tenant',
    namespaceFilter: 'Namespace Filter',
    loadingMenus: 'Loading menus...',
    failedMenus: 'Failed to load menus: {{error}}',
    iamLoadError: 'IAM operation failed: {{error}}',
    noMenus: 'No menus.',
    noEntries: 'No entries',
    loading: 'Loading...',
    createGroup: 'Create Group',
    groupName: 'Group Name',
    groupDescription: 'Group Description',
    groupPermissions: 'Group Permissions (comma separated)',
    savePermissions: 'Save Permissions',
    refresh: 'Refresh',
    iamPermissionsCatalog: 'Permissions Catalog',
    namespaced: 'Namespaced',
    clusterScoped: 'Cluster',
    controlPlane: 'Kubernetes Dashboard',
    clusterOverview: 'Cluster {{cluster}} Overview',
    apiTarget: 'API target ({{target}})',
    health: 'Health',
    ready: 'Ready',
    registryResourceTypes: 'Registry Resource Types',
    namespacedTypes: 'Namespaced Types',
    clusterScopedTypes: 'Cluster-Scoped Types',
    resourceCatalogGroups: 'Resource Catalog Groups',
    failureSummary: 'Failure summary',
    failureSummaryText: 'Failure summary: {{summary}}',
    createResources: 'Create Resources',
    createNamespace: 'Create Namespace',
    yamlDocuments: 'YAML documents: {{count}}',
    yamlLabel: 'YAML (multi-document supported)',
    applyStatus: 'Apply status: {{status}}',
    applyYaml: 'Apply YAML',
    close: 'Close',
    succeeded: 'succeeded',
    failed: 'failed',
    unknown: 'Unknown',
  },
  zh: {
    runtime: '运行状态',
    checked: '更新时间 {{time}}',
    theme: '主题',
    language: '语言',
    manageMenus: '菜单管理',
    accessControl: '访问控制',
    menuManagement: '菜单管理',
    menuSource: '来源',
    menuSourceSystem: '系统',
    menuSourceUser: '用户',
    menuVisible: '可见',
    menuGroup: '分组',
    menuOrder: '排序',
    resetMenuSettings: '重置菜单设置',
    system: '跟随系统',
    light: '浅色',
    dark: '深色',
    login: '登录',
    logout: '退出登录',
    username: '用户名',
    password: '密码',
    tenantCode: '租户编码',
    authError: '认证失败：{{error}}',
    acceptInvite: '接受邀请',
    inviteTokenDetected: '已从 URL Hash 检测到邀请令牌。',
    inviteAccepted: '邀请已接受：{{status}}',
    inviteAcceptError: '接受邀请失败：{{error}}',
    chinese: '中文',
    english: 'English',
    navigation: '导航',
    favorites: '收藏',
    favoriteAdd: '加入收藏',
    favoriteRemove: '取消收藏',
    fixedFavorite: '固定收藏',
    menu: '菜单',
    context: '上下文',
    cluster: '集群',
    tenant: '租户',
    namespaceFilter: '命名空间过滤',
    loadingMenus: '正在加载菜单...',
    failedMenus: '菜单加载失败：{{error}}',
    iamLoadError: 'IAM 操作失败：{{error}}',
    noMenus: '暂无菜单。',
    noEntries: '无条目',
    loading: '加载中...',
    createGroup: '创建分组',
    groupName: '分组名称',
    groupDescription: '分组描述',
    groupPermissions: '分组权限（逗号分隔）',
    savePermissions: '保存权限',
    refresh: '刷新',
    iamPermissionsCatalog: '权限目录',
    namespaced: '命名空间级',
    clusterScoped: '集群级',
    controlPlane: 'Kubernetes 概览',
    clusterOverview: '集群 {{cluster}} 总览',
    apiTarget: 'API 目标（{{target}}）',
    health: '健康',
    ready: '就绪',
    registryResourceTypes: '注册资源类型数',
    namespacedTypes: '命名空间级类型',
    clusterScopedTypes: '集群级类型',
    resourceCatalogGroups: '资源目录分组',
    failureSummary: '失败摘要',
    failureSummaryText: '失败摘要：{{summary}}',
    createResources: '创建资源',
    createNamespace: '创建命名空间',
    yamlDocuments: 'YAML 文档数：{{count}}',
    yamlLabel: 'YAML（支持多文档）',
    applyStatus: '应用状态：{{status}}',
    applyYaml: '应用 YAML',
    close: '关闭',
    succeeded: '成功',
    failed: '失败',
    unknown: '未知',
  },
} as const;

type MessageKey = keyof (typeof messages)['en'];

export function translate(locale: Locale, key: MessageKey, vars?: Record<string, string>): string {
  const template: string = messages[locale][key];
  if (!vars) {
    return template;
  }
  return Object.entries(vars).reduce((acc, [name, value]) => {
    const token = `{{${name}}}`;
    return acc.split(token).join(value);
  }, template);
}

export function sanitizeLocale(value: string | null | undefined): Locale {
  if (value === 'zh' || value === 'en') {
    return value;
  }
  return 'en';
}

export function detectBrowserLocale(): Locale {
  if (typeof navigator === 'undefined') {
    return 'en';
  }
  const lang = navigator.language.toLowerCase();
  return lang.startsWith('zh') ? 'zh' : 'en';
}

export function readStoredLocale(): Locale {
  if (typeof window === 'undefined') {
    return 'en';
  }
  const stored = window.localStorage.getItem(LOCALE_KEY);
  if (stored === null) {
    return detectBrowserLocale();
  }
  return sanitizeLocale(stored);
}

export function writeStoredLocale(locale: Locale): void {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.setItem(LOCALE_KEY, locale);
}
