DB_HOST ?= localhost
DB_PORT ?= 55432
DB_USER ?= migranthub
DB_PASSWORD ?= migranthub
DB_NAME ?= migranthub
DB_SSLMODE ?= disable

DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

MIGRATIONS_DIR := migrations
MIGRATE_VERSION := v4.20.1
MIGRATE := go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

.PHONY: run test build fmt vet tidy migrate-up migrate-down migrate-create migrate-version

run:
	go run ./services/api/cmd/server

test:
	go test ./...

build:
	go build -o bin/api ./services/api/cmd/server

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy

migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-create:
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-version:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version