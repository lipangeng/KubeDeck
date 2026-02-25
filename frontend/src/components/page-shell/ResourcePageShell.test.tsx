import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { DetailPageShell, ListPageShell } from './ResourcePageShell';

describe('ResourcePageShell', () => {
  it('renders list shell with toolbar and content slot', () => {
    render(
      <ListPageShell
        title="Pods"
        toolbar={<button type="button">Create</button>}
      >
        <div>List Content</div>
      </ListPageShell>,
    );

    expect(screen.getByRole('heading', { name: 'Pods' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Create' })).toBeTruthy();
    expect(screen.getByText('List Content')).toBeTruthy();
  });

  it('renders detail shell with summary and panels', () => {
    render(
      <DetailPageShell
        title="pod/nginx"
        summary={<div>Status: Running</div>}
        sidePanel={<div>Events</div>}
      >
        <div>Detail Body</div>
      </DetailPageShell>,
    );

    expect(screen.getByRole('heading', { name: 'pod/nginx' })).toBeTruthy();
    expect(screen.getByText('Status: Running')).toBeTruthy();
    expect(screen.getByText('Events')).toBeTruthy();
    expect(screen.getByText('Detail Body')).toBeTruthy();
  });
});
