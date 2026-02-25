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
      if (url.includes('cluster=dev')) {
        return new Response(
          JSON.stringify({
            menus: [
              {
                id: 'dev-workloads',
                title: 'Dev Workloads',
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
          menus: [
            {
              id: 'workloads',
              title: 'Workloads',
              source: 'system',
              order: 10,
              visible: true,
            },
            {
              id: 'favorites',
              title: 'Favorites',
              source: 'user',
              order: 20,
              visible: true,
            },
            {
              id: 'crd-dynamic',
              title: 'Custom Resources',
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
    await waitFor(() => {
      const calledUrls = fetchMock.mock.calls.map(([input]) => String(input));
      expect(calledUrls.some((url) => url.includes('cluster=default'))).toBe(true);
      expect(calledUrls.some((url) => url.includes('cluster=dev'))).toBe(true);
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
      return new Response(
        JSON.stringify({
          menus: [
            {
              id: 'workloads',
              title: 'Workloads',
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
