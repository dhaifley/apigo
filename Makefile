VERSION="0.1.1"

include ./tests/test.env

all: build

clean:
	rm -f apigo
.PHONY: clean

apigo:
	docker compose -f tests/docker-compose.yml build
	CGO_ENABLED=0 go build -v -o apigo \
	-ldflags="-X github.com/dhaifley/apigo/server.Version=${VERSION}" \
	./cmd/apigo

build: apigo
.PHONY: build

test-start: build
	docker compose -f tests/docker-compose.yml up -d --force-recreate
	@echo "Test services started."
.PHONY: test-start

test-stop:
	docker compose -f tests/docker-compose.yml down --remove-orphans --volumes
	@echo "All test services stopped."
.PHONY: test-stop

test:
	@make test-start
	go test -race -cover ./...
	@make test-stop
.PHONY: test

test-quick:
	go test -race -cover -short ./...
.PHONY: test-quick

run: build
	@echo "set -a && . ./tests/test.env && ./apigo" | ${SHELL}
.PHONY: run
