import { resolveDefaultTabs } from './tabs';
import type { ResolveDefaultTabsOptions, ResourcePageTab, ResourceTabExtension } from './types';

const builtInTabExtensions: ResourceTabExtension[] = [
  {
    kind: 'Deployment',
    createTab: (options) => ({
      id: 'runtime',
      title: 'Runtime',
      capabilityType: 'tab',
      content: options.runtimeContent ?? null,
    }),
  },
  {
    kind: 'Pod',
    createTab: (options) => ({
      id: 'logs',
      title: 'Logs',
      capabilityType: 'tab',
      content: options.logsContent ?? null,
    }),
  },
];

export function resolveResourcePage(options: ResolveDefaultTabsOptions = {}): ResourcePageTab[] {
  const tabs = resolveDefaultTabs(options);
  const extensions = [...builtInTabExtensions, ...(options.extensions ?? [])];
  const matchingExtensions = extensions.filter((extension) => extension.kind === options.resource?.kind);
  tabs.push(...matchingExtensions.map((extension) => extension.createTab(options)));

  return tabs;
}
