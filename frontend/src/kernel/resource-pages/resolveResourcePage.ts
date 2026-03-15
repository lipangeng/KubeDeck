import { resolveDefaultTabs } from './tabs';
import type { ResolveDefaultTabsOptions, ResourcePageTab } from './types';

export function resolveResourcePage(options: ResolveDefaultTabsOptions = {}): ResourcePageTab[] {
  const tabs = resolveDefaultTabs(options);

  if (options.resource?.kind === 'Deployment') {
    tabs.push({
      id: 'runtime',
      title: 'Runtime',
      capabilityType: 'tab',
      content: options.runtimeContent ?? null,
    });
  }

  return tabs;
}
