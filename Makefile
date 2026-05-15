.PHONY: build run test migrate-up migrate-down migrate-create migrate-status fmt vet install-hooks uninstall-hooks check-env

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

install-hooks:
	@cp scripts/pre-commit.sh .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed."

uninstall-hooks:
	@rm -f .git/hooks/pre-commit
	@echo "Pre-commit hook removed."

check-env:
	@bash scripts/pre-commit.sh --check-env-only
