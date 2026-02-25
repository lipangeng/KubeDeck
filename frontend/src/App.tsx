import { useEffect, useState } from 'react';

interface MenuItem {
  id: string;
  title: string;
}

interface MenusResponse {
  menus: MenuItem[];
}

function resolveMenusEndpoint(cluster: string): string {
  const path = `/api/meta/menus?cluster=${encodeURIComponent(cluster)}`;

  try {
    if (
      window.location.protocol === 'http:' ||
      window.location.protocol === 'https:'
    ) {
      return new URL(path, window.location.origin).toString();
    }
  } catch {
    // Fall through to local backend default below.
  }

  return `http://127.0.0.1:8080${path}`;
}

function App() {
  const [activeCluster, setActiveCluster] = useState('default');
  const [menus, setMenus] = useState<MenuItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;

    async function loadMenus() {
      setLoading(true);
      setError(null);
      try {
        const response = await fetch(resolveMenusEndpoint(activeCluster));
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

  return (
    <main>
      <h1>KubeDeck</h1>
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
