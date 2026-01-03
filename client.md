# LINKa Type API — гайд для клиентов

Документ для мобильных/веб‑клиентов и интеграций, которые работают с фразами и категориями пользователя.

## Base URL
Используйте домен API Gateway или URL контейнера `core-api`.

Пример:
```
https://<gateway-id>.apigw.yandexcloud.net
```

## Аутентификация
Все запросы требуют заголовок:
```
Authorization: Bearer <Firebase ID token>
```

Firebase ID token должен быть получен через Firebase Auth в клиентском приложении.

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
  https://<gateway-id>.apigw.yandexcloud.net/v1/categories
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
  https://<gateway-id>.apigw.yandexcloud.net/v1/statements
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
  https://<gateway-id>.apigw.yandexcloud.net/v1/quickes
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
  https://<gateway-id>.apigw.yandexcloud.net/v1/global/import
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
  https://<gateway-id>.apigw.yandexcloud.net/v1/user/delete
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

## Формат ошибок
```
{"error":{"code":"unauthorized","message":"..."}}
```

## Примечания
- Все timestamps — epoch в миллисекундах.
- ID могут задаваться клиентом или генерироваться сервером.
- Все ответы — JSON.
