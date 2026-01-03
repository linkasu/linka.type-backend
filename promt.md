# Prompt: Build Linka Type backend on Yandex Cloud in Go

## Objective
Build a Go backend for Linka Type on Yandex Cloud (serverless containers + serverless YDB) with realtime updates (WebSocket + long polling) and a safe migration path from Firebase. Backward compatibility is mandatory because users update apps slowly.

## Existing clients and behavior (must preserve)
Clients:
- PWA (Vue + TypeScript) in `linka-type-pwa`
- Android KMP app in `linkatype-android`

Auth:
- Firebase Auth email/password sign-in, sign-up, password reset.
- Clients keep persistent sessions.

Firebase Realtime Database paths currently used:
- `/users/{uid}/Category/{categoryId}`: { id, label, created, default? }
- `/users/{uid}/Category/{categoryId}/statements/{statementId}`: { id, categoryId, text, created }
- `/users/{uid}/quickes`: array of 6 strings (PWA)
- `/users/{uid}/inited`: boolean (PWA onboarding)
- `/global/Category`: global categories with nested statements (admin write)
- `/factory/questions`: onboarding question templates
- `/admins/{uid}`: admin list

Cloud Functions logic in the PWA that must be replicated:
- `createStatement`:
  - If called with `questions`, check `/users/{uid}/inited`.
  - If not inited, generate phrases from questions (replace "%%" with value), create categories by name if needed, set inited = true.
  - If already inited, do not generate duplicates, still return success.
- `importFromGlobal`:
  - Copy `/global/Category/{id}` to `/users/{uid}/Category/{id}`.
  - If already exists and `force=false`, return "exists".
- `createCategory`:
  - Create category with `default` flag when provided.
- No chatbot is required.

External services:
- TTS service:
  - Docs UI: `https://tts.linka.su/api`
  - Runtime endpoints: `https://tts.linka.su/tts` and `https://tts.linka.su/voices`
  - `/tts` supports GET (query params) and POST (JSON body), returns `audio/mpeg`.
  - `/voices` returns list of voice objects.
  - Do not call `/api/tts` (returns 404).
- Word prediction uses Yandex predictor API directly from the client; do not implement on backend.

Key client behaviors:
- PWA sorts categories by `default` first, then `created` ascending.
- Android sorts categories/statements by `created` descending.
- Categories created by Android may not include a `statements` field (PWA sets it to `{}`), so backend must not rely on it.
- Android deletes `/users/{uid}` on account deletion, then deletes Firebase user.
 - Quick phrases default list is defined in `linka-type-pwa/src/blocks/Quickes.vue` and should be used when `quickes` is missing.

## Target architecture (Yandex Cloud)
Components:
- `core-api` (REST): CRUD, onboarding, admin checks, importFromGlobal, quickes/inited, delete account.
- `realtime` (WS + long polling): pushes change events per user.
- `sync-worker`: Firebase -> YDB sync (for legacy clients).

Platform:
- Serverless Containers (Docker)
- Serverless YDB
- API Gateway or Application Load Balancer for HTTP + WebSocket routing
- Lockbox for secrets (Firebase Admin key, optional TTS credentials)
- Cloud Logging + metrics

## Backward compatibility and migration
- Dual-write: all new backend writes must also write to Firebase RTDB using the same IDs.
- Sync worker: stream Firebase RTDB changes and apply to YDB.
- Read-through: if user data missing in YDB, fetch from Firebase and seed.
- Feature flag by user cohort to switch reads to YDB while keeping Firebase as fallback.
- Keep Firebase as source of truth until stability is proven.

Conflict policy:
- Last-write-wins using `updated_at` (server time).
- If timestamps equal, prefer YDB write or latest event arrival.

## Data model (YDB)
Tables (proposed keys):
- `users`: (user_id PK), email nullable, created_at, inited bool, deleted_at nullable
- `admins`: (user_id PK)
- `categories`: (user_id PK, category_id PK), label, created_at, is_default nullable, updated_at, deleted_at nullable
- `statements`: (user_id PK, category_id PK, statement_id PK), text, created_at, updated_at, deleted_at nullable
- `quickes`: (user_id PK, slot PK 0-5), text, updated_at
- `global_categories`: (category_id PK), label, created_at, is_default nullable, updated_at, deleted_at nullable
- `global_statements`: (category_id PK, statement_id PK), text, created_at, updated_at, deleted_at nullable
- `factory_questions`: (question_id PK), label, phrases (json/list), category, type, order_index
- `changes`: (user_id PK, cursor PK), entity_type, entity_id, op, payload json, updated_at

Notes:
- Preserve existing IDs and `created` timestamps.
- Accept both 16-char IDs and Firebase push IDs.
- Always set `updated_at` on write; backfill missing values during sync.

## API requirements (v1)
Auth:
- Authorization: `Bearer <token>` on all endpoints except `/v1/auth` (default via /v1/auth).

Data:
- `GET /v1/categories`
- `POST /v1/categories`
- `PATCH /v1/categories/{id}`
- `DELETE /v1/categories/{id}`
- `GET /v1/categories/{id}/statements`
- `POST /v1/statements`
- `PATCH /v1/statements/{id}`
- `DELETE /v1/statements/{id}`

User state:
- `GET /v1/user/state` (returns `inited` + `quickes`)
- `PUT /v1/user/state`
- `GET /v1/quickes`
- `PUT /v1/quickes`

Global and onboarding:
- `GET /v1/global/categories` (optionally include nested statements for import preview)
- `GET /v1/global/categories/{id}/statements`
- `POST /v1/global/import` body `{category_id, force}`; return `"exists"` when applicable
- `GET /v1/factory/questions`
- `POST /v1/onboarding/phrases` body `{questions}`; mimic Firebase logic

Account:
- `POST /v1/user/delete` (delete YDB data; optionally delete Firebase user)

Realtime:
- `GET /v1/changes?cursor=...&timeout=25s&limit=100`
- `WS /v1/stream?cursor=...`

Optional proxy:
- `POST /v1/tts` -> `https://tts.linka.su/tts`
- `GET /v1/voices` -> `https://tts.linka.su/voices`

## Realtime design
- Every mutation writes a row into `changes`.
- Long polling waits up to `timeout` and returns when new changes exist.
- WebSocket pushes batches of changes and heartbeats.
- Cursor is opaque (ULID or monotonic counter) and must be persisted client-side.

## Auth and security
- Validate Firebase ID tokens using Firebase Admin SDK or public key verification.
- Enforce per-user access (equivalent to Firebase rules):
  - Users can read/write only their own data.
  - Global and factory read for authenticated users.
  - Global and factory write only for admins.
- Admin list stored in `admins` table (sync from Firebase `/admins`).

## Implementation notes (Go)
- Use Go modules, structured logging, and context-aware DB calls.
- Keep services stateless; store session state on client only.
- Ensure idempotent writes to support retries.
- Propagate request ID and user ID in logs.
- Provide Dockerfiles and entrypoints for each service.

## Acceptance criteria
- All endpoints above implemented and tested locally.
- Dual-write and sync worker keep YDB in parity with Firebase.
- New clients can switch to YDB without breaking legacy clients.
- Realtime updates work via long polling and WebSocket.
- TTS is accessible via direct service or optional proxy.
