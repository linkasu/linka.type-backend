#!/bin/bash

# Скрипт для миграции категорий из Firebase в PostgreSQL
# Использование: ./scripts/migrate.sh <email> <password>

set -e

# Проверяем количество аргументов
if [ $# -ne 2 ]; then
    echo "Использование: $0 <email> <password>"
    echo "Пример: $0 user@example.com password123"
    exit 1
fi

EMAIL=$1
PASSWORD=$2

echo "🚀 Запуск миграции категорий для пользователя: $EMAIL"
echo "=================================================="

# Проверяем, что Go установлен
if ! command -v go &> /dev/null; then
    echo "❌ Ошибка: Go не установлен"
    exit 1
fi

# Проверяем, что мы в корневой директории проекта
if [ ! -f "go.mod" ]; then
    echo "❌ Ошибка: Запустите скрипт из корневой директории проекта"
    exit 1
fi

# Скачиваем зависимости
echo "📦 Скачивание зависимостей..."
go mod download

# Запускаем миграцию
echo "🔄 Запуск миграции..."
echo "📝 Примечание: Пароль будет автоматически хеширован перед сохранением в базу данных"
go run cmd/import_example/main.go "$EMAIL" "$PASSWORD"

echo "✅ Миграция завершена!"
echo "📊 Результаты сохранены в логах выше" 