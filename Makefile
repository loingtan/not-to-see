GO ?= go
BINARY ?= bin/cobra-template
PKG := ./...

.PHONY: help fmt tidy build run up down logs test clean

help:
	@echo "Common targets:"
	@echo "  fmt     - Format code with gofmt"
	@echo "  tidy    - Sync go.mod/go.sum"
	@echo "  build   - Build CLI binary to $(BINARY)"
	@echo "  run     - Build then run registration server"
	@echo "  up      - Start Postgres and Redis with docker compose"
	@echo "  down    - Stop docker compose services"
	@echo "  logs    - Tail compose service logs"
	@echo "  test    - Run unit tests"
	@echo "  clean   - Remove build artifacts"

fmt:
	gofmt -s -w .

tidy:
	$(GO) mod tidy

build: fmt tidy
	$(GO) build -o $(BINARY)

run: build
	$(BINARY) registration -v

up:
	docker-compose up -d postgres redis

down:
	docker-compose down

logs:
	docker-compose logs -f

test:
	$(GO) test $(PKG)

clean:
	rm -rf bin
