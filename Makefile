.PHONY: build run test migrate-up migrate-down migrate-create migrate-status fmt vet

build:
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

test:
	go test ./...

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-create:
	@read -p "Migration name: " name; \
	goose -dir migrations create $$name sql

migrate-status:
	goose -dir migrations postgres "$(DATABASE_URL)" status

fmt:
	go fmt ./...

vet:
	go vet ./...
