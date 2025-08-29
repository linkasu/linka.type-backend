#!/bin/bash

# Скрипт для запуска e2e тестов через docker-compose

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для вывода сообщений
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# Проверяем, что мы в корневой директории проекта
if [ ! -f "docker-compose.yml" ]; then
    error "docker-compose.yml не найден. Убедитесь, что вы находитесь в корневой директории проекта"
    exit 1
fi

# Проверяем наличие директории e2e-tests
if [ ! -d "e2e-tests" ]; then
    error "Директория e2e-tests не найдена"
    exit 1
fi

# Создаем директорию для результатов тестов
mkdir -p e2e-tests/test-results

# Функция для очистки
cleanup() {
    log "Очистка контейнеров..."
    docker compose --profile test down --remove-orphans
}

# Обработка сигналов для корректной очистки
trap cleanup EXIT INT TERM

# Парсим аргументы командной строки
BUILD=false
WATCH=false
COVERAGE=false
DEBUG=false
CLEAN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --build)
            BUILD=true
            shift
            ;;
        --watch)
            WATCH=true
            shift
            ;;
        --coverage)
            COVERAGE=true
            shift
            ;;
        --debug)
            DEBUG=true
            shift
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        --help)
            echo "Использование: $0 [опции]"
            echo ""
            echo "Опции:"
            echo "  --build     Пересобрать образы перед запуском"
            echo "  --watch     Запустить тесты в watch режиме"
            echo "  --coverage  Запустить тесты с покрытием"
            echo "  --debug     Запустить тесты в режиме отладки"
            echo "  --clean     Очистить контейнеры перед запуском"
            echo "  --help      Показать эту справку"
            exit 0
            ;;
        *)
            error "Неизвестная опция: $1"
            echo "Используйте --help для получения справки"
            exit 1
            ;;
    esac
done

# Очистка если запрошена
if [ "$CLEAN" = true ]; then
    log "Очистка контейнеров..."
    docker compose --profile test down --remove-orphans --volumes
fi

# Определяем команду для тестов
if [ "$WATCH" = true ]; then
    TEST_CMD="npm run test:docker:watch"
    log "Запуск тестов в watch режиме..."
elif [ "$COVERAGE" = true ]; then
    TEST_CMD="npm run test:coverage"
    log "Запуск тестов с покрытием..."
elif [ "$DEBUG" = true ]; then
    TEST_CMD="npm run test:debug"
    log "Запуск тестов в режиме отладки..."
else
    TEST_CMD="npm run test:docker"
    log "Запуск тестов..."
fi

# Пересборка образов если запрошена
if [ "$BUILD" = true ]; then
    log "Пересборка образов..."
    docker compose --profile test build --no-cache
fi

# Запускаем тесты
log "Запуск e2e тестов через docker-compose..."

# Заменяем команду в docker-compose
export TEST_COMMAND="$TEST_CMD"

# Запускаем с профилем test
docker compose --profile test up \
    --abort-on-container-exit \
    --exit-code-from e2e-tests \
    --build

# Проверяем результат
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    log "Тесты выполнены успешно!"
    
    # Показываем результаты покрытия если есть
    if [ "$COVERAGE" = true ] && [ -f "e2e-tests/test-results/coverage/lcov-report/index.html" ]; then
        info "Отчет о покрытии доступен в: e2e-tests/test-results/coverage/lcov-report/index.html"
    fi
    
    # Показываем JUnit отчет если есть
    if [ -f "e2e-tests/test-results/junit.xml" ]; then
        info "JUnit отчет доступен в: e2e-tests/test-results/junit.xml"
    fi
else
    error "Тесты завершились с ошибками (код: $EXIT_CODE)"
    
    # Показываем логи контейнера с тестами
    log "Логи контейнера e2e-tests:"
    docker compose --profile test logs e2e-tests
fi

exit $EXIT_CODE
