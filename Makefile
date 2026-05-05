.PHONY: build run test test-race lint vet fmt mocks tidy migrate-up migrate-down up down

MIGRATE = migrate -path migrations -database "mysql://$(DB_DSN)"

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -count=1 ./internal/... ./pkg/...

test-race:
	go test -race -count=1 -timeout 120s ./internal/... ./pkg/...

test-integration:
	go test -race -count=1 -timeout 300s ./integration_test/...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

mocks:
	go generate ./internal/mocks/...

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

up:
	docker compose -f deploy/docker-compose.yml up -d

down:
	docker compose -f deploy/docker-compose.yml down
