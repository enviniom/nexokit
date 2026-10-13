.PHONY: dev init build test test-unit test-integration test-coverage migrate-up migrate-down migrate-create migrate-status migrate-reset seed create-root lint fmt vet install-hooks uninstall-hooks check-env check-module-errors

# The application loads .env as dotenv data; Make never parses or exports it.
# Externally supplied environment variables are inherited normally.

dev:
	go run ./cmd/nexokit serve

init:
	go run ./cmd/nexokit init

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
	go run ./cmd/nexokit migrate up

migrate-down:
	go run ./cmd/nexokit migrate down

migrate-create:
	@read -p "Migration name: " name; \
	go run ./cmd/nexokit migrate create $$name

migrate-status:
	go run ./cmd/nexokit migrate status

migrate-reset:
	go run ./cmd/nexokit migrate reset

seed:
	go run ./cmd/nexokit seed

create-root:
	go run ./cmd/nexokit create-root

lint: vet check-module-errors

fmt:
	go fmt ./...

vet:
	go vet ./...

# check-module-errors enforces the change-24 boundary contract:
# module handlers, services, and repositories must not import platform/apperror
# or gorm directly; legacy mapServiceError adapters are forbidden in non-test code.
check-module-errors:
	@echo "Checking for apperror. misuse in module handlers/services/repositories..."
	@if grep -RE 'apperror\.' internal/modules/ --include='*service.go' --include='*repository.go' --include='*handler.go' | grep -v _test.go; then \
		echo "error: apperror. found in module handler/service/repository files"; \
		exit 1; \
	fi
	@echo "Checking for gorm. leaks in module services..."
	@if grep -RE 'gorm\.' internal/modules/ --include='*service.go' | grep -v _test.go; then \
		echo "error: gorm. found in module service files"; \
		exit 1; \
	fi
	@echo "Checking for legacy mapServiceError in modules..."
	@if grep -RE 'mapServiceError' internal/modules/ | grep -v _test.go; then \
		echo "error: mapServiceError found in module files"; \
		exit 1; \
	fi
	@echo "Module error guard passed."

install-hooks:
	@cp scripts/pre-commit.sh .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed."

uninstall-hooks:
	@rm -f .git/hooks/pre-commit
	@echo "Pre-commit hook removed."

check-env:
	@bash scripts/pre-commit.sh --check-env-only
