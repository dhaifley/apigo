VERSION="0.1.1"

GO_FILES := $(shell find . -name "*.go")

YAML_FILES := $(shell find ./api -name "*.yaml")

include ./tests/test.env

all: build

clean:
	rm -f apigo
	rm -r apigo.test
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

docker.test: apigo Dockerfile tests/docker-compose.yml
	docker compose -f tests/docker-compose.yml build
	touch docker.test

build-docker: docker.test
.PHONY: build-docker

certs/tls.key:
	@sh certs/generate.sh

certs/tls.crt: certs/tls.key

build-certs: certs/tls.crt
.PHONY: build-certs

clean-certs:
	@rm -f certs/*.crt certs/*.key certs/*.csr certs/*.srl
.PHONY: clean-certs

start.test: build build-docker build-certs
	docker compose -f tests/docker-compose.yml up -d --force-recreate
	@touch start.test
	@echo "Test services started."

start: start.test
.PHONY: start

stop: clean-certs
	docker compose -f tests/docker-compose.yml down --remove-orphans --volumes
	@rm -f start.test
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

run: build start
	@echo "set -a && . ./tests/test.env && ./apigo" | ${SHELL}
.PHONY: run
