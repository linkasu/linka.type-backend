#!/bin/bash

# Универсальный скрипт деплоя на сервер
# Использование: ./deploy.sh [--force-rebuild] [--skip-tests]

set -e

# Конфигурация
SERVER="linka.su"
REMOTE_USER="aacidov"
REMOTE_DIR="/home/aacidov/linka.type-backend"
DOCKER_IMAGE="bakaidov/linka-type-backend"
TAG="latest"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Флаги
FORCE_REBUILD=false
SKIP_TESTS=false

# Парсинг аргументов
while [[ $# -gt 0 ]]; do
    case $1 in
        --force-rebuild)
            FORCE_REBUILD=true
            shift
            ;;
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        *)
            echo "Неизвестный аргумент: $1"
            echo "Использование: $0 [--force-rebuild] [--skip-tests]"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}🚀 Начинаю деплой на сервер $SERVER...${NC}"

# Проверка подключения к серверу
echo -e "${YELLOW}📡 Проверяю подключение к серверу...${NC}"
if ! ssh -o ConnectTimeout=10 $REMOTE_USER@$SERVER "echo 'SSH подключение работает'" >/dev/null 2>&1; then
    echo -e "${RED}❌ Не удается подключиться к серверу $SERVER${NC}"
    exit 1
fi

# Запуск тестов (если не пропущены)
if [ "$SKIP_TESTS" = false ]; then
    echo -e "${YELLOW}🧪 Запускаю тесты в Docker...${NC}"
    
    # Запускаем базу данных для тестов
    echo -e "${YELLOW}🗄️ Запускаю базу данных для тестов...${NC}"
    docker compose up -d db
    
    # Ждем готовности базы данных
    echo -e "${YELLOW}⏳ Жду готовности базы данных...${NC}"
    sleep 10
    
    # Запускаем тесты в Docker
    if ! docker compose --profile test run --rm go-tests; then
        echo -e "${RED}❌ Тесты не прошли!${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Тесты прошли успешно${NC}"
    
    # Останавливаем базу данных
    docker compose down
else
    echo -e "${YELLOW}⏭️ Пропускаю тесты${NC}"
fi

# Сборка образа
echo -e "${YELLOW}🔨 Собираю Docker образ...${NC}"
if [ "$FORCE_REBUILD" = true ]; then
    echo -e "${YELLOW}🔄 Принудительная пересборка...${NC}"
    docker compose build --no-cache
else
    docker compose build
fi

# Тегирование образа
echo -e "${YELLOW}🏷️ Тегирую образ для Docker Hub...${NC}"
docker tag linka.type-backend:latest $DOCKER_IMAGE:$TAG

# Проверка авторизации в Docker Hub
echo -e "${YELLOW}🔐 Проверяю авторизацию в Docker Hub...${NC}"
if ! docker info | grep -q "Username"; then
    echo -e "${RED}❌ Не авторизован в Docker Hub. Выполните: docker login${NC}"
    exit 1
fi

# Загрузка в Docker Hub
echo -e "${YELLOW}📤 Загружаю образ в Docker Hub...${NC}"
docker push $DOCKER_IMAGE:$TAG
echo -e "${GREEN}✅ Образ загружен в Docker Hub${NC}"

# Копирование файлов на сервер
echo -e "${YELLOW}📋 Копирую конфигурацию на сервер...${NC}"
scp docker-compose.yml $REMOTE_USER@$SERVER:$REMOTE_DIR/

# Деплой на сервер
echo -e "${YELLOW}🔧 Развертываю на сервере...${NC}"
ssh $REMOTE_USER@$SERVER << 'EOF'
cd ~/linka.type-backend

echo "🔄 Останавливаю существующие контейнеры..."
docker compose down

echo "🗑️ Удаляю старые образы..."
docker images | grep linka | awk '{print $3}' | xargs -r docker rmi -f 2>/dev/null || true

echo "📥 Загружаю новый образ из Docker Hub..."
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
if curl -f https://type-backend.linka.su/api/health; then
    echo "✅ Сервер доступен"
else
    echo "❌ Сервер недоступен"
    exit 1
fi

echo "📋 Логи сервера:"
docker compose logs --tail=10 server
EOF

# Проверка результата
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Деплой завершен успешно!${NC}"
    echo -e "${GREEN}🌐 Сервер доступен по адресу: https://type-backend.$SERVER${NC}"
    
    # Финальная проверка
    echo -e "${YELLOW}🔍 Финальная проверка...${NC}"
    if curl -f https://type-backend.linka.su/api/health >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Сервер работает корректно${NC}"
    else
        echo -e "${RED}⚠️ Сервер может быть недоступен${NC}"
    fi
else
    echo -e "${RED}❌ Деплой завершился с ошибкой${NC}"
    exit 1
fi

echo -e "${BLUE}🎉 Деплой завершен!${NC}"

