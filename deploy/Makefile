COMPOSE = docker compose -f docker-compose.yml --env-file .env

# Local dev: build for THIS machine and run the full stack.
dev: ; $(COMPOSE) up -d --build

# VPS: pull the prebuilt GHCR image and (re)start. No build on the box.
deploy: ; $(COMPOSE) pull && $(COMPOSE) up -d && docker image prune -f

# Migrations. `migrate` runs the service default (init || true; up); the others
# override the command to run a specific migration subcommand against the stack.
migrate:         ; $(COMPOSE) run --rm migrate
migrate-status:  ; $(COMPOSE) run --rm migrate ./migration status
migrate-history: ; $(COMPOSE) run --rm migrate ./migration history
migrate-down:    ; $(COMPOSE) run --rm migrate ./migration down

logs:   ; $(COMPOSE) logs -f api worker
ps:     ; $(COMPOSE) ps
down:   ; $(COMPOSE) down

# Hot DB backup to the current directory.
backup: ; $(COMPOSE) exec -T postgres pg_dump -U $${POSTGRES_USER:-auxstream} $${POSTGRES_DB:-auxstreamdb} | gzip > backup-$$(date +%Y%m%d-%H%M).sql.gz

.PHONY: dev deploy migrate migrate-status migrate-history migrate-down logs ps down backup
