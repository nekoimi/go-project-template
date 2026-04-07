FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGET=cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./${TARGET}

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/configs /app/configs

EXPOSE 8080

ENTRYPOINT ["/app/server"]
CMD ["--config", "configs/config.prod.yaml"]
