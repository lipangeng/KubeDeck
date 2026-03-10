import { registerBuiltInActions } from '../builtins/registerBuiltInActions';
import { registerBuiltInMenus } from '../builtins/registerBuiltInMenus';
import { registerBuiltInPages } from '../builtins/registerBuiltInPages';
import { registerBuiltInSlots } from '../builtins/registerBuiltInSlots';
import { KernelRegistry } from './kernelRegistry';
import type { KernelRegistrySnapshot } from './types';

export function createLocalKernelSnapshot(): KernelRegistrySnapshot {
  const registry = new KernelRegistry();
  registry.register({
    pages: registerBuiltInPages(),
    menus: registerBuiltInMenus(),
    actions: registerBuiltInActions(),
    slots: registerBuiltInSlots(),
  });
  return registry.snapshot();
}
