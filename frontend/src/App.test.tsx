import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import App, { shouldShowOAuthConfigDiagnostics } from './App';

afterEach(() => {
  cleanup();
  window.localStorage.clear();
  window.sessionStorage.clear();
  vi.restoreAllMocks();
});

describe('App', () => {
  it('shows oauth config diagnostics only for development and test modes', () => {
    expect(shouldShowOAuthConfigDiagnostics('test')).toBe(true);
    expect(shouldShowOAuthConfigDiagnostics('development')).toBe(true);
    expect(shouldShowOAuthConfigDiagnostics('production')).toBe(false);
  });

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
    expect((await screen.findAllByText('Health: ok')).length).toBeGreaterThan(0);
    expect((await screen.findAllByText('Ready: ok')).length).toBeGreaterThan(0);
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
    expect((await screen.findAllByText('Health: ok')).length).toBeGreaterThan(0);
    expect((await screen.findAllByText('Ready: error')).length).toBeGreaterThan(0);
    expect(await screen.findByText('Failure summary: readyz: status 503')).toBeTruthy();
  });

  it('renders polished top status panel and login dialog visual shell', async () => {
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
      return new Response(
        JSON.stringify({
          cluster: 'default',
          menus: [],
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

    expect(await screen.findByTestId('top-status-panel')).toBeTruthy();

    fireEvent.click(await screen.findByRole('button', { name: 'Login' }));
    const dialog = await screen.findByRole('dialog', { name: 'Login' });
    expect(within(dialog).getByTestId('login-visual-panel')).toBeTruthy();
    expect((within(dialog).getByLabelText('Username') as HTMLInputElement).value).toBe('');
    expect((within(dialog).getByLabelText('Password') as HTMLInputElement).value).toBe('');
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

  it('accepts invite when hash route contains invite token', async () => {
    const originalHash = window.location.hash;
    window.location.hash = '#/accept-invite?token=invite-xyz';

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
            menus: [],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/auth/accept-invite')) {
        return new Response(
          JSON.stringify({
            status: 'accepted',
            tenant_id: 'tenant-dev',
            username: 'new-user',
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

    const dialog = await screen.findByRole('dialog', { name: 'Accept Invite' });
    fireEvent.change(within(dialog).getByLabelText('Username'), {
      target: { value: 'new-user' },
    });
    fireEvent.change(within(dialog).getByLabelText('Password'), {
      target: { value: 'strong-pass' },
    });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Accept Invite' }));

    expect(await screen.findByText('Invite accepted: accepted')).toBeTruthy();
    window.location.hash = originalHash;
  });

  it('opens login dialog when api returns unauthorized with stale token', async () => {
    window.localStorage.setItem('kubedeck.auth.token', 'stale-token');
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith('/api/auth/me')) {
        return new Response('unauthorized', { status: 401 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response('unauthorized', { status: 401 });
      }
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      return new Response(
        JSON.stringify({
          cluster: 'default',
          menus: [],
          resourceTypes: [],
          clusters: ['default'],
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

    expect(await screen.findByRole('dialog', { name: 'Login' })).toBeTruthy();
    expect(window.localStorage.getItem('kubedeck.auth.token')).toBeNull();
  });

  it('shows oauth config diagnostics in login dialog under test mode', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith('/api/auth/oauth/config')) {
        return new Response(
          JSON.stringify({
            mode: 'oidc',
            ready: false,
            provider: 'corp-sso',
            missing: ['KUBEDECK_OIDC_CLIENT_ID', 'KUBEDECK_OIDC_CLIENT_SECRET'],
            oidc: {
              issuer_exists: true,
              client_id_exists: false,
              client_secret_exists: false,
              redirect_url_exists: true,
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(JSON.stringify({ clusters: ['default'] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/registry?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', resourceTypes: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/menus?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', menus: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
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
    const dialog = await screen.findByRole('dialog', { name: 'Login' });

    expect(await within(dialog).findByText('OAuth Config')).toBeTruthy();
    expect(await within(dialog).findByText('Mode: oidc')).toBeTruthy();
    expect(await within(dialog).findByText('Provider: corp-sso')).toBeTruthy();
    expect(await within(dialog).findByText('Ready: no')).toBeTruthy();
    expect(
      await within(dialog).findByText(
        'Missing Fields: KUBEDECK_OIDC_CLIENT_ID, KUBEDECK_OIDC_CLIENT_SECRET',
      ),
    ).toBeTruthy();
  });

  it('completes oauth callback from query params and clears url search', async () => {
    const originalURL = `${window.location.pathname}${window.location.search}${window.location.hash}`;
    window.history.replaceState({}, '', '/?code=oauth-admin&state=state-123');

    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith('/api/auth/oauth/callback')) {
        return new Response(
          JSON.stringify({
            token: 'oauth-token-1',
            user: { id: 'u-oauth', username: 'oauth-admin', roles: ['admin'] },
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({
            user: { id: 'u-oauth', username: 'oauth-admin', activeTenantID: 'tenant-dev', roles: ['admin'] },
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(JSON.stringify({ clusters: ['default'] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/registry?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', resourceTypes: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/menus?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', menus: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      return new Response(JSON.stringify({}), { status: 200, headers: { 'Content-Type': 'application/json' } });
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

    expect(await screen.findByText('oauth-admin')).toBeTruthy();
    expect(window.localStorage.getItem('kubedeck.auth.token')).toBe('oauth-token-1');
    expect(window.location.search).toBe('');
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/auth/oauth/callback',
      expect.objectContaining({ method: 'POST' }),
    );

    window.history.replaceState({}, '', originalURL || '/');
  });

  it('uses state-bound tenant code from sessionStorage for oauth callback', async () => {
    const originalURL = `${window.location.pathname}${window.location.search}${window.location.hash}`;
    window.history.replaceState({}, '', '/?code=oauth-admin&state=state-tenant-bind');
    window.sessionStorage.setItem('kubedeck.oauth.pending.state', 'state-tenant-bind');
    window.sessionStorage.setItem('kubedeck.oauth.pending.tenant_code', 'staging');

    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url.endsWith('/api/auth/oauth/callback')) {
        expect(String(init?.body)).toContain('"tenant_code":"staging"');
        return new Response(
          JSON.stringify({
            token: 'oauth-token-staging',
            user: { id: 'u-oauth', username: 'oauth-admin', roles: ['admin'] },
            tenants: [{ id: 'tenant-staging', code: 'staging', name: 'Staging' }],
            active_tenant_id: 'tenant-staging',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({
            user: { id: 'u-oauth', username: 'oauth-admin', activeTenantID: 'tenant-staging', roles: ['admin'] },
            tenants: [{ id: 'tenant-staging', code: 'staging', name: 'Staging' }],
            active_tenant_id: 'tenant-staging',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(JSON.stringify({ clusters: ['default'] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/registry?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', resourceTypes: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/menus?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', menus: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      return new Response(JSON.stringify({}), { status: 200, headers: { 'Content-Type': 'application/json' } });
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

    expect(await screen.findByText('oauth-admin')).toBeTruthy();
    expect(window.sessionStorage.getItem('kubedeck.oauth.pending.state')).toBeNull();
    expect(window.sessionStorage.getItem('kubedeck.oauth.pending.tenant_code')).toBeNull();
    window.history.replaceState({}, '', originalURL || '/');
  });

  it('loads iam groups, memberships, invites and permissions in access control dialog', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(JSON.stringify({ clusters: ['default'] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/registry?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', resourceTypes: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/menus?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', menus: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/auth/login')) {
        return new Response(
          JSON.stringify({
            token: 'token-abc',
            user: { id: 'u-1', username: 'admin' },
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({
            user: { id: 'u-1', username: 'admin', activeTenantID: 'tenant-dev' },
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/permissions')) {
        return new Response(
          JSON.stringify({
            permissions: [{ code: 'iam:read', scope: 'platform' }],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/groups')) {
        return new Response(
          JSON.stringify({
            groups: [
              {
                id: 'grp-admin',
                tenant_id: 'tenant-dev',
                name: 'admins',
                permissions: ['iam:read'],
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/memberships')) {
        return new Response(
          JSON.stringify({
            memberships: [
              {
                id: 'mbr-u1-tenant-dev',
                tenant_id: 'tenant-dev',
                user_id: 'u-1',
                user_label: 'admin',
                group_ids: ['grp-admin'],
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/users')) {
        return new Response(
          JSON.stringify({
            users: [
              {
                id: 'u-1',
                username: 'admin',
                roles: ['admin'],
                tenant_id: 'tenant-dev',
                membership_id: 'mbr-u1-tenant-dev',
                effective_from: '2026-02-25T00:00:00Z',
                effective_until: '',
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/tenants')) {
        return new Response(
          JSON.stringify({
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.includes('/api/iam/tenants/tenant-dev/members') && !url.includes('membership_id=')) {
        return new Response(
          JSON.stringify({
            members: [
              {
                id: 'mbr-u1-tenant-dev',
                tenant_id: 'tenant-dev',
                user_id: 'u-1',
                user_label: 'admin',
                group_ids: [],
                effective_from: '2026-02-25T00:00:00Z',
                effective_until: '',
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/invites')) {
        return new Response(
          JSON.stringify({
            invites: [
              {
                id: 'inv-1',
                tenant_id: 'tenant-dev',
                tenant_code: 'dev',
                invitee_email: 'user@example.com',
                role_hint: 'member',
                token: 'token-1',
                invite_link: '/accept-invite?token=token-1',
                expires_at: '2026-02-26T00:00:00Z',
                status: 'pending',
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/invites/inv-1')) {
        return new Response(
          JSON.stringify({
            id: 'inv-1',
            tenant_id: 'tenant-dev',
            tenant_code: 'dev',
            invitee_email: 'user@example.com',
            role_hint: 'member',
            token: 'token-1',
            invite_link: '/accept-invite?token=token-1',
            expires_at: '2026-02-26T00:00:00Z',
            status: 'revoked',
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
    fireEvent.click(within(await screen.findByRole('dialog', { name: 'Login' })).getByRole('button', { name: 'Login' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Access Control' }));

    expect(await screen.findByRole('dialog', { name: 'Access Control' })).toBeTruthy();
    expect(await screen.findByDisplayValue('admins')).toBeTruthy();
    expect(await screen.findByText('iam:read (platform)')).toBeTruthy();
    expect(await screen.findByText('Users')).toBeTruthy();
    expect(await screen.findByText('Tenant Members')).toBeTruthy();
    expect(await screen.findByText('Memberships')).toBeTruthy();
    expect(await screen.findByText('Invites')).toBeTruthy();
    expect(await screen.findByText('user@example.com')).toBeTruthy();
    fireEvent.click(await screen.findByRole('button', { name: 'Revoke' }));
    expect(await screen.findByText(/revoked/)).toBeTruthy();
  });

  it('creates tenant member from tenant members section', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(JSON.stringify({ clusters: ['default'] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/registry?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', resourceTypes: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/menus?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', menus: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/auth/login')) {
        return new Response(
          JSON.stringify({
            token: 'token-abc',
            user: { id: 'u-1', username: 'admin', roles: ['admin'] },
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({
            user: { id: 'u-1', username: 'admin', activeTenantID: 'tenant-dev', roles: ['admin'] },
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/permissions')) {
        return new Response(JSON.stringify({ permissions: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/groups')) {
        return new Response(JSON.stringify({ groups: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/memberships')) {
        return new Response(JSON.stringify({ memberships: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/invites')) {
        return new Response(JSON.stringify({ invites: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/users')) {
        return new Response(JSON.stringify({ users: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/tenants')) {
        return new Response(
          JSON.stringify({
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.includes('/api/iam/tenants/tenant-dev/members') && init?.method === 'POST') {
        return new Response(
          JSON.stringify({
            id: 'mbr-u2-tenant-dev',
            tenant_id: 'tenant-dev',
            user_id: 'u-2',
            user_label: 'new-user',
            group_ids: [],
            effective_from: '2026-02-25T00:00:00Z',
            effective_until: '',
          }),
          { status: 201, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.includes('/api/iam/tenants/tenant-dev/members') && (!init?.method || init.method === 'GET')) {
        return new Response(
          JSON.stringify({
            members: [
              {
                id: 'mbr-u2-tenant-dev',
                tenant_id: 'tenant-dev',
                user_id: 'u-2',
                user_label: 'new-user',
                group_ids: [],
                effective_from: '2026-02-25T00:00:00Z',
                effective_until: '',
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

    fireEvent.click(await screen.findByRole('button', { name: 'Login' }));
    fireEvent.click(
      within(await screen.findByRole('dialog', { name: 'Login' })).getByRole('button', {
        name: 'Login',
      }),
    );
    fireEvent.click(await screen.findByRole('button', { name: 'Access Control' }));
    const accessDialog = await screen.findByRole('dialog', { name: 'Access Control' });

    fireEvent.change(await within(accessDialog).findByLabelText('Tenant'), {
      target: { value: 'tenant-dev' },
    });
    fireEvent.change(within(accessDialog).getByLabelText('User ID'), { target: { value: 'u-2' } });
    fireEvent.change(within(accessDialog).getByLabelText('Username'), { target: { value: 'new-user' } });
    fireEvent.change(within(accessDialog).getByLabelText('Effective From (RFC3339)'), {
      target: { value: '2026-02-25T00:00:00Z' },
    });
    fireEvent.click(within(accessDialog).getByRole('button', { name: 'Add Member' }));

    expect(await within(accessDialog).findByText('new-user')).toBeTruthy();
  });

  it('loads audit events with filters in audit dialog', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(JSON.stringify({ clusters: ['default'] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/registry?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', resourceTypes: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/menus?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', menus: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/auth/login')) {
        return new Response(
          JSON.stringify({
            token: 'token-abc',
            user: { id: 'u-1', username: 'admin' },
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({
            user: { id: 'u-1', username: 'admin', activeTenantID: 'tenant-dev' },
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.includes('/api/audit/events')) {
        return new Response(
          JSON.stringify({
            events: [
              {
                tenant_id: 'tenant-dev',
                actor_id: 'u-1',
                action: 'auth.login',
                target_type: 'session',
                target_id: 'token-abc',
                result: 'allowed',
                created_at: '2026-02-25T00:00:00Z',
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

    fireEvent.click(await screen.findByRole('button', { name: 'Login' }));
    fireEvent.click(within(await screen.findByRole('dialog', { name: 'Login' })).getByRole('button', { name: 'Login' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Audit Events' }));
    fireEvent.change(await screen.findByLabelText('Action Filter'), {
      target: { value: 'auth' },
    });
    fireEvent.change(screen.getByLabelText('Result Filter'), {
      target: { value: 'allowed' },
    });
    fireEvent.change(screen.getByLabelText('Result Limit'), {
      target: { value: '10' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));

    expect(await screen.findByText('auth.login')).toBeTruthy();
    await waitFor(() => {
      const urls = fetchMock.mock.calls.map(([input]) => String(input));
      expect(
        urls.some((url) =>
          url.includes('/api/audit/events?action=auth&result=allowed&limit=10'),
        ),
      ).toBe(true);
    });
  });

  it('updates and deletes iam group in access control dialog', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(JSON.stringify({ clusters: ['default'] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/registry?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', resourceTypes: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/menus?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', menus: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/auth/login')) {
        return new Response(
          JSON.stringify({
            token: 'token-abc',
            user: { id: 'u-1', username: 'admin' },
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({
            user: { id: 'u-1', username: 'admin', activeTenantID: 'tenant-dev' },
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/permissions')) {
        return new Response(JSON.stringify({ permissions: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/groups')) {
        return new Response(
          JSON.stringify({
            groups: [
              {
                id: 'grp-admin',
                tenant_id: 'tenant-dev',
                name: 'admins',
                description: 'old description',
                permissions: ['iam:read'],
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/memberships')) {
        return new Response(JSON.stringify({ memberships: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/invites')) {
        return new Response(JSON.stringify({ invites: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/users')) {
        return new Response(JSON.stringify({ users: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/tenants')) {
        return new Response(
          JSON.stringify({
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.includes('/api/iam/tenants/tenant-dev/members') && (!init?.method || init.method === 'GET')) {
        return new Response(JSON.stringify({ members: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/groups/grp-admin') && init?.method === 'PATCH') {
        return new Response(
          JSON.stringify({
            id: 'grp-admin',
            tenant_id: 'tenant-dev',
            name: 'platform-admins',
            description: 'new description',
            permissions: ['iam:read'],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/groups/grp-admin') && init?.method === 'DELETE') {
        return new Response(JSON.stringify({ deleted: 'grp-admin' }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/groups/grp-admin/permissions')) {
        return new Response(
          JSON.stringify({
            id: 'grp-admin',
            tenant_id: 'tenant-dev',
            name: 'admins',
            permissions: ['iam:read'],
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
    fireEvent.click(
      within(await screen.findByRole('dialog', { name: 'Login' })).getByRole('button', {
        name: 'Login',
      }),
    );
    fireEvent.click(await screen.findByRole('button', { name: 'Access Control' }));

    fireEvent.change(await screen.findByDisplayValue('admins'), {
      target: { value: 'platform-admins' },
    });
    fireEvent.change(screen.getByDisplayValue('old description'), {
      target: { value: 'new description' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Save Group' }));
    expect(await screen.findByDisplayValue('platform-admins')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Delete Group' }));
    await waitFor(() => {
      expect(screen.queryByDisplayValue('platform-admins')).toBeNull();
    });
  });

  it('shows IAM read-only mode for viewer role', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
        return new Response('ok', { status: 200 });
      }
      if (url.endsWith('/api/meta/clusters')) {
        return new Response(JSON.stringify({ clusters: ['default'] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/registry?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', resourceTypes: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.includes('/api/meta/menus?cluster=default')) {
        return new Response(JSON.stringify({ cluster: 'default', menus: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/auth/login')) {
        return new Response(
          JSON.stringify({
            token: 'token-viewer',
            user: { id: 'u-2', username: 'viewer', roles: ['viewer'] },
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({
            user: {
              id: 'u-2',
              username: 'viewer',
              activeTenantID: 'tenant-dev',
              roles: ['viewer'],
            },
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
            active_tenant_id: 'tenant-dev',
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/permissions')) {
        return new Response(JSON.stringify({ permissions: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/groups')) {
        return new Response(
          JSON.stringify({
            groups: [
              {
                id: 'grp-view',
                tenant_id: 'tenant-dev',
                name: 'readers',
                permissions: ['iam:read'],
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.endsWith('/api/iam/memberships')) {
        return new Response(JSON.stringify({ memberships: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/invites')) {
        return new Response(JSON.stringify({ invites: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/users')) {
        return new Response(JSON.stringify({ users: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (url.endsWith('/api/iam/tenants')) {
        return new Response(
          JSON.stringify({
            tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.includes('/api/iam/tenants/tenant-dev/members')) {
        return new Response(JSON.stringify({ members: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
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
    fireEvent.click(
      within(await screen.findByRole('dialog', { name: 'Login' })).getByRole('button', {
        name: 'Login',
      }),
    );
    fireEvent.click(await screen.findByRole('button', { name: 'Access Control' }));

    expect(await screen.findByText('Read-only mode: viewer role cannot modify IAM data.')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Create Group' }).hasAttribute('disabled')).toBe(true);
    expect(screen.getByRole('button', { name: 'Save Group' }).hasAttribute('disabled')).toBe(true);
  });
});
