COMPOSE = docker compose -f docker-compose.yml --env-file .env
# Same stack, plus the frontend hot-reload override (Vite dev server + HMR).
COMPOSE_HOT = docker compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.dev.yml --env-file .env

# Local dev: build for THIS machine and run the full stack.
dev: ; $(COMPOSE) up --build

# Local dev with frontend hot-reload: api/worker/etc. run as usual, but `interface`
# runs the Vite dev server against live-mounted source — edit files in interface/
# and the browser updates with no rebuild. Compose builds any missing images on the
# first run; later runs reuse them.
dev-hot: ; $(COMPOSE_HOT) up

# VPS: pull the prebuilt GHCR images and (re)start. --no-build guarantees the box
# never compiles — it only runs what CI pushed to GHCR.
deploy: ; $(COMPOSE) pull && $(COMPOSE) up -d --no-build && docker image prune -f

# Migrations. `migrate` runs the service default (init || true; up); the others
# override the command to run a specific migration subcommand against the stack.
migrate:         ; $(COMPOSE) run --rm migrate
migrate-status:  ; $(COMPOSE) run --rm migrate ./migration status
migrate-history: ; $(COMPOSE) run --rm migrate ./migration history
migrate-down:    ; $(COMPOSE) run --rm migrate ./migration down

logs:   ; $(COMPOSE) logs -f api worker interface
ps:     ; $(COMPOSE) ps
down:   ; $(COMPOSE) down

# Hot DB backup to the current directory.
backup: ; $(COMPOSE) exec -T postgres pg_dump -U $${POSTGRES_USER:-auxstream} $${POSTGRES_DB:-auxstreamdb} | gzip > backup-$$(date +%Y%m%d-%H%M).sql.gz

.PHONY: dev dev-hot deploy migrate migrate-status migrate-history migrate-down logs ps down backup
