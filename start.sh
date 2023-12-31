#!/bin/sh

set -e

echo "run db migration"
/app/migrate -path /app/db/migration -database "$DATABASE_URL" -verbose up

echo "start app"
exec "$@"