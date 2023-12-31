# Build stage
FROM golang:1.21-alpine3.19 as builder
WORKDIR /app
COPY . .

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

RUN go build -o auxstream

# Run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/auxstream .
COPY --from=builder /go/bin/migrate ./migrate
COPY app.env .
COPY start.sh .
COPY db/migration ./db/migration

EXPOSE 5009

CMD ["/app/auxstream"]
ENTRYPOINT ["/app/start.sh"]