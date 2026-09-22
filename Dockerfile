FROM golang:1.26.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN set -eu; for command in server scheduler worker all migrate tool; do \
      CGO_ENABLED=0 GOOS=linux go build \
        -trimpath -ldflags="-s -w" \
        -o "/app/bin/${command}" "./cmd/${command}"; \
    done

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin /app/bin
COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/scripts/docker-entrypoint.sh /app/docker-entrypoint.sh

RUN chmod +x /app/docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["server", "--config", "config/config.prod.yaml"]
