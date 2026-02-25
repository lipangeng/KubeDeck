import type {
  ExtensionContribution,
  FrontendPlugin,
  MenuItem,
  PageContribution,
} from '../sdk/types';

export class PluginHost {
  private readonly plugins = new Map<string, FrontendPlugin>();

  register(plugin: FrontendPlugin): void {
    if (!plugin || !plugin.pluginId) {
      throw new Error('pluginId is required');
    }

    if (this.plugins.has(plugin.pluginId)) {
      throw new Error(`plugin already registered: ${plugin.pluginId}`);
    }

    this.plugins.set(plugin.pluginId, plugin);
  }

  getPages(): PageContribution[] {
    return Array.from(this.plugins.values()).flatMap(
      (plugin) => plugin.registerPages?.() ?? [],
    );
  }

  getExtensions(): ExtensionContribution[] {
    return Array.from(this.plugins.values()).flatMap(
      (plugin) => plugin.registerExtensions?.() ?? [],
    );
  }

  getMenus(): MenuItem[] {
    return Array.from(this.plugins.values()).flatMap(
      (plugin) => plugin.registerMenus?.() ?? [],
    );
  }
}
