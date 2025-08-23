GO ?= go
BINARY ?= bin/cobra-template
PKG := ./...

.PHONY: help fmt tidy build run up down logs test clean migrate-up migrate-status migrate-create


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
	docker-compose down -v

logs:
	docker-compose logs -f

test:
	$(GO) test $(PKG)

clean:
	rm -rf bin

migrate-up: build
	$(BINARY) migrate up

migrate-status: build
	$(BINARY) migrate status

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi; \
	TIMESTAMP=$$(date +%Y%m%d%H%M%S); \
	FILE="migrations/$${TIMESTAMP}_$(NAME).sql"; \
	echo "-- Migration: $${TIMESTAMP}_$(NAME)" > $$FILE; \
	echo "-- Description: $(NAME)" >> $$FILE; \
	echo "-- Created: $$(date +%Y-%m-%d)" >> $$FILE; \
	echo "" >> $$FILE; \
	echo "-- Add your migration SQL here" >> $$FILE; \
	echo "Created migration file: $$FILE"
