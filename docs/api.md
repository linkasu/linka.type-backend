# API Documentation

## CORS Configuration

The API supports configurable CORS settings through environment variables:

- `CORS_ORIGINS`: Comma-separated list of allowed origins (default: `http://localhost:3000,http://localhost:8080`)
- `CORS_METHODS`: Comma-separated list of allowed HTTP methods (default: `GET, POST, PUT, DELETE, OPTIONS`)
- `CORS_HEADERS`: Comma-separated list of allowed headers (default: `Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization`)

For production, set `CORS_ORIGINS` to your specific domain(s), e.g.:
```
CORS_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
```

## Authentication

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

### WebSocket

#### Подключение к WebSocket
```
GET /ws
```

**Headers:**
```
Authorization: Bearer <token>
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: <base64-encoded-key>
Sec-WebSocket-Version: 13
```

**Пример подключения на JavaScript:**
```javascript
const token = 'your-jwt-token';
const ws = new WebSocket(`ws://localhost:8081/api/ws`, {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

ws.onopen = () => {
  console.log('WebSocket connected');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
  
  if (message.type === 'category_update') {
    // Обработка обновления категории
    console.log('Category update:', message.payload);
  } else if (message.type === 'statement_update') {
    // Обработка обновления statement
    console.log('Statement update:', message.payload);
  }
};

ws.onclose = () => {
  console.log('WebSocket disconnected');
};
```

**Формат сообщений:**

1. **Уведомление о создании категории:**
```json
{
  "type": "category_update",
  "payload": {
    "action": "created",
    "category": {
      "id": "category_123",
      "title": "Новая категория",
      "userId": "user_123"
    }
  },
  "user_id": "user_123"
}
```

2. **Уведомление об обновлении категории:**
```json
{
  "type": "category_update",
  "payload": {
    "action": "updated",
    "category": {
      "id": "category_123",
      "title": "Обновленная категория",
      "userId": "user_123"
    }
  },
  "user_id": "user_123"
}
```

3. **Уведомление об удалении категории:**
```json
{
  "type": "category_update",
  "payload": {
    "action": "deleted",
    "categoryId": "category_123"
  },
  "user_id": "user_123"
}
```

4. **Уведомление о создании statement:**
```json
{
  "type": "statement_update",
  "payload": {
    "action": "created",
    "statement": {
      "id": "statement_123",
      "text": "Новый statement",
      "userId": "user_123",
      "categoryId": "category_123"
    }
  },
  "user_id": "user_123"
}
```

5. **Уведомление об обновлении statement:**
```json
{
  "type": "statement_update",
  "payload": {
    "action": "updated",
    "statement": {
      "id": "statement_123",
      "text": "Обновленный statement",
      "userId": "user_123",
      "categoryId": "category_123"
    }
  },
  "user_id": "user_123"
}
```

6. **Уведомление об удалении statement:**
```json
{
  "type": "statement_update",
  "payload": {
    "action": "deleted",
    "statementId": "statement_123"
  },
  "user_id": "user_123"
}
```

7. **Подтверждение получения сообщения:**
```json
{
  "type": "ack",
  "payload": "Message received"
}
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