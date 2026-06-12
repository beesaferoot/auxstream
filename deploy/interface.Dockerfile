# syntax=docker/dockerfile:1

# Stage 1 — build the Vite/React SPA
FROM node:20-alpine AS web-builder
WORKDIR /web
COPY interface/package.json interface/package-lock.json ./
RUN npm ci
COPY interface/ .
# Same-origin API base: requests go to /api/v1/* and the in-container nginx (stage 2)
# proxies them to the api service. Overridable at build time.
ARG VITE_BASE_URL=/api/v1
ENV VITE_BASE_URL=$VITE_BASE_URL
RUN npm run build        # -> /web/dist

# Stage 2 — serve the static build + proxy /api to the api service
FROM nginx:1.27-alpine
COPY deploy/interface.nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=web-builder /web/dist /usr/share/nginx/html
EXPOSE 80
