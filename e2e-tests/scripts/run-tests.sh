#!/bin/bash

# Скрипт для запуска e2e тестов с проверкой готовности сервера

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Проверяем, что мы в правильной директории
if [ ! -f "package.json" ]; then
    error "package.json не найден. Убедитесь, что вы находитесь в директории e2e-tests"
    exit 1
fi

# Проверяем наличие .env файла
if [ ! -f ".env" ]; then
    warn ".env файл не найден. Создаем из примера..."
    if [ -f "env.example" ]; then
        cp env.example .env
        log ".env файл создан из env.example"
    else
        error "env.example не найден"
        exit 1
    fi
fi

# Загружаем переменные окружения
source .env

# Устанавливаем базовый URL по умолчанию
BASE_URL=${BASE_URL:-"http://localhost:8081"}
API_HEALTH_URL="${BASE_URL}/api/health"

log "Проверяем готовность сервера по адресу: $API_HEALTH_URL"

# Функция для проверки готовности сервера
check_server_ready() {
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$API_HEALTH_URL" > /dev/null 2>&1; then
            log "Сервер готов! (попытка $attempt/$max_attempts)"
            return 0
        fi
        
        warn "Сервер не готов, попытка $attempt/$max_attempts"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    error "Сервер не готов после $max_attempts попыток"
    return 1
}

# Проверяем готовность сервера
if ! check_server_ready; then
    error "Не удалось подключиться к серверу. Убедитесь, что сервер запущен:"
    echo "  docker-compose up -d"
    echo "  docker-compose logs server"
    exit 1
fi

# Проверяем, что зависимости установлены
if [ ! -d "node_modules" ]; then
    log "Устанавливаем зависимости..."
    npm install
fi

# Запускаем тесты
log "Запускаем e2e тесты..."

# Проверяем аргументы командной строки
if [ "$1" = "--watch" ]; then
    log "Запуск в watch режиме..."
    npm run test:watch
elif [ "$1" = "--coverage" ]; then
    log "Запуск с покрытием..."
    npm run test:coverage
elif [ "$1" = "--debug" ]; then
    log "Запуск в режиме отладки..."
    npm run test:debug
elif [ -n "$1" ]; then
    log "Запуск конкретного теста: $1"
    npm test -- "$1"
else
    log "Запуск всех тестов..."
    npm test
fi

# Проверяем результат выполнения
if [ $? -eq 0 ]; then
    log "Тесты выполнены успешно!"
else
    error "Тесты завершились с ошибками"
    exit 1
fi
