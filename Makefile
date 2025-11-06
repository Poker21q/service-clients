include .env
export $(shell cat .env | grep -v '^#')

TOOLS_DIR := bin
OAPI_CODEGEN := $(TOOLS_DIR)/oapi-codegen
GOFUMPT := $(TOOLS_DIR)/gofumpt
GOOSE := $(TOOLS_DIR)/goose

MIGRATIONS_DIR = ./migrations

all: codegen fmt

codegen: $(OAPI_CODEGEN)
	$(OAPI_CODEGEN) --config api/codegen.config.yaml api/schema.yaml

fmt: $(GOFUMPT)
	$(GOFUMPT) -w .

$(OAPI_CODEGEN):
	go build -o $(OAPI_CODEGEN) github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen

$(GOFUMPT):
	go build -o $(GOFUMPT) mvdan.cc/gofumpt

$(GOOSE):
	go build -o $(GOOSE) github.com/pressly/goose/v3/cmd/goose

install-tools: $(OAPI_CODEGEN) $(GOFUMPT) $(GOOSE)

migrate-up: $(GOOSE)
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) goose -dir $(MIGRATIONS_DIR) up

migrate-down: $(GOOSE)
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) goose -dir $(MIGRATIONS_DIR) down

migrate-status: $(GOOSE)
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) goose -dir $(MIGRATIONS_DIR) status

docker-run-postgres:
	docker run -d \
	  --name postgres-db-container \
	  -p $(POSTGRES_PORT):5432 \
	  -e POSTGRES_DB=$(POSTGRES_DB) \
	  -e POSTGRES_USER=$(POSTGRES_USER) \
	  -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	  -v postgres_ostula_data:/var/lib/postgresql/data \
	  --restart unless-stopped \
	  postgres:15-alpine

.PHONY: all codegen fmt clean install-tools