#!make
include .env
LINTER=golangci-lint

deploy: lint docker-compose.up migrate.up
	@echo "----- deploy -----"

DB_CONNECTION="host=$(DB_HOST) port=$(DB_PORT) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) sslmode=$(DB_SSL_MODE)"
MIGRATIONS_FOLDER="migrations"
SQLC_FOLDER="pkg/repository"

.PHONY: docker-compose.up
docker-compose.up: 
	@echo "----- deploy by docker -----"
	@docker-compose up -d


.PHONY: docker-compose.down
docker-compose.down:
	docker-compose down

.PHONY: migrate.up
migrate.up:
	@echo "----- running migrations up -----"
	@cd $(MIGRATIONS_FOLDER);\
	goose postgres ${DB_CONNECTION} up


.PHONY: migrate.down
migrate.down:
	@cd $(MIGRATIONS_FOLDER);\
	goose postgres ${DB_CONNECTION} down


.PHONY: migrate.create
migrate.create:
	@cd $(MIGRATIONS_FOLDER);\
	goose create $(name) sql

.PHONY: migrate.reset
migrate.reset:
	@cd $(MIGRATIONS_FOLDER);\
	goose postgres ${DB_CONNECTION} reset

.PHONY: gen
gen: gen.sqlc gen.api

.PHONY: gen.sqlc
gen.sqlc:
	@echo "----------- Generate sqlc ----------------"
	@sqlc generate

.PHONY: gen.api
gen.api:
	@echo "----------- Generate apis ----------------"
	@oapi-codegen --config api/auth-config.yml api/openapi.yaml
	@oapi-codegen --config api/timetable-config.yml api/openapi.yaml

.PHONY: lint
lint:
	@echo "----------- Lint project ----------------"
	@$(LINTER) run


.PHONY: docker.build
docker.build:
	docker build -t dyleme/schedudler .

.PHONY: docker.push
docker.push: docker.build
	docker push dyleme/schedudler