# OTP API Documentation

## Обзор

API для работы с OTP (One-Time Password) кодами для подтверждения регистрации и сброса пароля.

## Endpoints

### 1. Регистрация с OTP

**POST** `/api/auth/register`

Регистрирует нового пользователя и отправляет OTP код на email.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (201):**
```json
{
  "message": "Registration successful. Please check your email for verification code.",
  "user_id": "generated-user-id"
}
```

### 2. Верификация Email

**POST** `/api/auth/verify-email`

Подтверждает email пользователя с помощью OTP кода.

**Request Body:**
```json
{
  "email": "user@example.com",
  "code": "123456"
}
```

**Response (200):**
```json
{
  "message": "Email verified successfully",
  "token": "jwt-token",
  "user": {
    "id": "user-id",
    "email": "user@example.com"
  }
}
```

### 3. Запрос сброса пароля

**POST** `/api/auth/reset-password`

Отправляет OTP код для сброса пароля.

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Response (200):**
```json
{
  "message": "If the email exists, a reset code has been sent"
}
```

### 4. Верификация OTP для сброса пароля

**POST** `/api/auth/reset-password/verify`

Проверяет OTP код для сброса пароля.

**Request Body:**
```json
{
  "email": "user@example.com",
  "code": "123456"
}
```

**Response (200):**
```json
{
  "message": "OTP verified successfully. You can now set a new password.",
  "otp_id": "otp-record-id"
}
```

### 5. Подтверждение сброса пароля

**POST** `/api/auth/reset-password/confirm`

Устанавливает новый пароль после верификации OTP.

**Request Body:**
```json
{
  "email": "user@example.com",
  "code": "123456",
  "password": "newpassword123"
}
```

**Response (200):**
```json
{
  "message": "Password reset successfully"
}
```

## Ошибки

### Общие ошибки

**400 Bad Request:**
```json
{
  "error": "Invalid OTP code format"
}
```

**400 Bad Request:**
```json
{
  "error": "Invalid or expired OTP code"
}
```

**400 Bad Request:**
```json
{
  "error": "OTP code has expired"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Failed to send OTP email"
}
```

### Специфичные ошибки

**401 Unauthorized (при логине):**
```json
{
  "error": "Email not verified. Please check your email for verification code."
}
```

## Безопасность

1. **OTP коды действительны 15 минут**
2. **OTP коды используются только один раз**
3. **При запросе нового OTP старые коды удаляются**
4. **Email должен быть верифицирован для доступа к защищенным ресурсам**

## Переменные окружения

```bash
# SMTP Configuration
MAIL_SERVER=smtp.yandex.ru
MAIL_ADRESS=feedback@linka.su
MAIL_PASSWORD=your-password
MAIL_PORT=587

# Server Configuration
PORT=8080
```

## Примеры использования

### Регистрация и верификация

```bash
# 1. Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'

# 2. Верификация email (после получения OTP в email)
curl -X POST http://localhost:8080/api/auth/verify-email \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "code": "123456"}'
```

### Сброс пароля

```bash
# 1. Запрос сброса пароля
curl -X POST http://localhost:8080/api/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'

# 2. Верификация OTP
curl -X POST http://localhost:8080/api/auth/reset-password/verify \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "code": "123456"}'

# 3. Установка нового пароля
curl -X POST http://localhost:8080/api/auth/reset-password/confirm \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "code": "123456", "password": "newpassword123"}'
``` 