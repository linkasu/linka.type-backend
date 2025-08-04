.PHONY: test test-e2e test-unit build clean run-server run-playground docker-build docker-run

# Переменные
BINARY_NAME=linka-type-backend
SERVER_BINARY=server
PLAYGROUND_BINARY=playground

# Тесты
test: test-unit test-integration test-e2e

test-unit:
	@echo "Running unit tests..."
	go test ./auth/... ./utils/... ./tests/unit/... -v

test-integration:
	@echo "Running integration tests..."
	go test ./tests/integration/... -v

test-e2e:
	@echo "Starting test database..."
	docker compose -f docker-compose.test.yml up -d test-db
	@echo "Waiting for database to be ready..."
	@sleep 15
	@echo "Running e2e tests..."
	go test ./tests/e2e/... -v
	@echo "Stopping test database..."
	docker compose -f docker-compose.test.yml down

# Сборка
build: build-server build-playground

build-server:
	@echo "Building server..."
	go build -o bin/$(SERVER_BINARY) ./cmd/server

build-playground:
	@echo "Building playground..."
	go build -o bin/$(PLAYGROUND_BINARY) ./cmd/playground

# Запуск
run-server: build-server
	@echo "Starting server on http://localhost:8081"
	./bin/$(SERVER_BINARY)

run-playground: build-playground
	@echo "Starting playground..."
	./bin/$(PLAYGROUND_BINARY)

# Docker
docker-build:
	@echo "Building Docker images..."
	docker compose build

docker-run:
	@echo "Starting services with Docker Compose..."
	docker compose up

docker-run-detached:
	@echo "Starting services with Docker Compose (detached)..."
	docker compose up -d

docker-stop:
	@echo "Stopping services..."
	docker compose down

# Очистка
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Помощь
help:
	@echo "Available commands:"
	@echo "  test              - Run all tests (unit + e2e)"
	@echo "  test-unit         - Run unit tests only"
	@echo "  test-e2e          - Run e2e tests only"
	@echo "  build             - Build all binaries"
	@echo "  build-server      - Build server binary"
	@echo "  build-playground  - Build playground binary"
	@echo "  run-server        - Build and run server"
	@echo "  run-playground    - Build and run playground"
	@echo "  docker-build      - Build Docker images"
	@echo "  docker-run        - Run with Docker Compose"
	@echo "  docker-run-detached - Run with Docker Compose (detached)"
	@echo "  docker-stop       - Stop Docker services"
	@echo "  clean             - Clean build artifacts"
	@echo "  help              - Show this help" 