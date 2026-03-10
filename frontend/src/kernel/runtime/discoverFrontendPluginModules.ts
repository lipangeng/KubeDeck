import type { FrontendCapabilityModule } from '../sdk';

type RawDiscoveredModule = {
  default?: FrontendCapabilityModule;
};

export function collectFrontendPluginModules(
  discoveredModules: Record<string, RawDiscoveredModule>,
): FrontendCapabilityModule[] {
  return Object.entries(discoveredModules)
    .filter(([path]) => !path.includes('/templates/'))
    .map(([, module]) => module.default)
    .filter((module): module is FrontendCapabilityModule => Boolean(module));
}

export function discoverFrontendPluginModules(): FrontendCapabilityModule[] {
  const discoveredModules = import.meta.glob('../../../../plugins/**/src/index.ts', {
    eager: true,
  }) as Record<string, RawDiscoveredModule>;
  return collectFrontendPluginModules(discoveredModules);
}
