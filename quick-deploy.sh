#!/bin/bash

# Быстрый скрипт деплоя (только обновление на сервере)
# Использование: ./quick-deploy.sh

set -e

# Конфигурация
SERVER="linka.su"
REMOTE_USER="aacidov"
REMOTE_DIR="/home/linka.type-backend"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}⚡ Быстрый деплой на сервер $SERVER...${NC}"

# Проверка подключения к серверу
echo -e "${YELLOW}📡 Проверяю подключение к серверу...${NC}"
if ! ssh -o ConnectTimeout=10 $REMOTE_USER@$SERVER "echo 'SSH подключение работает'" >/dev/null 2>&1; then
    echo -e "${RED}❌ Не удается подключиться к серверу $SERVER${NC}"
    exit 1
fi

# Копирование файлов на сервер
echo -e "${YELLOW}📋 Копирую конфигурацию на сервер...${NC}"
scp docker-compose.yml $REMOTE_USER@$SERVER:$REMOTE_DIR/

# Быстрое обновление на сервере
echo -e "${YELLOW}🔧 Обновляю на сервере...${NC}"
ssh $REMOTE_USER@$SERVER << 'EOF'
cd ~/linka.type-backend

echo "📥 Загружаю последний образ из Docker Hub..."
docker pull bakaidov/linka-type-backend:latest

echo "🔄 Перезапускаю сервисы..."
docker compose restart server

echo "⏳ Жду запуска сервисов..."
sleep 10

echo "📊 Статус контейнеров:"
docker ps

echo "🔍 Проверяю Unix сокет:"
ls -la /var/run/linka-type-backend.sock

echo "📋 Проверяю доступность:"
if curl -f https://type-backend.linka.su/api/health; then
    echo "✅ Сервер доступен"
else
    echo "❌ Сервер недоступен"
    exit 1
fi

echo "📋 Логи сервера:"
docker compose logs --tail=5 server
EOF

# Проверка результата
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Быстрое обновление завершено успешно!${NC}"
    echo -e "${GREEN}🌐 Сервер доступен по адресу: https://type-backend.$SERVER${NC}"
else
    echo -e "${RED}❌ Обновление завершилось с ошибкой${NC}"
    exit 1
fi

echo -e "${BLUE}🎉 Быстрое обновление завершено!${NC}"
