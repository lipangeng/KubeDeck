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
  };
  tenants: AuthTenant[];
  active_tenant_id: string;
}

export interface AuthLoginResponse {
  token: string;
  user: {
    id: string;
    username: string;
  };
  tenants: AuthTenant[];
  active_tenant_id: string;
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

function parseAuthLoginResponse(value: unknown): AuthLoginResponse {
  if (!isObject(value) || typeof value.token !== 'string' || !isObject(value.user)) {
    throw new Error('invalid auth login response');
  }
  return {
    token: value.token,
    user: {
      id: String((value.user as Record<string, unknown>).id ?? ''),
      username: String((value.user as Record<string, unknown>).username ?? ''),
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
