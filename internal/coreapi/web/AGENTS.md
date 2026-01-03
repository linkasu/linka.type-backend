# AGENTS.md

Этот репозиторий содержит backend для LINKa Type — сервиса, который помогает пользователям с нарушениями речи быстро собирать текстовые фразы.

## Структура проекта
- `cmd/core-api` — основной REST API (категории, фразы, состояния, импорт, TTS proxy).
- `cmd/realtime` — long-polling и WebSocket для изменений.
- `cmd/sync-worker` — синхронизация Firebase RTDB -> YDB.
- `internal/` — бизнес-логика, хранилища, конфигурация.
- `docs/` — архитектура, API, миграция и локальная разработка.
- `yc/` — заметки по деплою в Yandex Cloud.
- `internal/coreapi/web` — статическая веб‑морда (index + assets).

## Локальная разработка
- Запуск core-api: `go run ./cmd/core-api`
- Запуск realtime: `go run ./cmd/realtime`
- Запуск sync-worker: `go run ./cmd/sync-worker`
- Тесты: `go test ./...`

## Конфигурация
Все настройки читаются из env (см. `internal/config`). Основные:
- `FIREBASE_PROJECT_ID`, `FIREBASE_DATABASE_URL`
- `FIREBASE_CREDENTIALS_JSON` или `FIREBASE_CREDENTIALS_FILE`
- `YDB_ENDPOINT`, `YDB_DATABASE`, `YDB_TOKEN` (если нет metadata доступа)
- `TTS_PROXY_ENABLED`, `TTS_BASE_URL`
- `SYNC_POLL_INTERVAL`, `SYNC_STREAM_ENABLED`, `SYNC_STREAM_PATH`

## Деплой
См. `yc/README.md` и `.github/workflows/deploy.yml`.
