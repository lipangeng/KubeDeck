import { render, screen } from '@testing-library/react';
import { createElement } from 'react';
import { describe, expect, it } from 'vitest';
import { copy, useCopy } from './copy';
import { LocaleProvider } from './localeContext';

describe('copy', () => {
  it('resolves known message keys from the default locale', () => {
    expect(copy('app.title')).toBe('KubeDeck');
    expect(copy('app.cleanup.title')).toBe('Microkernel cleanup in progress');
  });

  it('resolves messages through the locale context boundary', () => {
    function Probe() {
      return createElement('span', null, useCopy('app.title'));
    }

    render(
      createElement(LocaleProvider, { initialLocale: 'en' }, createElement(Probe)),
    );

    expect(screen.getByText('KubeDeck')).toBeTruthy();
  });
});
