import type { FrontendCapabilityModule } from '../sdk';
import { registerBuiltInActions } from '../builtins/registerBuiltInActions';
import { registerBuiltInMenus } from '../builtins/registerBuiltInMenus';
import { registerBuiltInPages } from '../builtins/registerBuiltInPages';
import { registerBuiltInResourcePageExtensions } from '../builtins/registerBuiltInResourcePageExtensions';
import { registerBuiltInSlots } from '../builtins/registerBuiltInSlots';
import { composeKernelNavigation } from './composeKernelNavigation';
import { KernelRegistry } from './kernelRegistry';
import type { KernelRegistrySnapshot } from './types';

export function createLocalKernelSnapshot(
  pluginModules: FrontendCapabilityModule[] = [],
): KernelRegistrySnapshot {
  const registry = new KernelRegistry();
  registry.register({
    pages: registerBuiltInPages(),
    menus: registerBuiltInMenus(),
    actions: registerBuiltInActions(),
    slots: registerBuiltInSlots(),
  });
  const resourcePageExtensions = [...registerBuiltInResourcePageExtensions()];
  for (const pluginModule of pluginModules) {
    registry.register({
      pages: pluginModule.registerPages?.() ?? [],
      menus: pluginModule.registerMenus?.() ?? [],
      actions: pluginModule.registerActions?.() ?? [],
      slots: pluginModule.registerSlots?.() ?? [],
    });
    resourcePageExtensions.push(...(pluginModule.registerResourcePageExtensions?.() ?? []));
  }
  const snapshot = registry.snapshot();
  return {
    ...snapshot,
    menuGroups: composeKernelNavigation(snapshot.menus),
    resourcePageExtensions,
  };
}
