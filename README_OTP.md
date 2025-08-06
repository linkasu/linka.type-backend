# Linka Type Backend - OTP Integration

## Обзор

Добавлена поддержка OTP (One-Time Password) для подтверждения регистрации и сброса пароля через SMTP.

## Новые возможности

### 1. Регистрация с подтверждением email
- При регистрации пользователь получает OTP код на email
- Email должен быть подтвержден для доступа к защищенным ресурсам
- OTP код действителен 15 минут

### 2. Сброс пароля через OTP
- Пользователь может запросить сброс пароля
- OTP код отправляется на email
- Новый пароль устанавливается после верификации OTP

## Установка и настройка

### 1. Переменные окружения

Убедитесь, что в файле `.env` настроены SMTP параметры:

```bash
# SMTP Configuration
MAIL_SERVER=smtp.yandex.ru
MAIL_ADRESS=feedback@linka.su
MAIL_PASSWORD=your-password
MAIL_PORT=587

# Server Configuration
PORT=8080
```

### 2. Запуск сервера

```bash
# Сборка
go build -o server cmd/server/main.go

# Запуск
./server
```

## API Endpoints

### Аутентификация

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/auth/register` | Регистрация с отправкой OTP |
| POST | `/api/auth/login` | Логин (требует верификации email) |
| POST | `/api/auth/verify-email` | Верификация email с OTP |
| POST | `/api/auth/reset-password` | Запрос сброса пароля |
| POST | `/api/auth/reset-password/verify` | Верификация OTP для сброса |
| POST | `/api/auth/reset-password/confirm` | Установка нового пароля |

### Защищенные ресурсы

Все защищенные ресурсы требуют:
1. JWT токен в заголовке `Authorization: Bearer <token>`
2. Верифицированный email

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/statements` | Получение statements |
| POST | `/api/statements` | Создание statement |
| PUT | `/api/statements/:id` | Обновление statement |
| DELETE | `/api/statements/:id` | Удаление statement |
| GET | `/api/categories` | Получение категорий |
| POST | `/api/categories` | Создание категории |
| PUT | `/api/categories/:id` | Обновление категории |
| DELETE | `/api/categories/:id` | Удаление категории |

## Примеры использования

### Регистрация и верификация

```bash
# 1. Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'

# Ответ:
{
  "message": "Registration successful. Please check your email for verification code.",
  "user_id": "generated-user-id"
}

# 2. Верификация email (после получения OTP в email)
curl -X POST http://localhost:8080/api/auth/verify-email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "code": "123456"
  }'

# Ответ:
{
  "message": "Email verified successfully",
  "token": "jwt-token",
  "user": {
    "id": "user-id",
    "email": "user@example.com"
  }
}
```

### Сброс пароля

```bash
# 1. Запрос сброса пароля
curl -X POST http://localhost:8080/api/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com"
  }'

# 2. Верификация OTP
curl -X POST http://localhost:8080/api/auth/reset-password/verify \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "code": "123456"
  }'

# 3. Установка нового пароля
curl -X POST http://localhost:8080/api/auth/reset-password/confirm \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "code": "123456",
    "password": "newpassword123"
  }'
```

### Использование защищенных ресурсов

```bash
# Получение statements (требует JWT токен)
curl -X GET http://localhost:8080/api/statements \
  -H "Authorization: Bearer your-jwt-token"

# Создание statement
curl -X POST http://localhost:8080/api/statements \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New statement",
    "category_id": "category-id"
  }'
```

## Безопасность

### OTP коды
- **Длина**: 6 цифр
- **Время жизни**: 15 минут
- **Использование**: Одноразовые
- **Автоматическая очистка**: Старые коды удаляются при запросе новых

### Email верификация
- **Обязательна** для доступа к защищенным ресурсам
- **Проверяется** при каждом логине
- **Автоматическая отправка** приветственного письма после верификации

### Пароли
- **Хеширование**: MD5 (как в существующей системе)
- **Минимальная длина**: 6 символов
- **Сброс**: Только через OTP верификацию

## База данных

### Новые таблицы

#### Таблица `otp_codes`
```sql
CREATE TABLE otp_codes (
    id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    code VARCHAR(6) NOT NULL,
    type VARCHAR(20) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Обновленная таблица `users`
```sql
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;
```

## Тестирование

### Unit тесты
```bash
go test ./tests/unit/... -v
```

### Сборка
```bash
go build -o server cmd/server/main.go
```

## Структура проекта

```
├── cmd/server/main.go          # Основной сервер
├── handlers/auth.go            # Обработчики аутентификации
├── auth/middleware.go          # Middleware для проверки JWT и верификации
├── mail/mail.go               # SMTP функциональность
├── otp/otp.go                 # Генерация и валидация OTP
├── db/
│   ├── types.go               # Обновленные типы данных
│   ├── user_crud.go           # CRUD операции для пользователей
│   ├── otp_crud.go            # CRUD операции для OTP
│   └── connection.go          # Обновленные миграции
├── tests/unit/otp_test.go     # Тесты OTP функциональности
└── docs/api_otp.md           # Документация API
```

## Устранение неполадок

### Проблемы с SMTP
1. Проверьте настройки в `.env`
2. Убедитесь, что SMTP сервер доступен
3. Проверьте логи сервера

### Проблемы с OTP
1. Проверьте, что email корректный
2. Убедитесь, что OTP код введен правильно
3. Проверьте время жизни OTP (15 минут)

### Проблемы с верификацией
1. Убедитесь, что email верифицирован
2. Проверьте JWT токен
3. Проверьте права доступа

## Дополнительная информация

- [Документация API](docs/api_otp.md)
- [Примеры использования](docs/api_otp.md#примеры-использования)
- [Безопасность](docs/api_otp.md#безопасность) 