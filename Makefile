.PHONY: run run-server run-scheduler run-worker build test lint swagger migrate-up migrate-down tool-user-create docker-build docker-up docker-down clean

# Run the server
run: run-server

run-server:
	go run cmd/server/main.go --config config/config.dev.yaml

run-scheduler:
	go run cmd/scheduler/main.go --config config/config.dev.yaml

run-worker:
	go run cmd/worker/main.go --config config/config.dev.yaml

# Build
build:
	go build -o bin/server cmd/server/main.go
	go build -o bin/scheduler cmd/scheduler/main.go
	go build -o bin/worker cmd/worker/main.go

# Test
test:
	go test ./...

# Lint (requires golangci-lint in PATH: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest)
lint:
	golangci-lint run ./...

# Swagger
swagger:
	swag init -g cmd/server/main.go -o docs

# Migrate (uses the project's embedded migrate command)
migrate-up:
	go run ./cmd/migrate --config config/config.dev.yaml up

migrate-down:
	go run ./cmd/migrate --config config/config.dev.yaml down

tool-user-create:
	go run ./cmd/tool user create --config config/config.dev.yaml --username "$${APP_USER_USERNAME}" --email "$${APP_USER_EMAIL}" --password "$${APP_USER_PASSWORD}"

# Docker
docker-build:
	docker compose --profile full build

docker-up:
	docker compose --profile full up -d

docker-down:
	docker compose down

# Dev environment (PG + MinIO + Redis only)
dev-up:
	docker compose up -d

dev-down:
	docker compose down

# Clean
clean:
	rm -rf bin/
