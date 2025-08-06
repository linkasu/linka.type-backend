#!/bin/bash

# Быстрый тест сброса пароля
# Запускает минимальное окружение для тестирования

set -e

echo "🚀 Быстрый тест сброса пароля"
echo "=============================="

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

# Проверяем зависимости
check_dependencies() {
    print_info "Проверяем зависимости..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go не установлен"
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker не установлен"
        exit 1
    fi
    
    print_success "Все зависимости установлены"
}

# Запускаем базу данных
start_db() {
    print_info "Запускаем базу данных..."
    docker compose up -d db
    
    # Ждем готовности базы
    print_info "Ожидаем готовности базы данных..."
    timeout=30
    while ! docker compose exec -T db pg_isready -U postgres &> /dev/null; do
        sleep 1
        timeout=$((timeout - 1))
        if [ $timeout -le 0 ]; then
            print_error "База данных не запустилась"
            exit 1
        fi
    done
    print_success "База данных готова"
}

# Запускаем сервер
start_server() {
    print_info "Собираем и запускаем сервер..."
    go build -o server cmd/server/main.go
    ./server &
    SERVER_PID=$!
    
    # Ждем запуска сервера
    print_info "Ожидаем запуска сервера..."
    timeout=15
    while ! curl -s http://localhost:8080/api/health &> /dev/null; do
        sleep 1
        timeout=$((timeout - 1))
        if [ $timeout -le 0 ]; then
            print_error "Сервер не запустился"
            kill $SERVER_PID 2>/dev/null || true
            exit 1
        fi
    done
    print_success "Сервер запущен"
}

# Очистка
cleanup() {
    print_info "Очистка..."
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
    docker compose down 2>/dev/null || true
    rm -f server
}

trap cleanup EXIT

# Основная функция
main() {
    check_dependencies
    start_db
    start_server
    
    echo ""
    print_success "Система готова для тестирования!"
    echo ""
    echo "🎯 Доступные команды:"
    echo ""
    echo "1. Интерактивный CLI (тестовый режим):"
    echo "   go run scripts/password_reset_cli.go --test"
    echo ""
    echo "2. Автоматические тесты:"
    echo "   go test ./tests/e2e/... -v -run TestPasswordReset"
    echo ""
    echo "3. Ручное тестирование:"
    echo "   curl -X POST http://localhost:8080/api/auth/reset-password \\"
    echo "     -H 'Content-Type: application/json' \\"
    echo "     -d '{\"email\": \"test@example.com\"}'"
    echo ""
    echo "4. Проверка базы данных:"
    echo "   docker-compose exec db psql -U postgres -d linkatype -c \"SELECT * FROM otp_codes;\""
    echo ""
    
    print_info "Нажмите Ctrl+C для остановки"
    wait
}

main 