.PHONY: run run-server run-scheduler build test migrate-up migrate-down docker-build docker-up docker-down clean

# Run the server
run: run-server

run-server:
	go run cmd/server/main.go --config configs/config.dev.yaml

run-scheduler:
	go run cmd/scheduler/main.go --config configs/config.dev.yaml

# Build
build:
	go build -o bin/server cmd/server/main.go
	go build -o bin/scheduler cmd/scheduler/main.go

# Test
test:
	go test ./...

# Migrate (requires golang-migrate CLI)
migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/go_template?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/go_template?sslmode=disable" down

# Docker
docker-build:
	docker compose -f deployments/docker-compose.yml build

docker-up:
	docker compose -f deployments/docker-compose.yml up -d

docker-down:
	docker compose -f deployments/docker-compose.yml down

# Dev environment (PG + MinIO only)
dev-up:
	docker compose up -d

dev-down:
	docker compose down

# Clean
clean:
	rm -rf bin/
