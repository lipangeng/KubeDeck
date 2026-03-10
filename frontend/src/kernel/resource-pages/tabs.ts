import type { ResolveDefaultTabsOptions, ResourcePageTab } from './types';

export function resolveDefaultTabs(options: ResolveDefaultTabsOptions = {}): ResourcePageTab[] {
  return [
    {
      id: 'overview',
      title: 'Overview',
      capabilityType: 'tab',
      content: options.overviewContent ?? null,
    },
    {
      id: 'yaml',
      title: 'YAML',
      capabilityType: 'tab',
      content: options.yamlContent ?? null,
    },
  ];
}
