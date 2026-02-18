# LINKa Type API — гайд для клиентов

Документ для мобильных/веб‑клиентов и интеграций, которые работают с фразами и категориями пользователя.

## Base URL
Используйте базовый URL продакшн API.

Пример:
```
https://backend.linka.su
```

## Аутентификация
Все запросы (кроме `POST /v1/auth` и `GET /v1/voices`) требуют токен доступа.
По умолчанию используйте токен, полученный через открытый backend-эндпоинт `POST /v1/auth` (email + пароль).

Пример получения токена:
```bash
curl -sS -X POST \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret"}' \
  https://backend.linka.su/v1/auth
```

## Health
```
GET /healthz
```
Ответ:
```
{"status":"ok"}
```

## Категории
```
GET /v1/categories
POST /v1/categories
PATCH /v1/categories/{id}
DELETE /v1/categories/{id}
```

Создание категории:
```bash
curl -sS -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"label":"Мои фразы","created":1735689600000}' \
  https://backend.linka.su/v1/categories
```

## Фразы (statements)
```
GET /v1/categories/{id}/statements
POST /v1/statements
PATCH /v1/statements/{id}
DELETE /v1/statements/{id}
```

Создание фразы:
```bash
curl -sS -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"categoryId":"<categoryId>","text":"Здравствуйте!"}' \
  https://backend.linka.su/v1/statements
```

## Состояние пользователя и быстрые ответы
```
GET /v1/user/state
PUT /v1/user/state
GET /v1/quickes
PUT /v1/quickes
```

Пример обновления быстрых ответов:
```bash
curl -sS -X PUT \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"quickes":["Да","Нет","Спасибо","Мне нужна помощь"]}' \
  https://backend.linka.su/v1/quickes
```

## Глобальные наборы и онбординг
```
GET /v1/global/categories?include_statements=true
GET /v1/global/categories/{id}/statements
POST /v1/global/import
GET /v1/factory/questions
POST /v1/onboarding/phrases
```

Импорт глобальной категории:
```bash
curl -sS -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"category_id":"<categoryId>","force":false}' \
  https://backend.linka.su/v1/global/import
```

## Удаление пользователя
```
POST /v1/user/delete
```

```bash
curl -sS -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"delete_firebase":false}' \
  https://backend.linka.su/v1/user/delete
```

## Realtime
Long polling:
```
GET /v1/changes?cursor=...&timeout=25s&limit=100
```

WebSocket:
```
WS /v1/stream?cursor=...
```

Ответы:
- `type:"changes"` — список изменений + новый cursor.
- `type:"heartbeat"` — поддержка соединения.

## Optional TTS proxy
Если включен прокси, доступны:
```
GET /v1/voices
POST /v1/tts
```

`GET /v1/voices` доступен без токена. `POST /v1/tts` требует bearer token.

## Формат ошибок
```
{"error":{"code":"unauthorized","message":"..."}}
```

## Примечания
- Все timestamps — epoch в миллисекундах.
- ID могут задаваться клиентом или генерироваться сервером.
- Все ответы — JSON.
