postgres:
	docker run --name postgres14-auxstream -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:14-alpine

createdb:
	docker exec -it postgres14-auxstream createdb --username=root --owner=root auxstreamdb

init-migration-schema:
	go run cmd/migration/main.go init

setup-db:
	go run cmd/migration/main.go up

rollback-db:
	go run cmd/migration/main.go down

migration-history:
	go run cmd/migration/main.go history

migration-status:
	go run cmd/migration/main.go status

test:
	go test -v ./tests/... -coverpkg=./...

run:
	go run cmd/server/main.go


build:
	go build -o build/auxstream cmd/server/main.go

build-worker:
	go build -o build/index_worker cmd/workers/index_worker.go

build-all: build build-worker

build-frontend:
	cd interface && npm ci && npm run build

run-worker:
	./build/index_worker -interval 24

run-worker-once:
	./build/index_worker -once

# --- Docker full stack (api, worker, migrate, postgres, redis) via deploy/ ---
docker-up:
	$(MAKE) -C deploy dev

docker-down:
	$(MAKE) -C deploy down

docker-logs:
	$(MAKE) -C deploy logs

# Production backend deploy on the VPS: pull the prebuilt GHCR image and restart
# (replaces the old `systemctl restart`; the box runs containers now).
deploy-backend:
	$(MAKE) -C deploy deploy

# Frontend is still served by host nginx: build the SPA and reload nginx.
deploy-frontend: build-frontend
	sudo systemctl reload nginx

.PHONY: test run createdb setup-db teardown-db rollback-db init-migration-schema migration-history migration-status build build-worker build-all build-frontend run-worker run-worker-once docker-up docker-down docker-logs deploy-backend deploy-frontend
