.PHONY: help tidy test test-unit test-cover test-ci swagger run serve seed-rbac env local-serve local-seed-rbac local-deps docker-up docker-down docker-build docker-logs db-up db-down

help:
	@echo "Targets:"
	@echo "  tidy        - go mod tidy"
	@echo "  test        - run go test ./..."
	@echo "  test-unit   - run unit tests (no cache, verbose)"
	@echo "  test-cover  - run tests with coverage report"
	@echo "  test-ci     - tidy + test (CI-like)"
	@echo "  swagger     - regenerate swagger docs"
	@echo "  env         - create .env from .env.example (if missing)"
	@echo ""
	@echo "Local (no Docker):"
	@echo "  local-deps  - print required local dependencies"
	@echo "  local-serve - run API locally (go run ./cmd serve)"
	@echo "  local-seed-rbac - seed default RBAC (go run ./cmd seed rbac)"
	@echo ""
	@echo "Docker:"
	@echo "  db-up       - start MySQL only (docker compose)"
	@echo "  db-down     - stop MySQL only (docker compose)"
	@echo "  docker-up   - start MySQL + Redis + MinIO + API (docker compose)"
	@echo "  docker-down - stop docker compose stack"
	@echo "  docker-build- rebuild API image"
	@echo "  docker-logs - tail API logs"

tidy:
	go mod tidy

test:
	go clean -testcache
	go test ./...

test-unit:
	go test -count=1 -v ./...

test-cover:
	go test -count=1 ./... -coverprofile=tmp/coverage.out
	go tool cover -func=tmp/coverage.out

test-ci: tidy test

swagger:
	# Swag scans the provided directories for annotations; this avoids running `go list ./` in
	# the module root (which contains no Go files).
	# When `-d` is set, `-g/--generalInfo` must be relative to the first directory in `-d`.
	# `internal/` has no top-level Go files, only subpackages, so scan only concrete package dirs
	# that contain .go files (avoid `./internal` itself, which triggers `go list` warnings).
	swag init -g main.go -o docs -d ./cmd,./internal/config,./internal/database,./internal/handler,./internal/middleware,./internal/model,./internal/rbac,./internal/repository,./internal/seeder,./internal/service,./pkg/logger,./pkg/response

run:
	go run ./cmd

serve:
	go run ./cmd serve

seed-rbac:
	go run ./cmd seed rbac

env:
	@powershell -NoProfile -Command "if (!(Test-Path .env)) { Copy-Item .env.example .env; Write-Host 'Created .env from .env.example' } else { Write-Host '.env already exists' }"

local-deps:
	@echo "To run without Docker, you need these running locally:"
	@echo "  - MySQL 8.x (DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME)"
	@echo "  - Redis 7.x (REDIS_ADDR/REDIS_PASSWORD/REDIS_DB)"
	@echo "Optional:"
	@echo "  - MinIO (MINIO_ENDPOINT/MINIO_ACCESS_KEY/MINIO_SECRET_KEY) or leave unset to disable"
	@echo ""
	@echo "Then run:"
	@echo "  make env"
	@echo "  make local-serve"

local-serve: env
	go run ./cmd serve

local-seed-rbac:
	go run ./cmd seed rbac

db-up:
	docker compose up -d mysql

db-down:
	docker compose down

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-build:
	docker compose build api

docker-logs:
	docker compose logs -f --tail 100 api

