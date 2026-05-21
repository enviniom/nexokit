.PHONY: dev build test test-unit test-integration test-coverage migrate-up migrate-down migrate-create migrate-status migrate-reset seed create-root lint fmt vet install-hooks uninstall-hooks check-env

# Load .env if present for Makefile variable expansion.
# Export only the variables the app/CLI reads, avoiding accidental env bleed.
-include .env

export APP_NAME APP_ENV APP_URL APP_PORT
export DATABASE_URL DB_HOST DB_PORT DB_NAME DB_USER DB_PASSWORD DB_SSL_MODE DB_MAX_OPEN_CONNS DB_MAX_IDLE_CONNS DB_CONN_MAX_LIFETIME_SECONDS
export CORS_ALLOWED_ORIGINS
export LOG_LEVEL LOG_FORMAT LOG_FILE LOG_MAX_SIZE LOG_MAX_BACKUPS LOG_MAX_AGE LOG_COMPRESS LOG_GIN_FILE LOG_ERROR_FILE
export SHUTDOWN_TIMEOUT_SECONDS
export CACHE_DRIVER

dev:
	go run ./cmd/nexokit serve

build:
	go build -o bin/api ./cmd/api
	go build -o bin/nexokit ./cmd/nexokit

test:
	go test ./...

test-unit:
	go test -short $(shell go list ./... | grep -v '/tests/integration$$')

test-integration:
	go test ./tests/integration/...

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

migrate-up:
	@if [ -z "$(DATABASE_URL)" ] && [ -z "$(DB_HOST)" ]; then \
		echo "error: database configuration missing. Set DATABASE_URL or DB_* variables."; \
		exit 1; \
	fi
	go run ./cmd/nexokit migrate up

migrate-down:
	@if [ -z "$(DATABASE_URL)" ] && [ -z "$(DB_HOST)" ]; then \
		echo "error: database configuration missing. Set DATABASE_URL or DB_* variables."; \
		exit 1; \
	fi
	go run ./cmd/nexokit migrate down

migrate-create:
	@read -p "Migration name: " name; \
	go run ./cmd/nexokit migrate create $$name

migrate-status:
	@if [ -z "$(DATABASE_URL)" ] && [ -z "$(DB_HOST)" ]; then \
		echo "error: database configuration missing. Set DATABASE_URL or DB_* variables."; \
		exit 1; \
	fi
	go run ./cmd/nexokit migrate status

migrate-reset:
	@if [ -z "$(DATABASE_URL)" ] && [ -z "$(DB_HOST)" ]; then \
		echo "error: database configuration missing. Set DATABASE_URL or DB_* variables."; \
		exit 1; \
	fi
	go run ./cmd/nexokit migrate reset

seed:
	go run ./cmd/nexokit seed

create-root:
	go run ./cmd/nexokit create-root

lint: vet

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
