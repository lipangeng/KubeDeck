import { describe, expect, it } from 'vitest';
import type { FrontendCapabilityModule } from '../sdk';
import { collectFrontendPluginModules } from './discoverFrontendPluginModules';

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
});
