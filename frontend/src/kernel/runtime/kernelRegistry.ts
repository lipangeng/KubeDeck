import type { ActionContribution } from '../contracts/actionContribution';
import type { MenuContribution } from '../contracts/menuContribution';
import type { PageContribution } from '../contracts/pageContribution';
import type { SlotContribution } from '../contracts/slotContribution';
import type { KernelRegistrySnapshot } from './types';

export interface KernelContributionModule {
  pages?: PageContribution[];
  menus?: MenuContribution[];
  actions?: ActionContribution[];
  slots?: SlotContribution[];
  resourcePageExtensions?: KernelRegistrySnapshot['resourcePageExtensions'];
}

export class KernelRegistry {
  private readonly modules: KernelContributionModule[] = [];

  register(module: KernelContributionModule): void {
    this.modules.push({
      pages: module.pages ?? [],
      menus: module.menus ?? [],
      actions: module.actions ?? [],
      slots: module.slots ?? [],
      resourcePageExtensions: module.resourcePageExtensions ?? [],
    });
  }

  snapshot(): KernelRegistrySnapshot {
    return {
      pages: this.modules.flatMap((module) => module.pages ?? []),
      menus: this.modules.flatMap((module) => module.menus ?? []),
      menuGroups: [],
      actions: this.modules.flatMap((module) => module.actions ?? []),
      slots: this.modules.flatMap((module) => module.slots ?? []),
      resourcePageExtensions: this.modules.flatMap((module) => module.resourcePageExtensions ?? []),
    };
  }
}
