#!/bin/bash

# Быстрый скрипт обновления через Docker Hub
# Использование: ./quick-update.sh

set -e

SERVER="linka.su"
REMOTE_USER="aacidov"
REMOTE_DIR="/home/linka.type-backend"

echo "🚀 Быстрое обновление на сервере $SERVER..."

# Копируем обновленные файлы на сервер
echo "📋 Копирую конфигурацию на сервер..."
scp docker-compose.yml $REMOTE_USER@$SERVER:$REMOTE_DIR/

# Выполняем обновление на сервере
echo "🔧 Обновляю на сервере..."
ssh $REMOTE_USER@$SERVER << 'EOF'
cd ~/linka.type-backend

echo "🔄 Останавливаю существующие контейнеры..."
docker compose down

echo "📥 Загружаю последний образ из Docker Hub..."
docker pull bakaidov/linka-type-backend:latest

echo "🚀 Запускаю сервисы..."
docker compose up -d

echo "⏳ Жду запуска сервисов..."
sleep 15

echo "📊 Статус контейнеров:"
docker ps

echo "🔍 Проверяю Unix сокет:"
ls -la /var/run/linka-type-backend.sock

echo "📋 Проверяю доступность:"
curl -f https://type-backend.linka.su/api/health
EOF

echo "✅ Обновление завершено!"
echo "🌐 Сервер доступен по адресу: https://type-backend.$SERVER"
