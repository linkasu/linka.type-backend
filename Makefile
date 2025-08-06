.PHONY: test test-unit build clean run-server run-playground docker-build docker-run docs

# Переменные
BINARY_NAME=linka-type-backend
SERVER_BINARY=server
PLAYGROUND_BINARY=playground

# Тесты
test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	go test ./auth/... ./utils/... ./tests/unit/... -v

test-integration:
	@echo "Running integration tests..."
	go test ./tests/integration/... -v

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

# Документация
docs:
	@echo "Generating documentation..."
	go run ./cmd/docs

docs-serve:
	@echo "Serving documentation on http://localhost:8080"
	python3 -m http.server 8080 --directory docs

# Помощь
help:
	@echo "Available commands:"
	@echo "  test              - Run all tests (unit + integration)"
	@echo "  test-unit         - Run unit tests only"
	@echo "  test-integration  - Run integration tests only"
	@echo "  build             - Build all binaries"
	@echo "  build-server      - Build server binary"
	@echo "  build-playground  - Build playground binary"
	@echo "  run-server        - Build and run server"
	@echo "  run-playground    - Build and run playground"
	@echo "  docker-build      - Build Docker images"
	@echo "  docker-run        - Run with Docker Compose"
	@echo "  docker-run-detached - Run with Docker Compose (detached)"
	@echo "  docker-stop       - Stop Docker services"
	@echo "  docs              - Generate documentation"
	@echo "  docs-serve        - Serve documentation locally"
	@echo "  clean             - Clean build artifacts"
	@echo "  help              - Show this help" 