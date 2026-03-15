import { describe, expect, it } from 'vitest';
import { registerBuiltInResourcePageExtensions } from '../builtins/registerBuiltInResourcePageExtensions';
import { resolveResourcePage } from './resolveResourcePage';
import { resolveDefaultTabs } from './tabs';

describe('resource page tabs', () => {
  it('resolves overview and yaml as the default tabs', () => {
    expect(resolveDefaultTabs().map((tab) => tab.id)).toEqual(['overview', 'yaml']);
  });

  it('adds a runtime tab for deployment resources without replacing overview and yaml', () => {
    const resolution = resolveResourcePage({
      resource: {
        kind: 'Deployment',
        name: 'api',
        namespace: 'default',
      },
      extensions: registerBuiltInResourcePageExtensions(),
    });

    expect(resolution.tabs.map((tab) => tab.id)).toEqual(['overview', 'yaml', 'runtime']);
    expect(resolution.takeoverContent).toBeNull();
  });

  it('applies registered tab extensions after the default tabs', () => {
    const resolution = resolveResourcePage({
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

    expect(resolution.tabs.map((tab) => tab.id)).toEqual(['overview', 'yaml', 'endpoints']);
    expect(resolution.takeoverContent).toBeNull();
  });

  it('adds a logs tab for pod resources without replacing overview and yaml', () => {
    const resolution = resolveResourcePage({
      resource: {
        kind: 'Pod',
        name: 'api-7c9d8',
        namespace: 'default',
      },
      extensions: registerBuiltInResourcePageExtensions(),
    });

    expect(resolution.tabs.map((tab) => tab.id)).toEqual(['overview', 'yaml', 'logs']);
    expect(resolution.takeoverContent).toBeNull();
  });

  it('replaces a default tab when a matching replacement is registered', () => {
    const resolution = resolveResourcePage({
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

    expect(resolution.tabs.map((tab) => tab.id)).toEqual(['overview', 'yaml', 'logs']);
    expect(resolution.tabs[0]?.content).toBe('Pod overview replacement');
  });

  it('supports replacing the yaml tab with a resource-specific variant', () => {
    const resolution = resolveResourcePage({
      resource: {
        kind: 'Deployment',
        name: 'api',
        namespace: 'default',
      },
      yamlContent: 'Original YAML',
      yamlVariantContent: 'Deployment YAML v2 for api',
      extensions: registerBuiltInResourcePageExtensions(),
    });

    expect(resolution.tabs.map((tab) => tab.id)).toEqual(['overview', 'yaml', 'runtime']);
    expect(resolution.tabs[1]?.title).toBe('YAML v2');
    expect(resolution.tabs[1]?.content).toBe('Deployment YAML v2 for api');
  });

  it('returns a takeover page for statefulset resources', () => {
    const resolution = resolveResourcePage({
      resource: {
        kind: 'StatefulSet',
        name: 'db',
        namespace: 'default',
      },
      extensions: registerBuiltInResourcePageExtensions(),
    });

    expect(resolution.tabs).toEqual([]);
    expect(resolution.takeoverContent).toBe('StatefulSet takeover for db');
  });

  it('prefers the highest-priority takeover when multiple takeovers match', () => {
    const resolution = resolveResourcePage({
      resource: {
        kind: 'StatefulSet',
        name: 'db',
        namespace: 'default',
      },
      extensions: [
        {
          kind: 'StatefulSet',
          capabilityType: 'page-takeover',
          priority: 10,
          renderPage: () => 'Lower-priority takeover',
        },
        {
          kind: 'StatefulSet',
          capabilityType: 'page-takeover',
          priority: 50,
          renderPage: () => 'Higher-priority takeover',
        },
      ],
    });

    expect(resolution.takeoverContent).toBe('Higher-priority takeover');
  });

  it('prefers the latest takeover when the source and priority are the same', () => {
    const resolution = resolveResourcePage({
      resource: {
        kind: 'StatefulSet',
        name: 'db',
        namespace: 'default',
      },
      extensions: [
        {
          kind: 'StatefulSet',
          capabilityType: 'page-takeover',
          priority: 20,
          renderPage: () => 'First takeover',
        },
        {
          kind: 'StatefulSet',
          capabilityType: 'page-takeover',
          priority: 20,
          renderPage: () => 'Latest takeover',
        },
      ],
    });

    expect(resolution.takeoverContent).toBe('Latest takeover');
  });
});
