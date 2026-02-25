import type { FrontendPlugin } from '../../../../frontend/src/sdk/types';

const plugin: FrontendPlugin = {
  pluginId: 'example-frontend-plugin',
  registerPages: () => [],
  registerExtensions: () => [],
  registerMenus: () => [],
};

export default plugin;
