import { describe, expect, it } from 'vitest';
import { registerBuiltInResourcePageExtensions } from '../builtins/registerBuiltInResourcePageExtensions';
import { resolveResourcePage } from './resolveResourcePage';
import { resolveDefaultTabs } from './tabs';

describe('resource page tabs', () => {
  it('resolves overview and yaml as the default tabs', () => {
    expect(resolveDefaultTabs().map((tab) => tab.id)).toEqual(['overview', 'yaml']);
  });

  it('adds a runtime tab for deployment resources without replacing overview and yaml', () => {
    const tabs = resolveResourcePage({
      resource: {
        kind: 'Deployment',
        name: 'api',
        namespace: 'default',
      },
      extensions: registerBuiltInResourcePageExtensions(),
    });

    expect(tabs.map((tab) => tab.id)).toEqual(['overview', 'yaml', 'runtime']);
  });

  it('applies registered tab extensions after the default tabs', () => {
    const tabs = resolveResourcePage({
      resource: {
        kind: 'Service',
        name: 'api',
        namespace: 'default',
      },
      extensions: [
        {
          kind: 'Service',
          createTab: () => ({
            id: 'endpoints',
            title: 'Endpoints',
            capabilityType: 'tab',
            content: null,
          }),
        },
      ],
    });

    expect(tabs.map((tab) => tab.id)).toEqual(['overview', 'yaml', 'endpoints']);
  });

  it('adds a logs tab for pod resources without replacing overview and yaml', () => {
    const tabs = resolveResourcePage({
      resource: {
        kind: 'Pod',
        name: 'api-7c9d8',
        namespace: 'default',
      },
      extensions: registerBuiltInResourcePageExtensions(),
    });

    expect(tabs.map((tab) => tab.id)).toEqual(['overview', 'yaml', 'logs']);
  });

  it('replaces a default tab when a matching replacement is registered', () => {
    const tabs = resolveResourcePage({
      resource: {
        kind: 'Pod',
        name: 'api-7c9d8',
        namespace: 'default',
      },
      overviewContent: 'Generic overview',
      extensions: [
        ...registerBuiltInResourcePageExtensions(),
        {
          kind: 'Pod',
          targetTabId: 'overview',
          capabilityType: 'tab-replace',
          createTab: () => ({
            id: 'overview',
            title: 'Overview',
            capabilityType: 'tab-replace',
            content: 'Pod overview replacement',
          }),
        },
      ],
    });

    expect(tabs.map((tab) => tab.id)).toEqual(['overview', 'yaml', 'logs']);
    expect(tabs[0]?.content).toBe('Pod overview replacement');
  });

  it('supports replacing the yaml tab with a resource-specific variant', () => {
    const tabs = resolveResourcePage({
      resource: {
        kind: 'Deployment',
        name: 'api',
        namespace: 'default',
      },
      yamlContent: 'Original YAML',
      yamlVariantContent: 'Deployment YAML v2 for api',
      extensions: registerBuiltInResourcePageExtensions(),
    });

    expect(tabs.map((tab) => tab.id)).toEqual(['overview', 'yaml', 'runtime']);
    expect(tabs[1]?.title).toBe('YAML v2');
    expect(tabs[1]?.content).toBe('Deployment YAML v2 for api');
  });
});
