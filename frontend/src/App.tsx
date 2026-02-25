import { useEffect, useState } from 'react';

interface MenuItem {
  id: string;
  title: string;
}

interface MenusResponse {
  menus: MenuItem[];
}

function App() {
  const [menus, setMenus] = useState<MenuItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;

    async function loadMenus() {
      try {
        const response = await fetch('/api/meta/menus?cluster=default');
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
  }, []);

  return (
    <main>
      <h1>KubeDeck</h1>
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
