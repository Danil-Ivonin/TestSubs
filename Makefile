APP := subscriptions
MIGRATE_IMAGE := migrate/migrate:4
MIGRATE_DSN := postgres://$${POSTGRES_USER:-postgres}:$${POSTGRES_PASSWORD:-postgres}@postgres:5432/$${POSTGRES_DB:-wallet_test}?sslmode=$${POSTGRES_SSLMODE:-disable}
K6_IMAGE := grafana/k6:0.54.0
K6_SCRIPT ?= tests/load/subscriptions.js
K6_SCRIPTS_DIR := $(subst \,/,$(CURDIR))/tests/load
K6_BASE_URL ?= http://host.docker.internal:8080
K6_LOCAL_BASE_URL ?= http://localhost:8080
K6_RATE ?= 10000
K6_DURATION ?= 30s
K6_VUS ?= 300
K6_MAX_VUS ?= 1000
.PHONY: run test build docker-up docker-down docker-build migrate-up migrate-down load-test load-test-local

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

load-test:
	docker run --rm --add-host=host.docker.internal:host-gateway -e BASE_URL=$(K6_BASE_URL) -e RATE=$(K6_RATE) -e DURATION=$(K6_DURATION) -e VUS=$(K6_VUS) -e MAX_VUS=$(K6_MAX_VUS) -v "$(K6_SCRIPTS_DIR):/scripts:ro" $(K6_IMAGE) run /scripts/subscriptions.js

load-test-local:
	k6 run -e BASE_URL=$(K6_LOCAL_BASE_URL) -e RATE=$(K6_RATE) -e DURATION=$(K6_DURATION) -e VUS=$(K6_VUS) -e MAX_VUS=$(K6_MAX_VUS) $(K6_SCRIPT)
