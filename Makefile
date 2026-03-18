.PHONY: help tidy test swagger run serve seed-rbac db-up db-down docker-up docker-down docker-build docker-logs

help:
	@echo "Targets:"
	@echo "  tidy        - go mod tidy"
	@echo "  test        - run go test ./..."
	@echo "  swagger     - regenerate swagger docs"
	@echo "  run         - run API locally (default: serve)"
	@echo "  serve       - run API locally"
	@echo "  seed-rbac   - seed default RBAC (roles/permissions + casbin)"
	@echo "  db-up       - start MySQL only (docker compose)"
	@echo "  db-down     - stop MySQL only (docker compose)"
	@echo "  docker-up   - start MySQL + API (docker compose)"
	@echo "  docker-down - stop MySQL + API (docker compose)"
	@echo "  docker-build- rebuild API image"
	@echo "  docker-logs - tail API logs"

tidy:
	go mod tidy

test:
	go test ./...

swagger:
	swag init -g cmd/main.go -o docs

run:
	go run ./cmd

serve:
	go run ./cmd serve

seed-rbac:
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

