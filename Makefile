GO_PACKAGES := ./...
PYTEST := cd ai_runtime && python3 -m pytest -c pyproject.toml ../tests/ai_runtime
DOCKER_COMPOSE := docker-compose
DOCKER_BUILDKIT ?= 0

.PHONY: fmt lint test test-go test-python run-api up down build-image check-compose

fmt:
	gofmt -w cmd internal

lint:
	golangci-lint run
	$(PYTEST)

test: test-go test-python

test-go:
	go test $(GO_PACKAGES)

test-integration:
	AUTONOMA_DATABASE_URL=postgres://autonoma:autonoma@127.0.0.1:5432/autonoma?sslmode=disable go test ./tests/integration

test-python:
	$(PYTEST)

run-api:
	go run ./cmd/api

up:
	DOCKER_BUILDKIT=$(DOCKER_BUILDKIT) $(DOCKER_COMPOSE) up --build

down:
	DOCKER_BUILDKIT=$(DOCKER_BUILDKIT) $(DOCKER_COMPOSE) down

build-image:
	DOCKER_BUILDKIT=$(DOCKER_BUILDKIT) docker build -t autonoma-api:phase0 .

check-compose:
	$(DOCKER_COMPOSE) config >/dev/null