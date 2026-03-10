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
            {
              ID: 'page.operations',
              WorkflowDomainID: 'operations',
              Route: '/operations',
              EntryKey: 'operations',
              Title: { Key: 'operations.title', Fallback: 'Operations' },
              Description: {
                Key: 'operations.description',
                Fallback:
                  'This page is composed from backend capability metadata and rendered through the generic runtime page.',
              },
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
            {
              ID: 'menu.operations',
              WorkflowDomainID: 'operations',
              EntryKey: 'operations',
              Route: '/operations',
              Placement: 'primary',
              Order: 30,
              Visible: true,
              Title: { Key: 'operations.title', Fallback: 'Operations' },
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
          slots: [
            {
              ID: 'slot.workloads.summary.insights',
              WorkflowDomainID: 'workloads',
              SlotID: 'workloads.summary.insights',
              Placement: 'summary',
              Visible: true,
              Title: { Key: 'workloads.insights.title', Fallback: 'Kernel Insights' },
            },
          ],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    }

    if (url.includes('/api/workflows/workloads/items?workflowDomainId=workloads&cluster=default')) {
      return new Response(
        JSON.stringify([
          {
            id: 'workload-api-default',
            name: 'api',
            kind: 'Deployment',
            namespace: 'default',
            status: 'Running',
            health: 'Healthy',
            updatedAt: '2026-03-10T10:00:00Z',
          },
          {
            id: 'workload-web-default',
            name: 'web',
            kind: 'Deployment',
            namespace: 'default',
            status: 'Pending',
            health: 'Warning',
            updatedAt: '2026-03-10T10:05:00Z',
          },
        ]),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    }

    if (url.endsWith('/api/actions/execute')) {
      return new Response(
        JSON.stringify({
          Accepted: true,
          Summary: 'apply accepted',
          AffectedObjects: ['deployment/sample'],
          FailedObjects: [],
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
    expect(await screen.findByText('api')).toBeTruthy();
    expect(screen.getByText('Kernel Insights')).toBeTruthy();
  });

  it('executes a kernel action through the backend action entry', async () => {
    vi.stubGlobal('fetch', createKernelMetadataFetchMock());
    render(<App themePreference="system" onThemePreferenceChange={vi.fn()} />);

    expect(await screen.findByText('Kernel metadata source: backend')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Workloads' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Run Apply' }));

    expect(await screen.findByText('Last action result: apply accepted')).toBeTruthy();
  });

  it('renders a remote-only page through the generic kernel runtime page', async () => {
    vi.stubGlobal('fetch', createKernelMetadataFetchMock());
    render(<App themePreference="system" onThemePreferenceChange={vi.fn()} />);

    expect(await screen.findByText('Kernel metadata source: backend')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Operations' }));

    expect(screen.getByText('Remote Capability')).toBeTruthy();
    expect(
      screen.getByText(
        'This workflow page is rendered by the generic kernel runtime because no built-in page implementation was registered locally.',
      ),
    ).toBeTruthy();
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
