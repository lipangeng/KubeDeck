import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { KernelRuntimeProvider } from './kernel/runtime/KernelRuntimeContext';
import { discoverFrontendPluginModules } from './kernel/runtime/discoverFrontendPluginModules';
import { type ThemePreference } from './themeMode';
import { AppShell } from './AppShell';
import { LoginPage } from './pages/LoginPage';

interface AppProps {
  themePreference: ThemePreference;
  onThemePreferenceChange: (next: ThemePreference) => void;
}

export function App({ themePreference, onThemePreferenceChange }: AppProps) {
  const pluginModules = discoverFrontendPluginModules();

  return (
    <BrowserRouter>
      <KernelRuntimeProvider pluginModules={pluginModules}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/*" element={<AppShell themePreference={themePreference} onThemePreferenceChange={onThemePreferenceChange} />} />
        </Routes>
      </KernelRuntimeProvider>
    </BrowserRouter>
  );
}

export default App;
