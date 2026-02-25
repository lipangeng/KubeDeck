import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import App from './App';

afterEach(() => {
  vi.restoreAllMocks();
});

describe('App', () => {
  it('renders menus and reloads menus when cluster changes', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('cluster=dev')) {
        return new Response(
          JSON.stringify({
            menus: [{ id: 'dev-workloads', title: 'Dev Workloads' }],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        );
      }

      return new Response(
        JSON.stringify({
          menus: [
            { id: 'workloads', title: 'Workloads' },
            { id: 'crd-dynamic', title: 'Custom Resources' },
          ],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    });

    vi.stubGlobal('fetch', fetchMock);

    render(<App />);

    expect(
      screen.getByRole('heading', {
        level: 1,
        name: 'KubeDeck',
      }),
    ).toBeTruthy();

    expect(await screen.findByText('Workloads')).toBeTruthy();
    expect(await screen.findByText('Custom Resources')).toBeTruthy();

    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'dev' } });

    expect(await screen.findByText('Dev Workloads')).toBeTruthy();
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith('/api/meta/menus?cluster=default');
      expect(fetchMock).toHaveBeenCalledWith('/api/meta/menus?cluster=dev');
    });
  });
});
