# API Events - Документация

## Обзор

API Events предоставляет функциональность для создания событий пользователей. События используются для логирования действий пользователей и могут содержать произвольные данные.

## Модель данных

### Event

```go
type Event struct {
    ID        string `json:"id" db:"id"`
    UserID    string `json:"userId" db:"user_id"`
    Event     string `json:"event" db:"event"`
    Data      string `json:"data" db:"data"`
    CreatedAt string `json:"createdAt" db:"created_at"`
}
```

### Поля

- `id` - Уникальный идентификатор события (автогенерируется)
- `userId` - ID пользователя, которому принадлежит событие
- `event` - Тип события (обязательное поле)
- `data` - Дополнительные данные события в формате JSON строки (опционально)
- `createdAt` - Время создания события (автоматически устанавливается)

## Эндпойнты

Эндпойнт требует аутентификации через JWT токен.

### Создать событие

**POST** `/api/events`

Создает новое событие для текущего пользователя.

#### Заголовки
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

#### Тело запроса
```json
{
  "event": "user_action",
  "data": "{\"action\": \"button_click\", \"element\": \"submit_button\"}"
}
```

#### Поля запроса
- `event` (обязательное) - Тип события
- `data` (опциональное) - Дополнительные данные в формате JSON строки

#### Ответ
```json
{
  "id": "event_789",
  "userId": "user_456",
  "event": "user_action",
  "data": "{\"action\": \"button_click\", \"element\": \"submit_button\"}",
  "createdAt": "2024-01-15T10:35:00Z"
}
```

#### Коды ответов
- `201` - Событие создано
- `400` - Неверные данные запроса
- `401` - Не авторизован

## Примеры использования

### Создание события входа пользователя

```bash
curl -X POST http://localhost:8080/api/events \
  -H "Authorization: Bearer your_jwt_token" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "user_login",
    "data": "{\"ip\": \"192.168.1.100\", \"userAgent\": \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36\"}"
  }'
```

### Создание события действия пользователя

```bash
curl -X POST http://localhost:8080/api/events \
  -H "Authorization: Bearer your_jwt_token" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "category_created",
    "data": "{\"categoryId\": \"cat_123\", \"categoryTitle\": \"Work Tasks\"}"
  }'
```

## Рекомендации по использованию

1. **Типы событий**: Используйте осмысленные имена для поля `event`, например:
   - `user_login` - вход пользователя
   - `user_logout` - выход пользователя
   - `category_created` - создание категории
   - `statement_created` - создание утверждения
   - `data_exported` - экспорт данных

2. **Данные события**: Поле `data` должно содержать валидную JSON строку с дополнительной информацией о событии.

3. **Безопасность**: События автоматически привязываются к аутентифицированному пользователю.

## Ошибки

### Общие ошибки

- `400 Bad Request` - Неверные данные запроса (например, пустое поле `event`)
- `401 Unauthorized` - Отсутствует или неверный JWT токен
- `500 Internal Server Error` - Внутренняя ошибка сервера

### Примеры ошибок

```json
{
  "error": "Event cannot be empty"
}
```

```json
{
  "error": "User not authenticated"
}
```
