import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import { loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '.', '');
  const backendTarget = env.VITE_BACKEND_TARGET || 'http://127.0.0.1:8080';

  return {
    plugins: [react()],
    resolve: {
      alias: {
        '@xterm/xterm': '/src/test/mocks/xterm.ts',
        '@xterm/addon-fit': '/src/test/mocks/xterm-addons.ts',
        '@xterm/addon-attach': '/src/test/mocks/xterm-addons.ts',
        '@xterm/xterm/css/xterm.css': '/src/test/mocks/empty.css',
      },
    },
    server: {
      proxy: {
        '/api': {
          target: backendTarget,
          changeOrigin: true,
        },
      },
    },
    test: {
      environment: 'jsdom',
    },
  };
});
