import { useEffect, useState } from 'react';

interface MenuItem {
  id: string;
  title: string;
}

interface MenusResponse {
  menus: MenuItem[];
}

type ProbeStatus = 'checking' | 'ok' | 'error';

function resolveApiTarget(): string {
  const configured = import.meta.env.VITE_BACKEND_TARGET as string | undefined;
  if (configured && configured.trim() !== '') {
    return configured.replace(/\/$/, '');
  }

  return 'http://127.0.0.1:8080';
}

function resolveApiTargetHint(): string {
  const mode = import.meta.env.MODE;
  if (mode === 'development' || mode === 'test') {
    return `${mode}: ${resolveApiTarget()}`;
  }

  return mode;
}

function resolveProbePath(path: '/healthz' | '/readyz'): string {
  return `/api${path}`;
}

function App() {
  const apiTargetHint = resolveApiTargetHint();
  const [activeCluster, setActiveCluster] = useState('default');
  const [menus, setMenus] = useState<MenuItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [healthStatus, setHealthStatus] = useState<ProbeStatus>('checking');
  const [readyStatus, setReadyStatus] = useState<ProbeStatus>('checking');
  const [healthError, setHealthError] = useState<string | null>(null);
  const [readyError, setReadyError] = useState<string | null>(null);
  const [lastCheckedAt, setLastCheckedAt] = useState<string | null>(null);

  useEffect(() => {
    let active = true;

    async function loadMenus() {
      setLoading(true);
      setError(null);
      try {
        const response = await fetch(
          `/api/meta/menus?cluster=${encodeURIComponent(activeCluster)}`,
        );
        if (!response.ok) {
          throw new Error(`menus request failed: ${response.status}`);
        }

        const payload = (await response.json()) as MenusResponse;
        if (active) {
          setMenus(payload.menus ?? []);
        }
      } catch (e) {
        if (active) {
          setError(e instanceof Error ? e.message : 'unknown error');
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    }

    loadMenus();

    return () => {
      active = false;
    };
  }, [activeCluster]);

  useEffect(() => {
    let active = true;

    async function checkEndpoint(
      path: '/healthz' | '/readyz',
    ): Promise<{ status: ProbeStatus; error: string | null }> {
      try {
        const response = await fetch(resolveProbePath(path));
        if (!response.ok) {
          throw new Error(`status ${response.status}`);
        }
        return { status: 'ok', error: null };
      } catch (e) {
        if (e instanceof Error) {
          return { status: 'error', error: e.message };
        }
        return { status: 'error', error: 'network error' };
      }
    }

    async function probeRuntime() {
      if (active) {
        setHealthStatus('checking');
        setReadyStatus('checking');
        setHealthError(null);
        setReadyError(null);
      }
      const [healthResult, readyResult] = await Promise.all([
        checkEndpoint('/healthz'),
        checkEndpoint('/readyz'),
      ]);
      if (active) {
        setHealthStatus(healthResult.status);
        setReadyStatus(readyResult.status);
        setHealthError(healthResult.error);
        setReadyError(readyResult.error);
        setLastCheckedAt(new Date().toISOString());
      }
    }

    probeRuntime();
    const timer = setInterval(probeRuntime, 10_000);

    return () => {
      active = false;
      clearInterval(timer);
    };
  }, []);

  return (
    <main>
      <h1>KubeDeck</h1>
      <p>API target ({apiTargetHint})</p>
      <section aria-label="Runtime Status">
        <p>healthz: {healthStatus}</p>
        <p>readyz: {readyStatus}</p>
        <p>Last checked: {lastCheckedAt ?? 'never'}</p>
        <p>
          Failure summary:{' '}
          {healthError || readyError
            ? [healthError ? `healthz: ${healthError}` : null, readyError ? `readyz: ${readyError}` : null]
                .filter(Boolean)
                .join('; ')
            : 'none'}
        </p>
      </section>
      <label>
        Cluster
        <select
          value={activeCluster}
          onChange={(event) => setActiveCluster(event.target.value)}
        >
          <option value="default">default</option>
          <option value="dev">dev</option>
          <option value="staging">staging</option>
          <option value="prod">prod</option>
        </select>
      </label>
      {loading ? <p>Loading menus...</p> : null}
      {error ? <p>Failed to load menus: {error}</p> : null}
      <ul aria-label="Menu Items">
        {menus.map((menu) => (
          <li key={menu.id}>{menu.title}</li>
        ))}
      </ul>
    </main>
  );
}

export default App;
