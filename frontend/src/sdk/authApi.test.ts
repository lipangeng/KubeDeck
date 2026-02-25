import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  AUTH_TOKEN_KEY,
  clearAuthToken,
  login,
  me,
  readAuthToken,
  switchTenant,
  writeAuthToken,
} from './authApi';

afterEach(() => {
  window.localStorage.clear();
  vi.restoreAllMocks();
});

describe('authApi', () => {
  it('reads/writes/clears auth token', () => {
    expect(readAuthToken()).toBeNull();
    writeAuthToken('token-123');
    expect(window.localStorage.getItem(AUTH_TOKEN_KEY)).toBe('token-123');
    expect(readAuthToken()).toBe('token-123');
    clearAuthToken();
    expect(readAuthToken()).toBeNull();
  });

  it('logs in with tenant_code and parses response', async () => {
    const fetchMock = vi.fn(async () => {
      return new Response(
        JSON.stringify({
          token: 'token-abc',
          user: { id: 'u-1', username: 'admin' },
          tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
          active_tenant_id: 'tenant-dev',
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    });
    vi.stubGlobal('fetch', fetchMock);

    const payload = await login('admin', 'admin', 'dev');

    expect(payload.token).toBe('token-abc');
    expect(payload.user.username).toBe('admin');
    expect(fetchMock).toHaveBeenCalled();
    const init = (fetchMock.mock.calls[0] as unknown[])[1] as RequestInit;
    expect(init.method).toBe('POST');
    expect(String(init.body)).toContain('"tenant_code":"dev"');
  });

  it('loads /me with bearer token and fallback active tenant mapping', async () => {
    const fetchMock = vi.fn(async () => {
      return new Response(
        JSON.stringify({
          user: { id: 'u-1', username: 'admin' },
          tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
          active_tenant_id: 'tenant-dev',
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    });
    vi.stubGlobal('fetch', fetchMock);

    const payload = await me('token-abc');

    expect(payload.user.id).toBe('u-1');
    expect(payload.user.activeTenantID).toBe('tenant-dev');
    expect(fetchMock).toHaveBeenCalled();
    const init = (fetchMock.mock.calls[0] as unknown[])[1] as RequestInit;
    expect((init.headers as Record<string, string>).Authorization).toBe('Bearer token-abc');
  });

  it('switches tenant by tenant_code', async () => {
    const fetchMock = vi.fn(async () => {
      return new Response(
        JSON.stringify({
          active_tenant_id: 'tenant-staging',
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    });
    vi.stubGlobal('fetch', fetchMock);

    const payload = await switchTenant('token-abc', 'staging');

    expect(payload.active_tenant_id).toBe('tenant-staging');
    expect(fetchMock).toHaveBeenCalled();
    const init = (fetchMock.mock.calls[0] as unknown[])[1] as RequestInit;
    expect(init.method).toBe('POST');
    expect((init.headers as Record<string, string>).Authorization).toBe('Bearer token-abc');
    expect(String(init.body)).toContain('"tenant_code":"staging"');
  });
});
