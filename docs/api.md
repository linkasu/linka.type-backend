# API v1

All endpoints require `Authorization: Bearer <Firebase ID token>`.

## Conventions
- Timestamps are epoch milliseconds (int64).
- For compatibility, category fields use `created` and `default` (instead of `created_at`/`is_default`).
- `updated_at` is returned for conflict resolution and debug.
- IDs may be client-provided or server-generated.
- Error shape:
  ```json
  {"error": {"code": "unauthorized", "message": "..."}}
  ```

## Categories
- `GET /v1/categories`
  - Returns: `[{id, label, created, default?, updated_at?}]`

- `POST /v1/categories`
  - Body: `{id?, label, created?, default?}`
  - Returns: `{id, label, created, default?, updated_at}`

- `PATCH /v1/categories/{id}`
  - Body: `{label?, default?}`
  - Returns: `{id, label, created, default?, updated_at}`

- `DELETE /v1/categories/{id}`
  - Returns: `{status:"ok"}`

## Statements
- `GET /v1/categories/{id}/statements`
  - Returns: `[{id, categoryId, text, created, updated_at?}]`

- `POST /v1/statements`
  - Body: `{id?, categoryId, text, created?, questions?}`
  - If `questions` is present, runs onboarding phrase generation and sets `inited` if needed.
  - Returns:
    - `{status:"ok"}` for onboarding requests.
    - `{status:"ok", statement:{...}}` for regular statement creation.

- `PATCH /v1/statements/{id}`
  - Body: `{text?}`
  - Returns: `{id, categoryId, text, created, updated_at}`

- `DELETE /v1/statements/{id}`
  - Returns: `{status:"ok"}`

## User state
- `GET /v1/user/state`
  - Returns: `{inited: bool, quickes: [string, ...]}`

- `PUT /v1/user/state`
  - Body: `{inited?, quickes?}`
  - Returns: `{inited: bool, quickes: [string, ...]}`

- `GET /v1/quickes`
  - Returns: `[string, ...]`

- `PUT /v1/quickes`
  - Body: `{quickes: [string, ...]}`
  - Returns: `[string, ...]`

## Global and onboarding
- `GET /v1/global/categories?include_statements=true`
  - Returns: `[{id, label, created, default?, statements?}]`

- `GET /v1/global/categories/{id}/statements`
  - Returns: `[{id, categoryId, text, created}]`

- `POST /v1/global/import`
  - Body: `{category_id, force}`
  - Returns: `{status:"ok"}` or `{status:"exists"}`

- `GET /v1/factory/questions`
  - Returns: `[{id, label, phrases, category, type, order_index}]`

- `POST /v1/onboarding/phrases`
  - Body: `{questions: [{question_id, value}]}`
  - Returns: `{status:"ok"}`

## Account
- `POST /v1/user/delete`
  - Body: `{delete_firebase?: bool}`
  - Returns: `{status:"ok"}`

## Realtime
- `GET /v1/changes?cursor=...&timeout=25s&limit=100`
  - Returns: `{cursor, changes: [{entity_type, entity_id, op, payload, updated_at}]}`

- `WS /v1/stream?cursor=...`
  - Server messages:
    - `{type:"changes", cursor, changes:[...]}`
    - `{type:"heartbeat", cursor}`

## Optional TTS proxy
- `POST /v1/tts` -> `https://tts.linka.su/tts`
- `GET /v1/voices` -> `https://tts.linka.su/voices`
