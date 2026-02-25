export interface AuthTenant {
  id: string;
  code: string;
  name: string;
}

export interface AuthMeResponse {
  user: {
    id: string;
    username: string;
    activeTenantID: string;
    roles: string[];
  };
  tenants: AuthTenant[];
  active_tenant_id: string;
}

export interface AuthLoginResponse {
  token: string;
  user: {
    id: string;
    username: string;
    roles: string[];
  };
  tenants: AuthTenant[];
  active_tenant_id: string;
}

export interface OAuthURLResponse {
  provider: string;
  state: string;
  auth_url: string;
}

export const AUTH_TOKEN_KEY = 'kubedeck.auth.token';

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

export function readAuthToken(): string | null {
  if (typeof window === 'undefined') {
    return null;
  }
  const token = window.localStorage.getItem(AUTH_TOKEN_KEY);
  return token && token.trim() !== '' ? token : null;
}

export function writeAuthToken(token: string): void {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.setItem(AUTH_TOKEN_KEY, token);
}

export function clearAuthToken(): void {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.removeItem(AUTH_TOKEN_KEY);
}

export async function login(
  username: string,
  password: string,
  tenantCode: string,
): Promise<AuthLoginResponse> {
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      username,
      password,
      tenant_code: tenantCode,
    }),
  });
  if (!response.ok) {
    throw new Error(`auth login failed: ${response.status}`);
  }
  const payload = await response.json();
  return parseAuthLoginResponse(payload);
}

export async function me(token: string): Promise<AuthMeResponse> {
  const response = await fetch('/api/auth/me', {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!response.ok) {
    throw new Error(`auth me failed: ${response.status}`);
  }
  const payload = await response.json();
  return parseAuthMeResponse(payload);
}

export async function switchTenant(
  token: string,
  tenantCode: string,
): Promise<{ active_tenant_id: string }> {
  const response = await fetch('/api/auth/switch-tenant', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({
      tenant_code: tenantCode,
    }),
  });
  if (!response.ok) {
    throw new Error(`auth switch tenant failed: ${response.status}`);
  }
  const payload = await response.json();
  if (!isObject(payload) || typeof payload.active_tenant_id !== 'string') {
    throw new Error('invalid switch tenant response');
  }
  return {
    active_tenant_id: payload.active_tenant_id,
  };
}

export async function acceptInvite(
  token: string,
  username: string,
  password: string,
): Promise<{ status: string; tenant_id: string; username: string }> {
  const response = await fetch('/api/auth/accept-invite', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      token,
      username,
      password,
    }),
  });
  if (!response.ok) {
    throw new Error(`accept invite failed: ${response.status}`);
  }
  const payload = await response.json();
  if (!isObject(payload) || typeof payload.status !== 'string') {
    throw new Error('invalid accept invite response');
  }
  return {
    status: payload.status,
    tenant_id: String(payload.tenant_id ?? ''),
    username: String(payload.username ?? ''),
  };
}

export async function oauthURL(): Promise<OAuthURLResponse> {
  const response = await fetch('/api/auth/oauth/url');
  if (!response.ok) {
    throw new Error(`oauth url failed: ${response.status}`);
  }
  const payload = await response.json();
  if (!isObject(payload) || typeof payload.auth_url !== 'string') {
    throw new Error('invalid oauth url response');
  }
  return {
    provider: String(payload.provider ?? ''),
    state: String(payload.state ?? ''),
    auth_url: payload.auth_url,
  };
}

export async function oauthCallback(code: string, tenantCode: string): Promise<AuthLoginResponse> {
  const response = await fetch('/api/auth/oauth/callback', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      code,
      tenant_code: tenantCode,
    }),
  });
  if (!response.ok) {
    throw new Error(`oauth callback failed: ${response.status}`);
  }
  const payload = await response.json();
  return parseAuthLoginResponse(payload);
}

function parseAuthLoginResponse(value: unknown): AuthLoginResponse {
  if (!isObject(value) || typeof value.token !== 'string' || !isObject(value.user)) {
    throw new Error('invalid auth login response');
  }
  return {
    token: value.token,
    user: {
      id: String((value.user as Record<string, unknown>).id ?? ''),
      username: String((value.user as Record<string, unknown>).username ?? ''),
      roles: Array.isArray((value.user as Record<string, unknown>).roles)
        ? ((value.user as Record<string, unknown>).roles as unknown[])
            .map((role) => String(role))
            .filter((role) => role.trim() !== '')
        : [],
    },
    tenants: parseTenants(value.tenants),
    active_tenant_id: String(value.active_tenant_id ?? ''),
  };
}

function parseAuthMeResponse(value: unknown): AuthMeResponse {
  if (!isObject(value) || !isObject(value.user)) {
    throw new Error('invalid auth me response');
  }
  return {
    user: {
      id: String((value.user as Record<string, unknown>).id ?? ''),
      username: String((value.user as Record<string, unknown>).username ?? ''),
      activeTenantID: String(
        (value.user as Record<string, unknown>).activeTenantID ?? value.active_tenant_id ?? '',
      ),
      roles: Array.isArray((value.user as Record<string, unknown>).roles)
        ? ((value.user as Record<string, unknown>).roles as unknown[])
            .map((role) => String(role))
            .filter((role) => role.trim() !== '')
        : [],
    },
    tenants: parseTenants(value.tenants),
    active_tenant_id: String(value.active_tenant_id ?? ''),
  };
}

function parseTenants(value: unknown): AuthTenant[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value
    .filter((item) => isObject(item))
    .map((item) => ({
      id: String(item.id ?? ''),
      code: String(item.code ?? ''),
      name: String(item.name ?? ''),
    }));
}
