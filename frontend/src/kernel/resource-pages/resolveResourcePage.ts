import { resolveDefaultTabs } from './tabs';
import type {
  ResolveDefaultTabsOptions,
  ResolvedResourcePage,
  ResourcePageExtension,
  ResourcePageSummarySlotExtension,
  ResourceTabExtension,
} from './types';

export function resolveResourcePage(options: ResolveDefaultTabsOptions = {}): ResolvedResourcePage {
  let tabs = resolveDefaultTabs(options);
  const extensions = options.extensions ?? [];
  const takeovers = extensions.filter(
    (extension): extension is Extract<ResourcePageExtension, { capabilityType: 'page-takeover' }> =>
      extension.kind === options.resource?.kind && extension.capabilityType === 'page-takeover',
  );
  const takeover = takeovers.reduce<Extract<ResourcePageExtension, { capabilityType: 'page-takeover' }> | null>(
    (currentBest, candidate) => {
      if (!currentBest) {
        return candidate;
      }

      const bestPriority = currentBest.priority ?? 0;
      const candidatePriority = candidate.priority ?? 0;
      if (candidatePriority > bestPriority) {
        return candidate;
      }
      if (candidatePriority < bestPriority) {
        return currentBest;
      }

      const bestSourceRank = currentBest.origin === 'remote' ? 0 : 1;
      const candidateSourceRank = candidate.origin === 'remote' ? 0 : 1;
      if (candidateSourceRank > bestSourceRank) {
        return candidate;
      }
      if (candidateSourceRank < bestSourceRank) {
        return currentBest;
      }

      return candidate;
    },
    null,
  );
  if (takeover) {
    return {
      tabs: [],
      takeoverContent: takeover.renderPage(options),
      summaryContent: extensions
        .filter(
          (extension): extension is ResourcePageSummarySlotExtension =>
            extension.kind === options.resource?.kind &&
            extension.capabilityType === 'slot' &&
            extension.placement === 'summary',
        )
        .map((extension) => extension.renderSlot(options)),
    };
  }
  const matchingExtensions = extensions.filter((extension) => extension.kind === options.resource?.kind);
  const tabExtensions = matchingExtensions.filter(
    (extension): extension is ResourceTabExtension =>
      extension.capabilityType !== 'page-takeover' && extension.capabilityType !== 'slot',
  );
  const summaryExtensions = matchingExtensions.filter(
    (extension): extension is ResourcePageSummarySlotExtension =>
      extension.capabilityType === 'slot' && extension.placement === 'summary',
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
    summaryContent: summaryExtensions.map((extension) => extension.renderSlot(options)),
  };
}
