#!/bin/bash

# Запуск сервера через Docker Compose
# Сервер будет подключен к базе данных через внутреннюю сеть Docker

set -e

echo "🚀 Запуск сервера через Docker Compose"
echo "======================================"

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Останавливаем все контейнеры
cleanup() {
    print_info "Очистка..."
    docker compose down
}

trap cleanup EXIT

# Запускаем базу данных
print_info "Запускаем базу данных..."
docker compose up -d db

# Ждем готовности базы данных
print_info "Ожидаем готовности базы данных..."
timeout=60
counter=0
while ! docker compose exec -T db pg_isready -U postgres &> /dev/null; do
    sleep 1
    counter=$((counter + 1))
    if [ $counter -ge $timeout ]; then
        print_error "База данных не запустилась за $timeout секунд"
        exit 1
    fi
done
print_success "База данных готова"

# Собираем образ сервера
print_info "Собираем образ сервера..."
docker compose build server

# Запускаем сервер
print_info "Запускаем сервер..."
docker compose up -d server

# Ждем запуска сервера
print_info "Ожидаем запуска сервера..."
timeout=30
counter=0
while ! curl -s http://localhost:8081/api/health &> /dev/null; do
    sleep 1
    counter=$((counter + 1))
    if [ $counter -ge $timeout ]; then
        print_error "Сервер не запустился за $timeout секунд"
        exit 1
    fi
done
print_success "Сервер запущен на http://localhost:8081"

echo ""
print_success "Система готова для тестирования!"
echo ""
echo "🎯 Доступные команды:"
echo ""
echo "1. Интерактивный CLI (тестовый режим):"
echo "   go run scripts/password_reset_cli.go http://localhost:8081 --test"
echo ""
echo "2. Автоматические тесты:"
echo "   go test ./tests/e2e/... -v -run TestPasswordReset"
echo ""
echo "3. Ручное тестирование:"
echo "   curl -X POST http://localhost:8081/api/auth/reset-password \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"email\": \"test@example.com\"}'"
echo ""
echo "4. Проверка базы данных:"
echo "   docker compose exec db psql -U postgres -d linkatype -c \"SELECT * FROM otp_codes;\""
echo ""
echo "5. Логи сервера:"
echo "   docker compose logs -f server"
echo ""
echo "6. Логи базы данных:"
echo "   docker compose logs -f db"
echo ""

print_info "Нажмите Ctrl+C для остановки"
wait 