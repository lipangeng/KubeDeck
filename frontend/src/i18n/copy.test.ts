import { describe, expect, it } from 'vitest';
import { copy } from './copy';

describe('copy', () => {
  it('resolves known message keys from the default locale', () => {
    expect(copy('app.title')).toBe('KubeDeck');
    expect(copy('app.cleanup.title')).toBe('Microkernel cleanup in progress');
  });
});
