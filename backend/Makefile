-include ../.env.local

.PHONY: openapi-codegen migration-create migration-up migration-down migration-status jet-codegen run tidy fmt help

## Generate Go server + TypeScript client from OpenAPI spec
openapi-codegen:
	go tool oapi-codegen -config api/oapi-server.yml api/openapi.yml
	cd ../frontend && npx @rtk-query/codegen-openapi rtk-query.config.cjs

## Create a new migration file: make migration-create name=<migration_name>
migration-create:
	go tool goose -s -dir db/migrations create $(name) sql

## Run pending migrations then regenerate go-jet models
migration-up:
	go tool goose -dir db/migrations postgres "$(DATABASE_URL)" up
	go tool jet -dsn="$(DATABASE_URL)" -schema=public -path=./generated/db

## Rollback last migration then regenerate go-jet models
migration-down:
	go tool goose -dir db/migrations postgres "$(DATABASE_URL)" down
	go tool jet -dsn="$(DATABASE_URL)" -schema=public -path=./generated/db

jet-codegen:
	go tool jet -dsn="$(DATABASE_URL)" -schema=public -path=./generated/db

## Show migration status
migration-status:
	go tool goose -dir db/migrations postgres "$(DATABASE_URL)" status

## Run the backend server with hot reload (air)
run:
	go tool air

## Run without hot reload
run-once:
	go run main.go

## Format Go code
fmt:
	gofmt -w .

help:
	@echo ""
	@echo "Usage: make <target> [name=<migration_name>]"
	@echo ""
	@echo "Targets:"
	@echo "  openapi-codegen          Generate Go server + TypeScript client from OpenAPI spec"
	@echo "  migration-create         Create a new goose migration file (requires name=<name>)"
	@echo "  migration-up             Run pending migrations + regenerate go-jet models"
	@echo "  migration-down           Rollback last migration + regenerate go-jet models"
	@echo "  migration-status         Show migration status"
	@echo "  jet-codegen              Regenerate go-jet models from current DB schema"
	@echo "  run                      Run the backend server"
	@echo "  tidy                     Run go mod tidy"
	@echo "  fmt                      Format Go code"
	@echo ""
