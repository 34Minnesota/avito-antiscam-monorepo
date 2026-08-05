include .env
export

export PROJECT_ROOT=$(shell pwd)

test:
	@cd backend && go test ./... -cover

lint:
	@cd backend && golangci-lint run ./... --fix

db-up:
	@docker compose up -d postgres

db-down:
	@docker compose down postgres

ps:
	@docker compose ps

run:
	@cd backend && go run ./cmd/antiscam

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "No seq parameter found. Usage: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	MSYS_NO_PATHCONV=1 docker compose run --rm postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@MSYS_NO_PATHCONV=1 docker compose run --rm postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable \
		up

migrate-down:
	@MSYS_NO_PATHCONV=1 docker compose run --rm postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable \
		down

