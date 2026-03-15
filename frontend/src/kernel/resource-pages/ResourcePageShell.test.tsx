import { afterEach, describe, expect, it } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { ResourcePageShell } from './ResourcePageShell';
import { resolveDefaultTabs } from './tabs';

afterEach(() => {
  cleanup();
});

describe('ResourcePageShell', () => {
  it('renders the shared shell with default overview and yaml tabs', () => {
    render(
      <ResourcePageShell
        title="deployment/api"
        summary={<div>Status: Running</div>}
        actions={<button type="button">Apply</button>}
        tabs={resolveDefaultTabs({
          overviewContent: <div>Overview Body</div>,
          yamlContent: <div>apiVersion: apps/v1</div>,
        })}
      />,
    );

    expect(screen.getByRole('heading', { name: 'deployment/api' })).toBeTruthy();
    expect(screen.getByText('Status: Running')).toBeTruthy();
    expect(screen.getAllByRole('button', { name: 'Apply' }).length).toBeGreaterThan(0);
    expect(screen.getByRole('tab', { name: 'Overview' })).toBeTruthy();
    expect(screen.getByRole('tab', { name: 'YAML' })).toBeTruthy();
    expect(screen.getByText('Overview Body')).toBeTruthy();

    fireEvent.click(screen.getByRole('tab', { name: 'YAML' }));

    expect(screen.getByText('apiVersion: apps/v1')).toBeTruthy();
  });

  it('renders takeover content inside the shared shell without tabs', () => {
    render(
      <ResourcePageShell
        title="StatefulSet/db"
        summary={<div>Namespace: default</div>}
        actions={<button type="button">Apply</button>}
        takeoverContent={<div>Plugin takeover body</div>}
      />,
    );

    expect(screen.getByRole('heading', { name: 'StatefulSet/db' })).toBeTruthy();
    expect(screen.getByText('Namespace: default')).toBeTruthy();
    expect(screen.getAllByRole('button', { name: 'Apply' }).length).toBeGreaterThan(0);
    expect(screen.queryByRole('tab')).toBeNull();
    expect(screen.getByText('Plugin takeover body')).toBeTruthy();
  });
});
