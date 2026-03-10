import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import App from './App';

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

function createDefaultFetchMock() {
  return vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);

    if (url.endsWith('/api/healthz') || url.endsWith('/api/readyz')) {
      return new Response('ok', { status: 200 });
    }

    if (url.endsWith('/api/meta/clusters')) {
      return new Response(
        JSON.stringify({
          clusters: ['dev', 'staging', 'prod'],
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

    if (url.includes('/api/meta/registry?cluster=prod')) {
      return new Response(
        JSON.stringify({
          cluster: 'prod',
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

    if (url.includes('/api/meta/menus?cluster=dev')) {
      return new Response(
        JSON.stringify({
          cluster: 'dev',
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
          ],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    }

    if (url.includes('/api/meta/menus?cluster=prod')) {
      return new Response(
        JSON.stringify({
          cluster: 'prod',
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
              id: 'crd-dynamic',
              group: 'dynamic',
              title: 'Custom Resources',
              targetType: 'resource',
              targetRef: '/resources/custom',
              source: 'dynamic',
              order: 50,
              visible: true,
            },
          ],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    }

    if (url.endsWith('/api/resources/apply')) {
      expect(init?.method).toBe('POST');
      return new Response(
        JSON.stringify({
          status: 'accepted',
          message: 'resource apply stub',
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    }

    throw new Error(`Unhandled fetch request: ${url}`);
  });
}

function createApplyFailureFetchMock() {
  return vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);

    if (url.endsWith('/api/resources/apply')) {
      expect(init?.method).toBe('POST');
      return new Response(
        JSON.stringify({
          error: 'apply failed',
        }),
        { status: 500, headers: { 'Content-Type': 'application/json' } },
      );
    }

    return createDefaultFetchMock()(input, init);
  });
}

describe('App', () => {
  it('enters workloads as the primary workflow and reloads workload types on cluster change', async () => {
    const fetchMock = createDefaultFetchMock();
    const onThemePreferenceChange = vi.fn();
    vi.stubGlobal('fetch', fetchMock);

    render(
      <App
        themePreference="system"
        onThemePreferenceChange={onThemePreferenceChange}
      />,
    );

    expect(screen.getByRole('heading', { level: 1, name: 'KubeDeck' })).toBeTruthy();
    expect(await screen.findByText('Cluster dev')).toBeTruthy();
    expect(screen.getByText('Namespace scope: default')).toBeTruthy();
    expect(screen.getByText('Primary Workflow')).toBeTruthy();
    expect(screen.queryByText('System Menus')).toBeNull();

    fireEvent.change(screen.getByLabelText('Theme'), {
      target: { value: 'dark' },
    });
    expect(onThemePreferenceChange).toHaveBeenCalledWith('dark');

    fireEvent.click(screen.getAllByRole('button', { name: 'Enter Workloads' })[0]);

    expect(await screen.findByRole('heading', { level: 2, name: 'Workloads' })).toBeTruthy();
    expect(screen.getByText('Cluster: dev')).toBeTruthy();
    expect(screen.getByText('Namespace scope: default')).toBeTruthy();
    expect(screen.getByTestId('workload-type-count').textContent).toContain('1');
    expect(screen.getByText('Deployment')).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Cluster'), {
      target: { value: 'prod' },
    });

    await screen.findByText('Cluster: prod');
    await waitFor(() => {
      expect(screen.getByTestId('workload-type-count').textContent).toContain('2');
    });
    expect(screen.getByText('Service')).toBeTruthy();

    const calledUrls = fetchMock.mock.calls.map(([input]) => String(input));
    expect(calledUrls.some((url) => url.endsWith('/api/meta/clusters'))).toBe(true);
    expect(calledUrls.some((url) => url.includes('/api/meta/menus?cluster=dev'))).toBe(true);
    expect(calledUrls.some((url) => url.includes('/api/meta/menus?cluster=prod'))).toBe(true);
    expect(calledUrls.some((url) => url.includes('/api/meta/registry?cluster=dev'))).toBe(true);
    expect(calledUrls.some((url) => url.includes('/api/meta/registry?cluster=prod'))).toBe(true);
  });

  it('requires an explicit execution namespace when applying from all namespaces', async () => {
    const fetchMock = createDefaultFetchMock();
    vi.stubGlobal('fetch', fetchMock);

    render(
      <App
        themePreference="system"
        onThemePreferenceChange={vi.fn()}
      />,
    );

    fireEvent.click((await screen.findAllByRole('button', { name: 'Enter Workloads' }))[0]);
    await screen.findByText('Cluster: dev');

    fireEvent.change(screen.getByLabelText('Namespace Scope'), {
      target: { value: 'all' },
    });
    expect(screen.getByText('Namespace scope: All namespaces')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Apply' }));
    expect(await screen.findByText('Apply in dev')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Submit Apply' }));
    expect(
      await screen.findByText(
        'Namespace target is required when browsing all namespaces.',
      ),
    ).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Execution namespace'), {
      target: { value: 'team-a' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Submit Apply' }));

    await screen.findByText('Apply succeeded');
    expect(screen.getByText('Affected: apply accepted for team-a')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Back to Workloads' }));
    await screen.findByText('Apply accepted: apply accepted for team-a');

    const applyCall = fetchMock.mock.calls.find(([input]) =>
      String(input).endsWith('/api/resources/apply'),
    );
    expect(applyCall).toBeTruthy();
    expect(applyCall?.[1]?.body).toContain('"cluster":"dev"');
    expect(applyCall?.[1]?.body).toContain('"namespace":"team-a"');
  });

  it('supports create through the same action surface and result flow', async () => {
    const fetchMock = createDefaultFetchMock();
    vi.stubGlobal('fetch', fetchMock);

    render(
      <App
        themePreference="system"
        onThemePreferenceChange={vi.fn()}
      />,
    );

    fireEvent.click((await screen.findAllByRole('button', { name: 'Enter Workloads' }))[0]);
    await screen.findByText('Cluster: dev');

    fireEvent.click(screen.getByRole('button', { name: 'Create' }));
    expect(await screen.findByText('Create in dev')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Submit Create' })).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Manifest'), {
      target: {
        value: [
          'apiVersion: v1',
          'kind: Service',
          'metadata:',
          '  name: create-service',
          'spec:',
          '  selector:',
          '    app: demo',
        ].join('\n'),
      },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Submit Create' }));

    await screen.findByText('Create succeeded');
    expect(screen.getByText('Affected: create accepted for default')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Back to Workloads' }));
    await screen.findByText('Create accepted: create accepted for default');

    const createCall = fetchMock.mock.calls.find(
      ([input, init]) =>
        String(input).endsWith('/api/resources/apply') &&
        String(init?.body).includes('"actionType":"create"'),
    );
    expect(createCall).toBeTruthy();
    expect(createCall?.[1]?.body).toContain('"namespace":"default"');
  });

  it('preserves apply input on failure and lets the user return to editing', async () => {
    const fetchMock = createApplyFailureFetchMock();
    vi.stubGlobal('fetch', fetchMock);

    render(
      <App
        themePreference="system"
        onThemePreferenceChange={vi.fn()}
      />,
    );

    fireEvent.click((await screen.findAllByRole('button', { name: 'Enter Workloads' }))[0]);
    await screen.findByText('Cluster: dev');

    fireEvent.click(screen.getByRole('button', { name: 'Apply' }));
    expect(await screen.findByText('Apply in dev')).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Manifest'), {
      target: {
        value: [
          'apiVersion: apps/v1',
          'kind: Deployment',
          'metadata:',
          '  name: failed-api',
          'spec:',
          '  replicas: 2',
        ].join('\n'),
      },
    });

    fireEvent.click(screen.getByRole('button', { name: 'Submit Apply' }));

    await screen.findByText('Apply failed');
    expect(screen.getByText('Failed: apply request failed: 500')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Back to Edit' }));

    const manifestField = await screen.findByLabelText('Manifest');
    const namespaceField = screen.getByLabelText('Execution namespace');
    expect((manifestField as HTMLInputElement).value).toContain('failed-api');
    expect((namespaceField as HTMLInputElement).value).toBe('default');
  });

  it('shows blocking summary when readiness check fails', async () => {
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
            clusters: ['dev'],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }
      if (url.includes('/api/meta/menus?cluster=dev')) {
        return new Response(
          JSON.stringify({
            cluster: 'dev',
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
      }
      if (url.includes('/api/meta/registry?cluster=dev')) {
        return new Response(
          JSON.stringify({
            cluster: 'dev',
            resourceTypes: [],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }

      throw new Error(`Unhandled fetch request: ${url}`);
    });

    vi.stubGlobal('fetch', fetchMock);

    render(
      <App
        themePreference="system"
        onThemePreferenceChange={vi.fn()}
      />,
    );

    expect(await screen.findByText(/Blocking summary:/)).toBeTruthy();
    expect(screen.getByText(/readyz: status 503/)).toBeTruthy();
  });
});
