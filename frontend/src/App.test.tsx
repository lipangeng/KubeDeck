import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import App from './App';

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe('App', () => {
  it('renders the homepage built-in contribution by default', () => {
    render(<App themePreference="system" onThemePreferenceChange={vi.fn()} />);

    expect(screen.getAllByText('Homepage')).toHaveLength(2);
    expect(
      screen.getByText(
        'The shell now renders built-in workflow pages through the kernel registry.',
      ),
    ).toBeTruthy();
  });

  it('switches to the workloads built-in contribution through kernel navigation', () => {
    render(<App themePreference="system" onThemePreferenceChange={vi.fn()} />);

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
