import { resolveDefaultTabs } from './tabs';
import type { ResolveDefaultTabsOptions, ResourcePageTab, ResourceTabExtension } from './types';

export function resolveResourcePage(options: ResolveDefaultTabsOptions = {}): ResourcePageTab[] {
  let tabs = resolveDefaultTabs(options);
  const extensions = options.extensions ?? [];
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
