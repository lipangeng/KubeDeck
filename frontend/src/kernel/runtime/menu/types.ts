import type { MenuContribution } from '../../contracts/menuContribution';
import type { LocalizedText } from '../../contracts/types';

export interface KernelNavigationGroup {
  key: string;
  order: number;
  title: LocalizedText;
  entries: MenuContribution[];
}
