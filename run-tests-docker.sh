#!/bin/bash

# Скрипт для запуска тестов в Docker

set -e

echo "Запуск тестов в Docker..."

# Создаем директорию для результатов тестов
mkdir -p test-results

# Запускаем базу данных
echo "Запуск базы данных..."
docker compose up -d db

# Ждем, пока база данных будет готова
echo "Ожидание готовности базы данных..."
docker compose exec db pg_isready -U postgres

# Запускаем Go тесты
echo "Запуск Go тестов..."
docker compose --profile test run --rm go-tests

echo "Тесты завершены!"
