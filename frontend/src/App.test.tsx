import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import App from './App';

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe('App', () => {
  it('renders the cleanup baseline message', () => {
    render(<App themePreference="system" onThemePreferenceChange={vi.fn()} />);

    expect(screen.getByText('Microkernel cleanup in progress')).toBeTruthy();
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
