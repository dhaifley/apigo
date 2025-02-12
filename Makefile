VERSION="0.1.1"

GO_FILES := $(shell find . -name "*.go")

YAML_FILES := $(shell find ./api -name "*.yaml")

include ./tests/test.env

all: build

clean:
	rm -f apigo
.PHONY: clean

internal/static/openapi.yaml: $(YAML_FILES)
	@./api/generate.sh

docs: internal/static/openapi.yaml
.PHONY: docs

apigo: $(GO_FILES) Dockerfile tests/docker-compose.yml internal/static/*
	CGO_ENABLED=0 go build -v -o apigo \
	-ldflags="-X github.com/dhaifley/apigo/internal/server.Version=${VERSION}" \
	./cmd/apigo

build: apigo
.PHONY: build

build-docker: build Dockerfile tests/docker-compose.yml
	docker compose -f tests/docker-compose.yml build
.PHONY: build-docker

start: build build-docker
	docker compose -f tests/docker-compose.yml up -d --force-recreate
	@echo "Test services started."
.PHONY: start

stop:
	docker compose -f tests/docker-compose.yml down --remove-orphans --volumes
	@echo "All test services stopped."
.PHONY: stop

test:
	@make start
	go test -race -cover ./...
	@make stop
.PHONY: test

test-quick:
	go test -race -cover -short ./...
.PHONY: test-quick

run: build
	@echo "set -a && . ./tests/test.env && ./apigo" | ${SHELL}
.PHONY: run
