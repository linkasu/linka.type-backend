#!/bin/bash

# Скрипт для развертывания через Docker Hub
# Использование: ./deploy-via-dockerhub.sh

set -e

DOCKER_IMAGE="bakaidov/linka-type-backend"
TAG="latest"
SERVER="linka.su"
REMOTE_USER="aacidov"
REMOTE_DIR="/home/linka.type-backend"

echo "🚀 Начинаю развертывание через Docker Hub..."

# Проверяем, что Docker Hub доступен
echo "📡 Проверяю доступность Docker Hub..."
if ! docker pull hello-world:latest >/dev/null 2>&1; then
    echo "❌ Не удается подключиться к Docker Hub"
    exit 1
fi

# Собираем образ
echo "🔨 Собираю Docker образ..."
docker compose build --no-cache

# Тегируем образ для Docker Hub
echo "🏷️ Тегирую образ для Docker Hub..."
docker tag linka.type-backend:latest $DOCKER_IMAGE:$TAG

# Логинимся в Docker Hub (если нужно)
echo "🔐 Проверяю авторизацию в Docker Hub..."
if ! docker info | grep -q "Username"; then
    echo "⚠️ Не авторизован в Docker Hub. Выполните: docker login"
    echo "Продолжаю без загрузки в Docker Hub..."
    UPLOAD_TO_HUB=false
else
    UPLOAD_TO_HUB=true
fi

# Загружаем в Docker Hub (если авторизованы)
if [ "$UPLOAD_TO_HUB" = true ]; then
    echo "📤 Загружаю образ в Docker Hub..."
    docker push $DOCKER_IMAGE:$TAG
    echo "✅ Образ загружен в Docker Hub"
else
    echo "⏭️ Пропускаю загрузку в Docker Hub"
fi

# Копируем обновленные файлы на сервер
echo "📋 Копирую конфигурацию на сервер..."
scp docker-compose.yml $REMOTE_USER@$SERVER:$REMOTE_DIR/
scp env.server.example $REMOTE_USER@$SERVER:$REMOTE_DIR/

# Выполняем развертывание на сервере
echo "🔧 Развертываю на сервере..."
ssh $REMOTE_USER@$SERVER << 'EOF'
cd ~/linka.type-backend

echo "🔄 Останавливаю существующие контейнеры..."
docker compose down 2>/dev/null || true

echo "🗑️ Удаляю старые образы..."
docker images | grep linka | awk '{print $3}' | xargs -r docker rmi -f 2>/dev/null || true

echo "📥 Загружаю образ из Docker Hub..."
docker pull aacidov/linka-type-backend:latest

echo "🚀 Запускаю сервисы..."
docker compose up -d

echo "⏳ Жду запуска сервисов..."
sleep 15

echo "📊 Статус контейнеров:"
docker ps

echo "🔍 Проверяю здоровье сервисов..."
docker compose ps

echo "📋 Логи сервера:"
docker compose logs --tail=20 server
EOF

echo "✅ Развертывание завершено!"
echo "🌐 Сервер доступен по адресу: https://type-backend.$SERVER"

if [ "$UPLOAD_TO_HUB" = false ]; then
    echo ""
    echo "💡 Для загрузки в Docker Hub выполните:"
    echo "   docker login"
    echo "   docker push $DOCKER_IMAGE:$TAG"
fi
