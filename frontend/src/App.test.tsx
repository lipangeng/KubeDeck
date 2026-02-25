import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import App from './App';

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe('App', () => {
  it('renders grouped menus, theme control, runtime checks, and reloads menus when cluster changes', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(
          JSON.stringify({
            clusters: ['default', 'dev', 'staging', 'prod'],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.includes('/api/meta/registry?cluster=dev')) {
        return new Response(
          JSON.stringify({
            cluster: 'dev',
            resourceTypes: [
              {
                id: 'apps.v1.deployments',
                group: 'apps',
                version: 'v1',
                kind: 'Deployment',
                plural: 'deployments',
                namespaced: true,
                preferredVersion: 'v1',
                source: 'system',
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.includes('/api/meta/registry?cluster=default')) {
        return new Response(
          JSON.stringify({
            cluster: 'default',
            resourceTypes: [
              {
                id: 'apps.v1.deployments',
                group: 'apps',
                version: 'v1',
                kind: 'Deployment',
                plural: 'deployments',
                namespaced: true,
                preferredVersion: 'v1',
                source: 'system',
              },
              {
                id: 'v1.services',
                group: '',
                version: 'v1',
                kind: 'Service',
                plural: 'services',
                namespaced: true,
                preferredVersion: 'v1',
                source: 'system',
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.includes('cluster=dev')) {
        return new Response(
          JSON.stringify({
            cluster: 'dev',
            menus: [
              {
                id: 'dev-workloads',
                group: 'system',
                title: 'Dev Workloads',
                targetType: 'page',
                targetRef: '/workloads',
                source: 'system',
                order: 10,
                visible: true,
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }

      return new Response(
        JSON.stringify({
          cluster: 'default',
          menus: [
            {
              id: 'workloads',
              group: 'system',
              title: 'Workloads',
              targetType: 'page',
              targetRef: '/workloads',
              source: 'system',
              order: 10,
              visible: true,
            },
            {
              id: 'favorites',
              group: 'user',
              title: 'Favorites',
              targetType: 'page',
              targetRef: '/favorites',
              source: 'user',
              order: 20,
              visible: true,
            },
            {
              id: 'crd-dynamic',
              group: 'dynamic',
              title: 'Custom Resources',
              targetType: 'resource',
              targetRef: '/resources/custom',
              source: 'dynamic',
              order: 30,
              visible: true,
            },
          ],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    });

    const onThemePreferenceChange = vi.fn();
    vi.stubGlobal('fetch', fetchMock);

    render(
      <App
        themePreference="system"
        onThemePreferenceChange={onThemePreferenceChange}
      />,
    );

    expect(screen.getByRole('heading', { level: 1, name: 'KubeDeck' })).toBeTruthy();
    expect(screen.getByRole('navigation', { name: 'Primary Sidebar' })).toBeTruthy();
    expect(screen.getByLabelText('Theme')).toBeTruthy();
    expect(screen.getByText('System Menus')).toBeTruthy();
    expect(screen.getByText('User Menus')).toBeTruthy();
    expect(screen.getByText('Dynamic Menus')).toBeTruthy();
    expect(screen.getByText('API target (test: http://127.0.0.1:8080)')).toBeTruthy();
    expect(await screen.findByText('healthz: ok')).toBeTruthy();
    expect(await screen.findByText('readyz: ok')).toBeTruthy();
    expect(await screen.findByText('Registry resource types: 2')).toBeTruthy();
    expect(await screen.findByText(/Last checked:/)).toBeTruthy();
    expect(await screen.findByText('Failure summary: none')).toBeTruthy();

    expect(await screen.findByText('Workloads')).toBeTruthy();
    expect(await screen.findByText('Favorites')).toBeTruthy();
    expect(await screen.findByText('Custom Resources')).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Cluster'), {
      target: { value: 'dev' },
    });
    fireEvent.change(screen.getByLabelText('Theme'), {
      target: { value: 'dark' },
    });

    expect(onThemePreferenceChange).toHaveBeenCalledWith('dark');
    expect(await screen.findByText('Dev Workloads')).toBeTruthy();
    expect(await screen.findByText('Registry resource types: 1')).toBeTruthy();
    await waitFor(() => {
      const calledUrls = fetchMock.mock.calls.map(([input]) => String(input));
      expect(calledUrls.some((url) => url.includes('cluster=default'))).toBe(true);
      expect(calledUrls.some((url) => url.includes('cluster=dev'))).toBe(true);
      expect(calledUrls.some((url) => url.endsWith('/api/meta/clusters'))).toBe(true);
      expect(calledUrls.some((url) => url.includes('/api/meta/registry?cluster=default'))).toBe(true);
      expect(calledUrls.some((url) => url.includes('/api/meta/registry?cluster=dev'))).toBe(true);
      expect(calledUrls.some((url) => url.endsWith('/api/healthz'))).toBe(true);
      expect(calledUrls.some((url) => url.endsWith('/api/readyz'))).toBe(true);
    });
  });

  it('shows failure summary when readiness check fails', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith('/api/healthz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/readyz')) {
        return new Response('not ready', { status: 503 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(
          JSON.stringify({
            clusters: ['default', 'dev'],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.includes('/api/meta/registry?cluster=default')) {
        return new Response(
          JSON.stringify({
            cluster: 'default',
            resourceTypes: [],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      return new Response(
        JSON.stringify({
          cluster: 'default',
          menus: [
            {
              id: 'workloads',
              group: 'system',
              title: 'Workloads',
              targetType: 'page',
              targetRef: '/workloads',
              source: 'system',
              order: 10,
              visible: true,
            },
          ],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    });

    vi.stubGlobal('fetch', fetchMock);

    render(
      <App
        themePreference="system"
        onThemePreferenceChange={vi.fn()}
      />,
    );

    expect(await screen.findByText('healthz: ok')).toBeTruthy();
    expect(await screen.findByText('readyz: error')).toBeTruthy();
    expect(await screen.findByText('Failure summary: readyz: status 503')).toBeTruthy();
  });
});
