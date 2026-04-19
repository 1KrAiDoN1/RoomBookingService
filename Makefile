include .env

DOCKER_COMPOSE=docker-compose
GOLINT=golangci-lint

DB_URL ?= $(DB_URL)

docker-build:
	@$(DOCKER_COMPOSE) build

docker-up:
	@$(DOCKER_COMPOSE) up -d

docker-down:
	@$(DOCKER_COMPOSE) down

docker-logs:
	@$(DOCKER_COMPOSE) logs -f

migrate-up: 
	${DOCKER_COMPOSE} exec app migrate -path /app/migrations -database "${DB_URL}" up

migrate-down:
	${DOCKER_COMPOSE} exec app migrate -path /app/migrations -database "${DB_URL}" down
lint:
	@$(GOLINT) run ./...