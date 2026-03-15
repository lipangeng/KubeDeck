import { resolveDefaultTabs } from './tabs';
import type {
  ResolveDefaultTabsOptions,
  ResolvedResourcePage,
  ResourcePageExtension,
  ResourceTabExtension,
} from './types';

export function resolveResourcePage(options: ResolveDefaultTabsOptions = {}): ResolvedResourcePage {
  let tabs = resolveDefaultTabs(options);
  const extensions = options.extensions ?? [];
  const takeover = extensions.find(
    (extension): extension is Extract<ResourcePageExtension, { capabilityType: 'page-takeover' }> =>
      extension.kind === options.resource?.kind && extension.capabilityType === 'page-takeover',
  );
  if (takeover) {
    return {
      tabs: [],
      takeoverContent: takeover.renderPage(options),
    };
  }
  const matchingExtensions = extensions.filter((extension) => extension.kind === options.resource?.kind);
  const tabExtensions = matchingExtensions.filter(
    (extension): extension is ResourceTabExtension => extension.capabilityType !== 'page-takeover',
  );
  const replacements = tabExtensions.filter((extension) => extension.capabilityType === 'tab-replace');
  for (const replacement of replacements) {
    const replacementTab = replacement.createTab(options);
    tabs = tabs.map((tab) => (tab.id === replacement.targetTabId ? replacementTab : tab));
  }
  const appendedTabs = tabExtensions
    .filter((extension) => extension.capabilityType !== 'tab-replace')
    .map((extension) => extension.createTab(options));
  tabs.push(...appendedTabs);

  return {
    tabs,
    takeoverContent: null,
  };
}
