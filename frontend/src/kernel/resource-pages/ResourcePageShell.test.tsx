import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ResourcePageShell } from './ResourcePageShell';
import { resolveDefaultTabs } from './tabs';

describe('ResourcePageShell', () => {
  it('renders the shared shell with default overview and yaml tabs', () => {
    render(
      <ResourcePageShell
        title="deployment/api"
        summary={<div>Status: Running</div>}
        tabs={resolveDefaultTabs({
          overviewContent: <div>Overview Body</div>,
          yamlContent: <div>apiVersion: apps/v1</div>,
        })}
      />,
    );

    expect(screen.getByRole('heading', { name: 'deployment/api' })).toBeTruthy();
    expect(screen.getByText('Status: Running')).toBeTruthy();
    expect(screen.getByRole('tab', { name: 'Overview' })).toBeTruthy();
    expect(screen.getByRole('tab', { name: 'YAML' })).toBeTruthy();
    expect(screen.getByText('Overview Body')).toBeTruthy();

    fireEvent.click(screen.getByRole('tab', { name: 'YAML' }));

    expect(screen.getByText('apiVersion: apps/v1')).toBeTruthy();
  });
});
