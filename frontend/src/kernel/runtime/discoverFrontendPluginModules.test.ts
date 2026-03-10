import { describe, expect, it } from 'vitest';
import type { FrontendCapabilityModule } from '../sdk';
import {
  collectFrontendPluginModules,
  discoverFrontendPluginModules,
} from './discoverFrontendPluginModules';

describe('collectFrontendPluginModules', () => {
  it('returns default frontend plugin modules and skips template entries', () => {
    const pluginModule: FrontendCapabilityModule = {
      pluginId: 'plugin.ops-console',
      registerPages: () => [],
      registerMenus: () => [],
      registerActions: () => [],
      registerSlots: () => [],
    };

    const discovered = collectFrontendPluginModules({
      '../../../../plugins/templates/frontend-plugin-template/src/index.ts': {
        default: {
          pluginId: 'template-plugin',
          registerPages: () => [],
        },
      },
      '../../../../plugins/ops-console/src/index.ts': {
        default: pluginModule,
      },
    });

    expect(discovered).toEqual([pluginModule]);
  });

  it('discovers the repository sample frontend plugin and skips templates', () => {
    const discovered = discoverFrontendPluginModules();

    expect(discovered.some((module) => module.pluginId === 'plugin.sample-ops-console')).toBe(true);
    expect(discovered.some((module) => module.pluginId === 'example-frontend-plugin')).toBe(false);
  });
});
