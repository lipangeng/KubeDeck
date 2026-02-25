import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import App from './App';

afterEach(() => {
  vi.restoreAllMocks();
});

describe('App', () => {
  it('renders KubeDeck header and menu items from API', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => ({
        ok: true,
        json: async () => ({
          menus: [
            { id: 'workloads', title: 'Workloads' },
            { id: 'crd-dynamic', title: 'Custom Resources' },
          ],
        }),
      })),
    );

    render(<App />);

    expect(
      screen.getByRole('heading', {
        level: 1,
        name: 'KubeDeck',
      }),
    ).toBeTruthy();

    expect(await screen.findByText('Workloads')).toBeTruthy();
    expect(await screen.findByText('Custom Resources')).toBeTruthy();
  });
});
