APP := subscriptions
MIGRATE_IMAGE := migrate/migrate:4
MIGRATE_DSN := postgres://$${POSTGRES_USER:-postgres}:$${POSTGRES_PASSWORD:-postgres}@postgres:5432/$${POSTGRES_DB:-wallet_test}?sslmode=$${POSTGRES_SSLMODE:-disable}

.PHONY: run test build docker-up docker-down docker-build migrate-up migrate-down

run:
	go run ./cmd/$(APP)

test:
	go test ./...

build:
	go build -o bin/$(APP) ./cmd/$(APP)

docker-up:
	docker compose up --build

docker-down:
	docker compose down

docker-build:
	docker compose build

migrate-up:
	docker compose run --rm migrate -path=/migrations -database "$(MIGRATE_DSN)" up

migrate-down:
	docker compose run --rm migrate -path=/migrations -database "$(MIGRATE_DSN)" down 1
