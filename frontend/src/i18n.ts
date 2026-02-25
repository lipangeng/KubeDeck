export type Locale = 'en' | 'zh';

export const LOCALE_KEY = 'kubedeck.locale';

const messages = {
  en: {
    runtime: 'Runtime',
    checked: 'Updated {{time}}',
    theme: 'Theme',
    language: 'Language',
    system: 'System',
    light: 'Light',
    dark: 'Dark',
    chinese: '中文',
    english: 'English',
    navigation: 'Navigation',
    cluster: 'Cluster',
    namespaceFilter: 'Namespace Filter',
    loadingMenus: 'Loading menus...',
    failedMenus: 'Failed to load menus: {{error}}',
    configuredMenus: 'Configured Menus',
    noConfiguredMenus: 'No configured menus.',
    resourceCatalog: 'Resource Catalog (for menu config)',
    resourceCatalogDesc: 'Built-in resources + discovered CRDs from current cluster.',
    noEntries: 'No entries',
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
    system: '跟随系统',
    light: '浅色',
    dark: '深色',
    chinese: '中文',
    english: 'English',
    navigation: '导航',
    cluster: '集群',
    namespaceFilter: '命名空间过滤',
    loadingMenus: '正在加载菜单...',
    failedMenus: '菜单加载失败：{{error}}',
    configuredMenus: '已配置菜单',
    noConfiguredMenus: '暂无已配置菜单。',
    resourceCatalog: '资源目录（用于菜单配置）',
    resourceCatalogDesc: '当前集群内置资源与已发现 CRD。',
    noEntries: '无条目',
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
