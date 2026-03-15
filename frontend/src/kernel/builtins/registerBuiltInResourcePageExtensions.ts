import type { ResourceTabExtension } from '../resource-pages/types';

export function registerBuiltInResourcePageExtensions(): ResourceTabExtension[] {
  return [
    {
      kind: 'Deployment',
      capabilityType: 'tab-replace',
      targetTabId: 'yaml',
      tabId: 'yaml',
      createTab: (options) => ({
        id: 'yaml',
        title: 'YAML v2',
        capabilityType: 'tab-replace',
        content: options.yamlVariantContent ?? null,
      }),
    },
    {
      kind: 'Deployment',
      capabilityType: 'tab',
      tabId: 'runtime',
      createTab: (options) => ({
        id: 'runtime',
        title: 'Runtime',
        capabilityType: 'tab',
        content: options.runtimeContent ?? null,
      }),
    },
    {
      kind: 'Pod',
      capabilityType: 'tab-replace',
      targetTabId: 'overview',
      tabId: 'overview',
      createTab: (options) => ({
        id: 'overview',
        title: 'Overview',
        capabilityType: 'tab-replace',
        content: `Pod-specific overview for ${options.resource?.name ?? 'pod'}`,
      }),
    },
    {
      kind: 'Pod',
      capabilityType: 'tab',
      tabId: 'logs',
      createTab: (options) => ({
        id: 'logs',
        title: 'Logs',
        capabilityType: 'tab',
        content: options.logsContent ?? null,
      }),
    },
  ];
}
