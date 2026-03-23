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
        // Test mocks for xterm (only used in tests)
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
    build: {
      // Code splitting optimization
      rollupOptions: {
        output: {
          manualChunks: {
            // Vendor chunks
            'vendor-react': ['react', 'react-dom', 'react-router-dom'],
            'vendor-mui': ['@mui/material', '@mui/icons-material', '@mui/system'],
            'vendor-xterm': ['@xterm/xterm', '@xterm/addon-fit', '@xterm/addon-attach'],
          },
        },
      },
      // Chunk size warning limit
      chunkSizeWarningLimit: 500,
      // Minification
      minify: 'terser',
      terserOptions: {
        compress: {
          drop_console: true,
          drop_debugger: true,
        },
      },
      // Source maps
      sourcemap: false,
      // Target modern browsers
      target: 'esnext',
    },
  };
});
