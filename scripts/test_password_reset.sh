#!/bin/bash

# Скрипт для тестирования сброса пароля
# Запускает сервер и тестирует функциональность сброса пароля

set -e

echo "🔐 Тестирование сброса пароля"
echo "=============================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода с цветом
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Проверяем, запущен ли Docker
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker не запущен. Запустите Docker и попробуйте снова."
        exit 1
    fi
    print_success "Docker доступен"
}

# Запускаем базу данных
start_database() {
    print_status "Запускаем базу данных..."
    docker compose up -d db
    
    # Ждем, пока база данных будет готова
    print_status "Ожидаем готовности базы данных..."
    timeout=60
    counter=0
    while ! docker compose exec -T db pg_isready -U postgres > /dev/null 2>&1; do
        sleep 1
        counter=$((counter + 1))
        if [ $counter -ge $timeout ]; then
            print_error "База данных не запустилась за $timeout секунд"
            exit 1
        fi
    done
    print_success "База данных готова"
}

# Собираем и запускаем сервер
start_server() {
    print_status "Собираем сервер..."
    go build -o server cmd/server/main.go
    
    print_status "Запускаем сервер..."
    ./server &
    SERVER_PID=$!
    
    # Ждем, пока сервер запустится
    print_status "Ожидаем запуска сервера..."
    timeout=30
    counter=0
    while ! curl -s http://localhost:8080/api/health > /dev/null 2>&1; do
        sleep 1
        counter=$((counter + 1))
        if [ $counter -ge $timeout ]; then
            print_error "Сервер не запустился за $timeout секунд"
            kill $SERVER_PID 2>/dev/null || true
            exit 1
        fi
    done
    print_success "Сервер запущен на http://localhost:8080"
}

# Останавливаем сервер
stop_server() {
    if [ ! -z "$SERVER_PID" ]; then
        print_status "Останавливаем сервер..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        print_success "Сервер остановлен"
    fi
}

# Останавливаем базу данных
stop_database() {
    print_status "Останавливаем базу данных..."
    docker compose down
    print_success "База данных остановлена"
}

# Очистка при выходе
cleanup() {
    print_status "Выполняем очистку..."
    stop_server
    stop_database
}

# Устанавливаем обработчик сигналов
trap cleanup EXIT INT TERM

# Основная функция
main() {
    print_status "Начинаем тестирование сброса пароля"
    
    # Проверяем Docker
    check_docker
    
    # Запускаем базу данных
    start_database
    
    # Запускаем сервер
    start_server
    
    print_status "Система готова для тестирования"
    echo ""
    echo "🎯 Теперь вы можете:"
    echo "1. Запустить интерактивный CLI:"
    echo "   go run scripts/password_reset_cli.go"
    echo ""
    echo "2. Запустить тестовый режим (автоматическое получение OTP):"
    echo "   go run scripts/password_reset_cli.go --test"
    echo ""
    echo "3. Протестировать через curl:"
    echo "   # Запрос сброса пароля"
    echo "   curl -X POST http://localhost:8080/api/auth/reset-password \\"
    echo "     -H 'Content-Type: application/json' \\"
    echo "     -d '{\"email\": \"test@example.com\"}'"
    echo ""
    echo "4. Запустить автоматические тесты:"
    echo "   go test ./tests/e2e/... -v"
    echo ""
    
    print_warning "Нажмите Ctrl+C для остановки"
    
    # Ждем сигнала для остановки
    wait
}

# Запускаем основную функцию
main 