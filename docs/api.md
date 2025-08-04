# API Documentation

## Base URL
```
http://localhost:8081/api
```

## Аутентификация

API использует JWT токены для аутентификации. Токен должен передаваться в заголовке `Authorization` в формате:
```
Authorization: Bearer <token>
```

## Endpoints

### Публичные endpoints (без авторизации)

#### Регистрация
```
POST /register
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "user_1234567890",
    "email": "user@example.com"
  }
}
```

#### Логин
```
POST /login
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "user_1234567890",
    "email": "user@example.com"
  }
}
```

#### Health Check
```
GET /health
```

**Response:**
```json
{
  "status": "ok"
}
```

### Защищенные endpoints (требуют JWT токен)

#### Получить профиль пользователя
```
GET /profile
```

**Response:**
```json
{
  "user_id": "user_1234567890",
  "email": "user@example.com"
}
```

### Categories

#### Получить все категории пользователя
```
GET /categories
```

**Response:**
```json
{
  "categories": [
    {
      "id": "category_123",
      "title": "Работа",
      "userId": "user_1234567890"
    }
  ]
}
```

#### Получить конкретную категорию
```
GET /categories/:id
```

#### Создать категорию
```
POST /categories
```

**Request Body:**
```json
{
  "title": "Новая категория"
}
```

#### Обновить категорию
```
PUT /categories/:id
```

**Request Body:**
```json
{
  "title": "Обновленная категория"
}
```

#### Удалить категорию
```
DELETE /categories/:id
```

### Statements

#### Получить все statements пользователя
```
GET /statements
```

**Response:**
```json
{
  "statements": [
    {
      "id": "statement_123",
      "text": "Текст statement",
      "userId": "user_1234567890",
      "categoryId": "category_123"
    }
  ]
}
```

#### Получить конкретный statement
```
GET /statements/:id
```

#### Создать statement
```
POST /statements
```

**Request Body:**
```json
{
  "title": "Текст statement",
  "categoryId": "category_123"
}
```

#### Обновить statement
```
PUT /statements/:id
```

**Request Body:**
```json
{
  "title": "Обновленный текст statement",
  "categoryId": "category_123"
}
```

#### Удалить statement
```
DELETE /statements/:id
```

## Коды ошибок

- `400 Bad Request` - Неверный формат запроса
- `401 Unauthorized` - Неверные учетные данные или отсутствует токен
- `403 Forbidden` - Доступ запрещен
- `404 Not Found` - Ресурс не найден
- `409 Conflict` - Конфликт (например, пользователь уже существует)
- `500 Internal Server Error` - Внутренняя ошибка сервера

## Примеры использования

### Регистрация и получение токена
```bash
curl -X POST http://localhost:8081/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Создание категории
```bash
curl -X POST http://localhost:8081/api/categories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Мои задачи"
  }'
```

### Создание statement
```bash
curl -X POST http://localhost:8081/api/statements \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Выполнить задачу",
    "categoryId": "category_123"
  }'
``` 