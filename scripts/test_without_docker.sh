#!/bin/bash

# Тестирование сброса пароля без Docker
# Использует локальную базу данных PostgreSQL

set -e

echo "🚀 Тестирование сброса пароля (без Docker)"
echo "=========================================="

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
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

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Проверяем зависимости
check_dependencies() {
    print_info "Проверяем зависимости..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go не установлен"
        exit 1
    fi
    
    if ! command -v psql &> /dev/null; then
        print_warning "PostgreSQL клиент не установлен. Установите:"
        echo "  Ubuntu/Debian: sudo apt-get install postgresql-client"
        echo "  macOS: brew install postgresql"
        echo "  Или используйте Docker версию: ./scripts/quick_test.sh"
        exit 1
    fi
    
    print_success "Все зависимости установлены"
}

# Проверяем подключение к базе данных
check_database() {
    print_info "Проверяем подключение к базе данных..."
    
    # Проверяем переменные окружения
    if [ -z "$POSTGRES_HOST" ]; then
        export POSTGRES_HOST=localhost
    fi
    if [ -z "$POSTGRES_PORT" ]; then
        export POSTGRES_PORT=5432
    fi
    if [ -z "$POSTGRES_USER" ]; then
        export POSTGRES_USER=postgres
    fi
    if [ -z "$POSTGRES_PASSWORD" ]; then
        export POSTGRES_PASSWORD=postgres
    fi
    if [ -z "$POSTGRES_DB" ]; then
        export POSTGRES_DB=linkatype
    fi
    
    # Проверяем подключение
    if ! PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT 1;" &> /dev/null; then
        print_error "Не удается подключиться к базе данных"
        echo "Проверьте настройки:"
        echo "  POSTGRES_HOST=$POSTGRES_HOST"
        echo "  POSTGRES_PORT=$POSTGRES_PORT"
        echo "  POSTGRES_USER=$POSTGRES_USER"
        echo "  POSTGRES_DB=$POSTGRES_DB"
        echo ""
        echo "Убедитесь, что PostgreSQL запущен и доступен"
        exit 1
    fi
    
    print_success "Подключение к базе данных установлено"
}

# Создаем таблицы если их нет
create_tables() {
    print_info "Проверяем наличие таблиц..."
    
    # Проверяем существование таблицы users
    if ! PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT 1 FROM users LIMIT 1;" &> /dev/null; then
        print_info "Создаем таблицы..."
        
        # Создаем таблицу users
        PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "
        CREATE TABLE IF NOT EXISTS users (
            id VARCHAR(255) PRIMARY KEY,
            email VARCHAR(255) UNIQUE NOT NULL,
            password VARCHAR(255) NOT NULL,
            email_verified BOOLEAN DEFAULT FALSE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );"
        
        # Создаем таблицу otp_codes
        PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "
        CREATE TABLE IF NOT EXISTS otp_codes (
            id VARCHAR(255) PRIMARY KEY,
            email VARCHAR(255) NOT NULL,
            code VARCHAR(6) NOT NULL,
            type VARCHAR(20) NOT NULL,
            expires_at TIMESTAMP NOT NULL,
            used BOOLEAN DEFAULT FALSE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );"
        
        print_success "Таблицы созданы"
    else
        print_success "Таблицы уже существуют"
    fi
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
    print_success "Сервер запущен на http://localhost:8080"
}

# Очистка
cleanup() {
    print_info "Очистка..."
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
    rm -f server
}

trap cleanup EXIT

# Основная функция
main() {
    check_dependencies
    check_database
    create_tables
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
    echo "   PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c \"SELECT * FROM otp_codes;\""
    echo ""
    
    print_warning "Нажмите Ctrl+C для остановки"
    wait
}

main 