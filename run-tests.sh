#!/bin/bash

# Скрипт для запуска тестов с правильными переменными окружения

set -e

echo "Запуск тестов..."

# Устанавливаем переменные окружения для тестов
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=postgres
export POSTGRES_DB=linkatype
export JWT_SECRET=test-secret-key
export JWT_ISSUER=test-issuer
export JWT_AUDIENCE=test-audience

# Запускаем unit тесты (не требуют БД)
echo "Запуск unit тестов..."
go test ./auth -v
go test ./utils -v
go test ./otp -v
go test ./tests/unit -v

# Запускаем integration тесты (не требуют БД)
echo "Запуск integration тестов..."
go test ./tests/integration -v

# Запускаем тесты, которые требуют БД (только если БД доступна)
echo "Проверка доступности базы данных..."
if docker compose ps db | grep -q "Up"; then
    echo "База данных доступна, запуск тестов с БД..."
    go test ./bl -v
    go test ./tests/e2e -v
else
    echo "База данных недоступна, пропуск тестов с БД"
    echo "Для запуска тестов с БД выполните: docker compose up -d db"
fi

echo "Тесты завершены!"
