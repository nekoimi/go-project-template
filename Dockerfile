FROM golang:1.26.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
RUN set -eu; for command in server scheduler worker all migrate tool version; do \
      CGO_ENABLED=0 GOOS=linux go build \
        -trimpath -ldflags="-s -w \
          -X github.com/nekoimi/go-project-template/internal/buildinfo.Version=${VERSION} \
          -X github.com/nekoimi/go-project-template/internal/buildinfo.Commit=${COMMIT} \
          -X github.com/nekoimi/go-project-template/internal/buildinfo.BuildTime=${BUILD_TIME}" \
        -o "/app/bin/${command}" "./cmd/${command}"; \
    done

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=builder /app/bin /app/bin
COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/scripts/docker-entrypoint.sh /app/docker-entrypoint.sh

RUN chmod +x /app/docker-entrypoint.sh

RUN mkdir -p /app/uploads \
    && chown -R app:app /app

USER app

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["server", "--config", "config/config.prod.yaml"]
