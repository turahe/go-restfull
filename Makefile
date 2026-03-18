.PHONY: help tidy test swagger run db-up db-down docker-up docker-down

help:
	@echo "Targets:"
	@echo "  tidy        - go mod tidy"
	@echo "  test        - run go test ./..."
	@echo "  swagger     - regenerate swagger docs"
	@echo "  run         - run API locally (uses .env if present)"
	@echo "  db-up       - start MySQL only (docker compose)"
	@echo "  db-down     - stop MySQL only (docker compose)"
	@echo "  docker-up   - start MySQL + API (docker compose)"
	@echo "  docker-down - stop MySQL + API (docker compose)"

tidy:
	go mod tidy

test:
	go test ./...

swagger:
	swag init -g cmd/main.go -o docs

run:
	go run cmd/main.go

db-up:
	docker compose up -d mysql

db-down:
	docker compose down

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

