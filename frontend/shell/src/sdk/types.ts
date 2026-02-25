export type MenuSource = 'system' | 'user' | 'dynamic';

export interface MenuItem {
  id: string;
  group: string;
  title: string;
  targetType: 'page' | 'resource';
  targetRef: string;
  source: MenuSource;
  order: number;
  visible: boolean;
}

export interface PageContribution {
  pageId: string;
  route: string;
  replacementFor?: string;
}

export interface ExtensionContribution {
  slotId: string;
  accepts: 'tab' | 'panel' | 'action' | 'widget';
  order?: number;
}

export interface FrontendPlugin {
  pluginId: string;
  registerPages?: () => PageContribution[];
  registerExtensions?: () => ExtensionContribution[];
  registerMenus?: () => MenuItem[];
}
