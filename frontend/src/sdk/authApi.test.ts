import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  acceptInvite,
  AUTH_TOKEN_KEY,
  clearAuthToken,
  login,
  me,
  oauthCallback,
  oauthConfig,
  oauthURL,
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

  it('accepts invite by token with signup credentials', async () => {
    const fetchMock = vi.fn(async () => {
      return new Response(
        JSON.stringify({
          status: 'accepted',
          tenant_id: 'tenant-dev',
          username: 'new-user',
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    });
    vi.stubGlobal('fetch', fetchMock);

    const payload = await acceptInvite('invite-token', 'new-user', 'strong-pass');

    expect(payload.status).toBe('accepted');
    expect(payload.tenant_id).toBe('tenant-dev');
    const init = (fetchMock.mock.calls[0] as unknown[])[1] as RequestInit;
    expect(init.method).toBe('POST');
    expect(String(init.body)).toContain('"token":"invite-token"');
  });

  it('loads oauth authorize url payload', async () => {
    const fetchMock = vi.fn(async () => {
      return new Response(
        JSON.stringify({
          provider: 'github',
          state: 'state-token',
          auth_url: 'https://example.com/oauth/authorize?state=state-token',
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    });
    vi.stubGlobal('fetch', fetchMock);

    const payload = await oauthURL();

    expect(payload.provider).toBe('github');
    expect(payload.state).toBe('state-token');
    expect(payload.auth_url).toContain('state=state-token');
  });

  it('parses oauth config diagnostics payload fields', async () => {
    const fetchMock = vi.fn(async () => {
      return new Response(
        JSON.stringify({
          mode: 'oidc',
          provider: 'corp-sso',
          ready: false,
          missing: ['KUBEDECK_OIDC_CLIENT_ID', 'KUBEDECK_OIDC_CLIENT_SECRET'],
          oidc: {
            issuer_exists: true,
            client_id_exists: false,
            client_secret_exists: false,
            redirect_url_exists: true,
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    });
    vi.stubGlobal('fetch', fetchMock);

    const payload = await oauthConfig();

    expect(payload.mode).toBe('oidc');
    expect(payload.ready).toBe(false);
    expect(payload.provider).toBe('corp-sso');
    expect(payload.missing).toEqual(['KUBEDECK_OIDC_CLIENT_ID', 'KUBEDECK_OIDC_CLIENT_SECRET']);
    expect(payload.oidc.client_id_exists).toBe(false);
    expect(payload.oidc.issuer_exists).toBe(true);
  });

  it('calls oauth callback with code and tenant_code', async () => {
    const fetchMock = vi.fn(async () => {
      return new Response(
        JSON.stringify({
          token: 'oauth-token',
          user: { id: 'u-oauth', username: 'oauth-admin' },
          tenants: [{ id: 'tenant-dev', code: 'dev', name: 'Development' }],
          active_tenant_id: 'tenant-dev',
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    });
    vi.stubGlobal('fetch', fetchMock);

    const payload = await oauthCallback('oauth-admin', 'dev', 'oauth-state');

    expect(payload.token).toBe('oauth-token');
    const init = (fetchMock.mock.calls[0] as unknown[])[1] as RequestInit;
    expect(init.method).toBe('POST');
    expect(String(init.body)).toContain('"code":"oauth-admin"');
    expect(String(init.body)).toContain('"tenant_code":"dev"');
    expect(String(init.body)).toContain('"state":"oauth-state"');
  });
});
