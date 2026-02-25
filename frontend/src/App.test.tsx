import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
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
                group: 'WORKLOAD',
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
              group: 'WORKLOAD',
              title: 'Workloads',
              targetType: 'page',
              targetRef: '/workloads',
              source: 'system',
              order: 10,
              visible: true,
            },
            {
              id: 'favorites',
              group: 'FAVORITES',
              title: 'Favorites',
              targetType: 'page',
              targetRef: '/favorites',
              source: 'user',
              order: 20,
              visible: true,
            },
          ],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    });

    const onThemePreferenceChange = vi.fn();
    const onLocaleChange = vi.fn();
    vi.stubGlobal('fetch', fetchMock);

    render(
      <App
        locale="en"
        onLocaleChange={onLocaleChange}
        themePreference="system"
        onThemePreferenceChange={onThemePreferenceChange}
      />,
    );

    expect(screen.getByRole('heading', { level: 1, name: 'KubeDeck' })).toBeTruthy();
    expect(screen.getByRole('navigation', { name: 'Primary Sidebar' })).toBeTruthy();
    expect(screen.getByLabelText('Theme')).toBeTruthy();
    expect(screen.getByLabelText('Language')).toBeTruthy();
    expect(screen.getByText('Favorites')).toBeTruthy();
    expect(screen.getByText('Context')).toBeTruthy();
    expect(screen.getByText('API target (test: http://127.0.0.1:8080)')).toBeTruthy();
    expect(await screen.findByText('Runtime: ok')).toBeTruthy();
    expect(await screen.findByText('Health: ok')).toBeTruthy();
    expect(await screen.findByText('Ready: ok')).toBeTruthy();
    expect(await screen.findByText('Registry Resource Types')).toBeTruthy();
    expect(await screen.findByText('Namespaced Types')).toBeTruthy();
    expect(await screen.findByText('Cluster-Scoped Types')).toBeTruthy();
    expect(screen.getByTestId('registry-resource-type-count').textContent).toBe('2');
    expect(screen.getByTestId('namespaced-resource-type-count').textContent).toBe('2');
    expect(screen.getByTestId('cluster-scoped-resource-type-count').textContent).toBe('0');
    expect(await screen.findByText(/Updated/)).toBeTruthy();
    expect(await screen.findByText('Failure summary: none')).toBeTruthy();

    expect(await screen.findByText('Workloads')).toBeTruthy();
    expect((await screen.findAllByText('Favorites')).length).toBeGreaterThan(0);
    expect(await screen.findByText(/WORKLOAD/)).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Cluster'), {
      target: { value: 'dev' },
    });
    fireEvent.change(screen.getByLabelText('Theme'), {
      target: { value: 'dark' },
    });
    fireEvent.change(screen.getByLabelText('Language'), {
      target: { value: 'zh' },
    });

    expect(onThemePreferenceChange).toHaveBeenCalledWith('dark');
    expect(onLocaleChange).toHaveBeenCalledWith('zh');
    expect(await screen.findByText('Dev Workloads')).toBeTruthy();
    await waitFor(() => {
      expect(screen.getByTestId('registry-resource-type-count').textContent).toBe('1');
    });
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
              group: 'WORKLOAD',
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
        locale="en"
        onLocaleChange={vi.fn()}
        themePreference="system"
        onThemePreferenceChange={vi.fn()}
      />,
    );

    expect(await screen.findByText('Runtime: error')).toBeTruthy();
    expect(await screen.findByText('Health: ok')).toBeTruthy();
    expect(await screen.findByText('Ready: error')).toBeTruthy();
    expect(await screen.findByText('Failure summary: readyz: status 503')).toBeTruthy();
  });

  it('applies multi-yaml payload with namespace default linkage', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
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
      if (url.includes('/api/meta/menus?cluster=default')) {
        return new Response(
          JSON.stringify({
            cluster: 'default',
            menus: [],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.includes('/api/resources/apply?')) {
        expect(url.includes('defaultNs=dev')).toBe(true);
        expect(init?.method).toBe('POST');
        return new Response(
          JSON.stringify({
            status: 'partial',
            cluster: 'default',
            defaultNamespace: 'dev',
            total: 2,
            succeeded: 1,
            failed: 1,
            results: [
              {
                index: 1,
                kind: 'ConfigMap',
                name: 'cm-ok',
                namespace: 'dev',
                status: 'succeeded',
              },
              {
                index: 2,
                kind: 'Service',
                name: 'svc-fail',
                namespace: 'dev',
                status: 'failed',
                reason: 'simulated apply failure',
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      return new Response('not found', { status: 404 });
    });

    vi.stubGlobal('fetch', fetchMock);

    render(
      <App
        locale="en"
        onLocaleChange={vi.fn()}
        themePreference="system"
        onThemePreferenceChange={vi.fn()}
      />,
    );

    fireEvent.click(await screen.findByRole('button', { name: 'Create Resources' }));
    expect(await screen.findByRole('dialog', { name: 'Create Resources' })).toBeTruthy();
    fireEvent.change(screen.getByLabelText('Create Namespace'), {
      target: { value: 'dev' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Apply YAML' }));

    expect(await screen.findByText('Apply status: partial')).toBeTruthy();
    expect(await screen.findByText(/#1 ConfigMap cm-ok/)).toBeTruthy();
    expect(await screen.findByText(/#2 Service svc-fail/)).toBeTruthy();
  });

  it('allows toggling menu favorites and persists in sidebar favorites section', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(
          JSON.stringify({
            clusters: ['default'],
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
      if (url.includes('/api/meta/menus?cluster=default')) {
        return new Response(
          JSON.stringify({
            cluster: 'default',
            menus: [
              {
                id: 'workloads',
                group: 'WORKLOAD',
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
      }
      return new Response('not found', { status: 404 });
    });

    vi.stubGlobal('fetch', fetchMock);

    render(
      <App
        locale="en"
        onLocaleChange={vi.fn()}
        themePreference="system"
        onThemePreferenceChange={vi.fn()}
      />,
    );

    const addFavoriteButton = await screen.findByRole('button', { name: 'Add to favorites' });
    fireEvent.click(addFavoriteButton);

    expect(await screen.findByRole('button', { name: 'Remove from favorites' })).toBeTruthy();
    expect((await screen.findAllByText('Workloads')).length).toBeGreaterThan(0);
  });

  it('manages menu visibility from menu management dialog', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(
          JSON.stringify({
            clusters: ['default'],
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
      if (url.includes('/api/meta/menus?cluster=default')) {
        return new Response(
          JSON.stringify({
            cluster: 'default',
            menus: [
              {
                id: 'workloads',
                group: 'WORKLOAD',
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
      }
      return new Response('not found', { status: 404 });
    });

    vi.stubGlobal('fetch', fetchMock);

    render(
      <App
        locale="en"
        onLocaleChange={vi.fn()}
        themePreference="system"
        onThemePreferenceChange={vi.fn()}
      />,
    );

    expect(await screen.findByText('Workloads')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Manage Menus' }));
    fireEvent.click((await screen.findAllByRole('switch'))[0]);
    fireEvent.click(screen.getByRole('button', { name: 'Close' }));

    await waitFor(() => {
      expect(screen.queryByText('Workloads')).toBeNull();
    });
  });

  it('logs in and switches tenant in header', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(
          JSON.stringify({
            clusters: ['default'],
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
      if (url.includes('/api/meta/menus?cluster=default')) {
        return new Response(
          JSON.stringify({
            cluster: 'default',
            menus: [],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/auth/login')) {
        return new Response(
          JSON.stringify({
            token: 'token-abc',
            user: { id: 'u-1', username: 'admin' },
            tenants: [
              { id: 'tenant-dev', code: 'dev', name: 'Development' },
              { id: 'tenant-staging', code: 'staging', name: 'Staging' },
            ],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({
            user: { id: 'u-1', username: 'admin', activeTenantID: 'tenant-dev' },
            tenants: [
              { id: 'tenant-dev', code: 'dev', name: 'Development' },
              { id: 'tenant-staging', code: 'staging', name: 'Staging' },
            ],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/auth/switch-tenant')) {
        expect(init?.method).toBe('POST');
        expect(String(init?.body)).toContain('"tenant_code":"staging"');
        return new Response(
          JSON.stringify({
            active_tenant_id: 'tenant-staging',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      return new Response('not found', { status: 404 });
    });

    vi.stubGlobal('fetch', fetchMock);

    render(
      <App
        locale="en"
        onLocaleChange={vi.fn()}
        themePreference="system"
        onThemePreferenceChange={vi.fn()}
      />,
    );

    fireEvent.click(await screen.findByRole('button', { name: 'Login' }));
    const loginDialog = await screen.findByRole('dialog', { name: 'Login' });
    const submitButton = within(loginDialog).getByRole('button', { name: 'Login' });
    fireEvent.click(submitButton);

    expect(await screen.findByText('admin')).toBeTruthy();
    fireEvent.change(await screen.findByLabelText('Tenant'), {
      target: { value: 'staging' },
    });
    expect(await screen.findByDisplayValue('staging')).toBeTruthy();
  });
});
