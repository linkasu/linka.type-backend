# Linka Type backend on Yandex Cloud

## Context from existing apps
The current clients are:
- linka-type-pwa (Vue + TypeScript)
- linkatype-android (Kotlin Multiplatform, Android app)

Both use Firebase Auth (email/password) and Firebase Realtime Database. The PWA also uses Firebase Cloud Functions.

Firebase data model currently in use (from repo inspection):
- /users/{uid}/Category/{categoryId} -> category object: {id, label, created, default?}
- /users/{uid}/Category/{categoryId}/statements/{statementId} -> statement object: {id, categoryId, text, created}
- /users/{uid}/quickes -> array of 6 strings (PWA)
- /users/{uid}/inited -> boolean (PWA onboarding)
- /global/Category -> global categories with statements (admin write)
- /factory/questions -> onboarding question templates
- /admins/{uid} -> admin list

Firebase functions used by the PWA:
- createStatement (can accept questions and generate phrases)
- importFromGlobal
- createCategory (may be used by other clients)

External service:
- TTS API docs: https://tts.linka.su/api (Swagger UI)
- Runtime endpoints: https://tts.linka.su/tts and https://tts.linka.su/voices
  - /tts supports GET (query params) and POST (JSON body) and returns audio/mpeg

Android deletes /users/{uid} on account removal.

## Target platform
Yandex Cloud:
- Serverless Containers (Docker) for API, realtime, and sync worker
- Serverless YDB for data storage
- API Gateway or Application Load Balancer for HTTP and WebSocket routing
- Lockbox for secrets (Firebase admin key, TTS credentials if required)
- Cloud Logging and monitoring

## Service responsibilities
- Authentication and identity mapping (accept Firebase ID tokens initially)
- CRUD for categories, statements, quick phrases, and onboarding flags
- Global categories and admin write access
- Onboarding question templates and phrase generation
- Optional TTS proxy to https://tts.linka.su/tts and https://tts.linka.su/voices for a single backend origin
- Realtime updates for data changes (long polling + WebSocket)
- Bidirectional sync with Firebase during migration

## YDB data model
Tables (proposed):
- users: user_id (PK, Firebase UID), email (nullable), created_at, inited (bool), deleted_at (nullable)
- admins: user_id (PK)
- categories: user_id (PK), category_id (PK), label, created_at, is_default (nullable), updated_at, deleted_at (nullable)
- statements: user_id (PK), category_id (PK), statement_id (PK), text, created_at, updated_at, deleted_at (nullable)
- quickes: user_id (PK), slot (PK, 0-5), text, updated_at
- global_categories: category_id (PK), label, created_at, is_default (nullable), updated_at, deleted_at (nullable)
- global_statements: category_id (PK), statement_id (PK), text, created_at, updated_at, deleted_at (nullable)
- factory_questions: question_id (PK), label, phrases (json or list), category, type, order_index
- changes: user_id (PK), cursor (PK, ULID or monotonic), entity_type, entity_id, op, payload (json), updated_at

Key requirements:
- Preserve existing Firebase ids and created timestamps.
- Accept both 16-char ids (PWA) and Firebase push ids (Android).
- Add updated_at on all writes; legacy data may be missing it.

## API surface (v1)
Auth:
- Authorization: Bearer <Firebase ID token> (initially required for all endpoints).

Data:
- GET /v1/categories
- POST /v1/categories
- PATCH /v1/categories/{id}
- DELETE /v1/categories/{id}
- GET /v1/categories/{id}/statements
- POST /v1/statements
- PATCH /v1/statements/{id}
- DELETE /v1/statements/{id}

User state:
- GET /v1/user/state (inited, quickes)
- PUT /v1/user/state
- GET /v1/quickes
- PUT /v1/quickes

Global and onboarding:
- GET /v1/global/categories
- GET /v1/global/categories/{id}/statements
- POST /v1/global/import {category_id, force}
- GET /v1/factory/questions
- POST /v1/onboarding/phrases {questions} (equivalent to createStatement questions flow)

Account:
- POST /v1/user/delete (delete data + optional Firebase user delete)

Realtime:
- GET /v1/changes?cursor=...&timeout=25s&limit=100
- WS /v1/stream?cursor=...

Optional compatibility:
- POST /v1/tts (proxy to https://tts.linka.su/tts if needed)
- GET /v1/voices (proxy to https://tts.linka.su/voices if needed)

## Realtime model
- Every mutation writes a row into changes.
- Long polling returns changes newer than cursor or waits up to timeout.
- WebSocket pushes batched changes and heartbeats.
- Cursor is opaque (ULID or time-based) and persisted client-side.

## Backward compatibility strategy
- Dual-write: new backend writes to YDB and Firebase Realtime DB.
- Sync worker: stream Firebase RTDB changes to YDB for legacy clients.
- Read-through: for users not yet in YDB, read from Firebase and seed YDB.
- Feature flags by user cohort to enable Yandex backend.

## Non-functional requirements
- Stateless containers with horizontal scale.
- Strict auth and per-user access checks (equivalent to Firebase rules).
- Logs, metrics, tracing per request and per sync pipeline.
