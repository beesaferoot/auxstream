import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// API origin that the dev/preview server proxies /api and /health to. In the
// compose network this is the `api` service; override with VITE_API_PROXY for
// other setups (e.g. http://localhost:5009 when running the API on the host).
const apiProxy = process.env.VITE_API_PROXY || 'http://api:5009'

const proxy = {
  '/api': { target: apiProxy, changeOrigin: true },
  '/health': { target: apiProxy, changeOrigin: true },
}

const watch = process.env.VITE_USE_POLLING
  ? { usePolling: true, interval: 120 }
  : undefined

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 3000,
    proxy,
    watch,
  },
  // `vite preview` serves the production build (dist/) and, like the dev server,
  // proxies the API so the SPA stays same-origin. host:true binds 0.0.0.0 so the
  // container is reachable; the proxy re-resolves DNS per request (no stale IPs).
  preview: {
    host: true,
    port: 8080,
    proxy,
  },
})
