#!/bin/bash

# Скрипт для быстрого обновления только на сервере
# Использование: ./update-server-only.sh

set -e

SERVER="linka.su"
REMOTE_USER="aacidov"
REMOTE_DIR="/home/linka.type-backend"

echo "🚀 Быстрое обновление на сервере $SERVER..."

# Копируем обновленные файлы на сервер
echo "📋 Копирую конфигурацию на сервер..."
scp docker-compose.yml $REMOTE_USER@$SERVER:$REMOTE_DIR/
scp env.server.example $REMOTE_USER@$SERVER:$REMOTE_DIR/

# Выполняем обновление на сервере
echo "🔧 Обновляю на сервере..."
ssh $REMOTE_USER@$SERVER << 'EOF'
cd ~/linka.type-backend

echo "🔄 Останавливаю существующие контейнеры..."
docker compose down 2>/dev/null || true

echo "📥 Загружаю последний образ из Docker Hub..."
docker pull bakaidov/linka-type-backend:latest

echo "🚀 Запускаю сервисы..."
docker compose up -d

echo "⏳ Жду запуска сервисов..."
sleep 10

echo "📊 Статус контейнеров:"
docker ps

echo "🔍 Проверяю здоровье сервисов..."
docker compose ps

echo "📋 Последние логи сервера:"
docker compose logs --tail=10 server
EOF

echo "✅ Обновление завершено!"
echo "🌐 Сервер доступен по адресу: https://type-backend.$SERVER"
