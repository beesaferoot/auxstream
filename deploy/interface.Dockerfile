# syntax=docker/dockerfile:1

# Build the Vite/React SPA and serve it with `vite preview` — Vite's static
# preview server. It serves dist/ with SPA history fallback and proxies /api +
# /health to the api service (see interface/vite.config.ts). Node is kept in the
# runtime image because preview is a Node process (heavier than a static nginx,
# but simpler — no separate web server config).
FROM node:20-alpine
WORKDIR /web

COPY interface/package.json interface/package-lock.json ./
RUN npm ci

COPY interface/ .
# Same-origin API base: requests go to /api/v1/* and the preview proxy forwards
# them to the api service. Overridable at build time.
ARG VITE_BASE_URL=/api/v1
ENV VITE_BASE_URL=$VITE_BASE_URL
RUN npm run build        # -> /web/dist

# Where the preview proxy sends /api + /health: the api service on the compose
# network; overridable at runtime (compose sets this explicitly).
ENV VITE_API_PROXY=http://api:5009

EXPOSE 8080
# --host binds 0.0.0.0 (also set in vite.config); port matches preview.port.
CMD ["npm", "run", "preview", "--", "--host", "--port", "8080"]
