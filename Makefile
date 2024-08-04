include .env

PROJECT_NAME = $(PROJECT_NAME)
SRC_DIR = ./
TEST_DIR = $(SRC_DIR)...
GO = @go

# GO commands
GO_BUILD = $(GO) build
GO_TEST = $(GO) test
GO_TIDY = $(GO) mod tidy

# Target

build: 
	$(GO_BUILD) -o bin/$(PROJECT_NAME) -v cmd/main.go

run: build
	bin/$(PROJECT_NAME)
	

test:
	$(GO_TEST) -v $(TEST_DIR)

clean_cached:
	$(GO) clean -testcache

tidy:
	$(GO_TIDY)
create_container:
	$(DOC) run --name $(POSTGRES_DOCKER_CONTAINER) -e POSTGRES_USER=$(POSTGRES_USER_DOCKER_CONTAINER) -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD_DOCKER_CONTAINER) -p $(POSTGRES_DB_PORT_HOST_DOCKER_CONTAINER):$(POSTGRES_DB_PORT_DOCKER_CONTAINER) -d postgres:12-alpine

create_db:
	$(DOC) exec -it $(POSTGRES_DOCKER_CONTAINER) psql -U $(POSTGRES_USER_DOCKER_CONTAINER) -c "CREATE DATABASE $(POSTGRES_DB_NAME_DOCKER_CONTAINER)"

start_container:
	$(DOC) start $(POSTGRES_DOCKER_CONTAINER)