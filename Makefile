APP := subscriptions

.PHONY: run test build

run:
	go run ./cmd/$(APP)

test:
	go test ./...

build:
	go build -o bin/$(APP) ./cmd/$(APP)
