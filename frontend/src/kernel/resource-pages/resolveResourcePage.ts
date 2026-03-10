import { resolveDefaultTabs } from './tabs';
import type { ResolveDefaultTabsOptions, ResourcePageTab } from './types';

export function resolveResourcePage(options: ResolveDefaultTabsOptions = {}): ResourcePageTab[] {
  return resolveDefaultTabs(options);
}
