# Local development and CI.
# The full stack (api, worker, migrate, postgres, redis) runs via Docker — see deploy/.
# Run `make docker-up` to bring everything up locally; migrations apply automatically.

# Go test suite (uses sqlmock + miniredis — no real services needed).
test:
	go test -v ./tests/... -coverpkg=./...

# Build the Vite/React SPA (served by host nginx in production).
build-frontend:
	cd interface && npm ci && npm run build

# --- Docker: local full stack (api, worker, migrate, postgres, redis) via deploy/ ---
docker-up:
	$(MAKE) -C deploy dev

# Like docker-up, but the frontend runs the Vite dev server with hot-reload (HMR):
# edit anything in interface/ and the browser updates instantly — no image rebuild.
docker-up-hot:
	$(MAKE) -C deploy dev-hot

docker-down:
	$(MAKE) -C deploy down

docker-logs:
	$(MAKE) -C deploy logs

# Migrations apply automatically on `docker-up`; these run manual ops against the stack.
migrate:
	$(MAKE) -C deploy migrate

migrate-status:
	$(MAKE) -C deploy migrate-status

migrate-history:
	$(MAKE) -C deploy migrate-history

migrate-down:
	$(MAKE) -C deploy migrate-down

# --- Deploy ---
# Backend: pull the prebuilt GHCR image and restart containers on the VPS.
deploy-backend:
	$(MAKE) -C deploy deploy

# Frontend: build the SPA and reload host nginx.
deploy-frontend: build-frontend
	sudo systemctl reload nginx

.PHONY: test build-frontend docker-up docker-up-hot docker-down docker-logs migrate migrate-status migrate-history migrate-down deploy-backend deploy-frontend
