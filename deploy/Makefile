COMPOSE = docker compose -f docker-compose.yml --env-file .env

# Local dev: build for THIS machine and run the full stack.
dev: ; $(COMPOSE) up -d --build

# VPS: pull the prebuilt GHCR image and (re)start. No build on the box.
deploy: ; $(COMPOSE) pull && $(COMPOSE) up -d && docker image prune -f

# Run migrations on demand (also runs automatically as the one-shot service).
migrate: ; $(COMPOSE) run --rm migrate

logs:   ; $(COMPOSE) logs -f api worker
ps:     ; $(COMPOSE) ps
down:   ; $(COMPOSE) down

# Hot DB backup to the current directory.
backup: ; $(COMPOSE) exec -T postgres pg_dump -U $${POSTGRES_USER:-auxstream} $${POSTGRES_DB:-auxstreamdb} | gzip > backup-$$(date +%Y%m%d-%H%M).sql.gz

.PHONY: dev deploy migrate logs ps down backup
