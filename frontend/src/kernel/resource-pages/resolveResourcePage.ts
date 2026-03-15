import { resolveDefaultTabs } from './tabs';
import type { ResolveDefaultTabsOptions, ResourcePageTab, ResourceTabExtension } from './types';

const builtInTabExtensions: ResourceTabExtension[] = [
  {
    kind: 'Deployment',
    capabilityType: 'tab',
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
    createTab: (options) => ({
      id: 'logs',
      title: 'Logs',
      capabilityType: 'tab',
      content: options.logsContent ?? null,
    }),
  },
];

export function resolveResourcePage(options: ResolveDefaultTabsOptions = {}): ResourcePageTab[] {
  let tabs = resolveDefaultTabs(options);
  const extensions = [...builtInTabExtensions, ...(options.extensions ?? [])];
  const matchingExtensions = extensions.filter((extension) => extension.kind === options.resource?.kind);
  const replacements = matchingExtensions.filter((extension) => extension.capabilityType === 'tab-replace');
  for (const replacement of replacements) {
    const replacementTab = replacement.createTab(options);
    tabs = tabs.map((tab) => (tab.id === replacement.targetTabId ? replacementTab : tab));
  }
  const appendedTabs = matchingExtensions
    .filter((extension) => extension.capabilityType !== 'tab-replace')
    .map((extension) => extension.createTab(options));
  tabs.push(...appendedTabs);

  return tabs;
}
