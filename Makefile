GO ?= go
BINARY ?= bin/cobra-template
PKG := ./...

.PHONY: help fmt tidy build run up down logs test clean migrate-up migrate-status migrate-create redis-up redis-down redis-status redis-test redis-logs reset-db reset-db-quick

help:
	@echo "Course Registration System - Available Commands"
	@echo "=============================================="
	@echo ""
	@echo "🔧 Development:"
	@echo "  fmt          Format Go code"
	@echo "  tidy         Tidy Go modules"
	@echo "  build        Build the application"
	@echo "  run          Build and run the registration service"
	@echo "  test         Run Go tests"
	@echo "  clean        Clean build artifacts"
	@echo ""
	@echo "🐳 Docker & Services:"
	@echo "  docker-build Build Docker image for the API"
	@echo "  db-up        Start database services (PostgreSQL + PGBouncer)"
	@echo "  db-down      Stop database services"
	@echo "  up           Start all application services"
	@echo "  down         Stop all services"
	@echo "  full-up      Start DB services, build image, and start app services"
	@echo "  full-down    Stop all services including database"
	@echo "  start        Complete automated startup"
	@echo "  status       Check system status"
	@echo "  logs         View application container logs"
	@echo "  db-logs      View database container logs"
	@echo ""
	@echo "🗄️ Database:"
	@echo "  migrate-up        Run database migrations"
	@echo "  migrate-status    Check migration status"
	@echo "  migrate-create    Create new migration (use NAME=migration_name)"
	@echo ""
	@echo "🔴 Redis:"
	@echo "  redis-up     Start Redis Sentinel cluster"
	@echo "  redis-down   Stop Redis Sentinel cluster"
	@echo "  redis-status Check Redis cluster health"
	@echo "  redis-test   Test Redis write/read operations"
	@echo "  redis-logs   View Redis container logs"
	@echo ""
	@echo "🔄 Load Testing Reset:"
	@echo "  reset-db        Full database reset with dependency checks"
	@echo "  reset-db-quick  Quick Python-only database reset"
	@echo ""
	@echo "📚 Documentation:"
	@echo "  See RESET_DATABASE_README.md for reset tool details"
	@echo "  See docs/ directory for architecture and testing guides"
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  make start       Complete system startup (recommended)"
	@echo "  make full-up     Manual step-by-step startup"
	@echo "  make docker-build && make db-up && make up"

fmt:
	gofmt -s -w .

tidy:
	$(GO) mod tidy

build: fmt tidy
	$(GO) build -o $(BINARY)

run: build
	$(BINARY) registration -v

# Docker commands
docker-build:
	./build.sh

network-create:
	./scripts/network-manager.sh create

db-up: network-create
	docker-compose -f docker-compose.db.yml up -d

db-down:
	docker-compose -f docker-compose.db.yml down

db-logs:
	docker-compose -f docker-compose.db.yml logs -f

up: docker-build network-create
	docker-compose up -d

down:
	docker-compose down

full-up: network-create db-up docker-build up

full-down:
	docker-compose down
	docker-compose -f docker-compose.db.yml down -v
	./scripts/network-manager.sh remove

logs:
	docker-compose logs -f

# System management
start:
	./start.sh

status:
	@echo "🔍 Database Services Status:"
	@docker-compose -f docker-compose.db.yml ps
	@echo ""
	@echo "🔍 Application Services Status:"
	@docker-compose ps
	@echo ""
	@echo "🌐 Quick Health Check:"
	@curl -s http://localhost/health 2>/dev/null | head -n 1 || echo "❌ Load balancer not responding"

test:
	$(GO) test $(PKG)

clean:
	rm -rf bin

# Redis commands
redis-up:
	docker-compose up -d redis-master redis-slave-1 redis-slave-2 redis-sentinel-1 redis-sentinel-2 redis-sentinel-3

redis-down:
	docker-compose stop redis-master redis-slave-1 redis-slave-2 redis-sentinel-1 redis-sentinel-2 redis-sentinel-3

redis-status:
	@echo "Checking Redis Sentinel cluster status..."
	@./scripts/redis-sentinel-manager.sh status

redis-test:
	@echo "Testing Redis Sentinel cluster..."
	@./scripts/redis-sentinel-manager.sh test-write

redis-logs:
	docker-compose logs -f redis-master redis-slave-1 redis-slave-2 redis-sentinel-1 redis-sentinel-2 redis-sentinel-3

# Migration commands
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


reset-db:
	@echo "🔄 Resetting registration database for load testing..."
	./reset_database.sh

reset-db-quick:
	@echo "🔄 Quick database reset (Python script only)..."
	python3 reset_registration_database.py
