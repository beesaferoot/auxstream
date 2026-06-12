# syntax=docker/dockerfile:1

# Stage 1 — build all three binaries from the same module
FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
# Pure-Go (no cgo): gorm/postgres + uuid are pure Go, so binaries are static and
# run on a minimal base with just ca-certificates.
ENV CGO_ENABLED=0 GOOS=linux
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -o /out/auxstream    ./cmd/server     && \
    go build -trimpath -o /out/index_worker ./cmd/workers    && \
    go build -trimpath -o /out/migration    ./cmd/migration

# Stage 2 — minimal runtime
FROM alpine:3.20
# ca-certificates: outbound HTTPS to Cloudinary/YouTube/SoundCloud/Google OAuth.
# wget: used by the compose healthcheck against /health.
RUN apk add --no-cache ca-certificates wget && adduser -D -u 10001 app
WORKDIR /app
COPY --from=builder /out/auxstream    ./auxstream
COPY --from=builder /out/index_worker ./index_worker
COPY --from=builder /out/migration    ./migration
RUN chown -R app:app /app
USER app
EXPOSE 5009
# Default command is the API; worker/migrate override `command` in compose.
CMD ["./auxstream"]
