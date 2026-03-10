import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import App from './App';

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

function createKernelMetadataFetchMock() {
  return vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.endsWith('/api/meta/kernel')) {
      return new Response(
        JSON.stringify({
          pages: [
            {
              ID: 'page.homepage',
              WorkflowDomainID: 'homepage',
              Route: '/',
              EntryKey: 'homepage',
              Title: { Key: 'homepage.title', Fallback: 'Homepage' },
              Description: {
                Key: 'homepage.description',
                Fallback: 'The shell now renders built-in workflow pages through the kernel registry.',
              },
            },
            {
              ID: 'page.workloads',
              WorkflowDomainID: 'workloads',
              Route: '/workloads',
              EntryKey: 'workloads',
              Title: { Key: 'workloads.title', Fallback: 'Workloads' },
            },
          ],
          menus: [
            {
              ID: 'menu.homepage',
              WorkflowDomainID: 'homepage',
              EntryKey: 'homepage',
              Route: '/',
              Placement: 'primary',
              Order: 10,
              Visible: true,
              Title: { Key: 'homepage.title', Fallback: 'Homepage' },
            },
            {
              ID: 'menu.workloads',
              WorkflowDomainID: 'workloads',
              EntryKey: 'workloads',
              Route: '/workloads',
              Placement: 'primary',
              Order: 20,
              Visible: true,
              Title: { Key: 'workloads.title', Fallback: 'Workloads' },
            },
          ],
          actions: [
            {
              ID: 'create',
              WorkflowDomainID: 'workloads',
              Surface: 'drawer',
              Visible: true,
              Title: { Key: 'actions.create', Fallback: 'Create' },
            },
            {
              ID: 'apply',
              WorkflowDomainID: 'workloads',
              Surface: 'drawer',
              Visible: true,
              Title: { Key: 'actions.apply', Fallback: 'Apply' },
            },
          ],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    }

    throw new Error(`Unhandled fetch request: ${url}`);
  });
}

describe('App', () => {
  it('renders the homepage built-in contribution by default', async () => {
    vi.stubGlobal('fetch', createKernelMetadataFetchMock());
    render(<App themePreference="system" onThemePreferenceChange={vi.fn()} />);

    expect(await screen.findByText('Kernel metadata source: backend')).toBeTruthy();
    expect(screen.getAllByText('Homepage')).toHaveLength(2);
    expect(
      screen.getByText(
        'The shell now renders built-in workflow pages through the kernel registry.',
      ),
    ).toBeTruthy();
  });

  it('switches to the workloads built-in contribution through kernel navigation', async () => {
    vi.stubGlobal('fetch', createKernelMetadataFetchMock());
    render(<App themePreference="system" onThemePreferenceChange={vi.fn()} />);

    expect(await screen.findByText('Kernel metadata source: backend')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Workloads' }));

    expect(screen.getByText('Built-in Capability')).toBeTruthy();
    expect(
      screen.getByText(
        'This is the first built-in workflow domain registered through the kernel contribution system.',
      ),
    ).toBeTruthy();
    expect(screen.getByText('Registered actions: Create, Apply')).toBeTruthy();
  });

  it('cycles the theme preference', () => {
    vi.stubGlobal('fetch', createKernelMetadataFetchMock());
    const onThemePreferenceChange = vi.fn();

    render(
      <App
        themePreference="system"
        onThemePreferenceChange={onThemePreferenceChange}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Theme: system' }));
    expect(onThemePreferenceChange).toHaveBeenCalledWith('light');
  });
});
