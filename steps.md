# Migration steps from Firebase to Yandex Cloud

## Phase 1: Inventory and contracts
1) Map current Firebase usage from code and config:
   - Auth: email/password sign in, sign up, password reset.
   - RTDB paths: /users/{uid}/Category, /users/{uid}/quickes, /users/{uid}/inited, /global/Category, /factory/questions, /admins.
   - Cloud Functions: createStatement, importFromGlobal, createCategory.
   - TTS external service: docs at https://tts.linka.su/api, runtime at /tts and /voices.
2) Export sample data and record id formats (PWA 16-char ids, Android push ids).
3) Define the API contract and YDB schema that preserves ids, created timestamps, and optional fields (default).
4) Decide conflict policy (last-write-wins) and change cursor format.

## Phase 2: Yandex Cloud foundation
1) Create Yandex Cloud folder, service accounts, and IAM roles.
2) Create VPC, subnets, and security groups for Serverless Containers.
3) Provision Serverless YDB and create tables (users, categories, statements, quickes, global, factory, admins, changes).
4) Create Container Registry and base images for:
   - core-api
   - realtime-ws
   - sync-worker
5) Configure API Gateway or Application Load Balancer for HTTP + WebSocket routes.
6) Set up Lockbox for secrets and Cloud Logging/metrics.

## Phase 3: Build the backend (compatibility-first)
1) Implement auth middleware that validates bearer tokens (Firebase ID tokens).
2) Implement CRUD endpoints for categories and statements.
3) Implement quickes and inited state endpoints.
4) Implement global categories list and importFromGlobal logic.
5) Implement onboarding phrase generation using factory/questions (createStatement questions flow).
6) Implement user delete endpoint (delete YDB data and optionally Firebase user).
7) Write a changes entry on every mutation.

## Phase 4: Realtime updates
1) Implement long polling: GET /v1/changes with cursor, limit, and timeout.
2) Implement WebSocket stream using the same cursor semantics.
3) Define client reconnect and cursor persistence rules.

## Phase 5: Migration bridge (bidirectional sync)
1) Dual-write: all new API writes go to YDB and Firebase RTDB using the same ids.
2) Sync worker from Firebase to YDB:
   - Use RTDB REST streaming to capture create/update/delete events.
   - Translate RTDB paths to YDB rows and write changes entries.
   - Track worker resume cursor to avoid data loss.
3) Backfill:
   - Export RTDB JSON.
   - Transform to YDB import format.
   - Bulk import and validate counts/hashes per user.
4) Read-through for users not yet in YDB (seed on first access).

## Phase 6: Client updates
PWA (linka-type-pwa):
1) Add a backend selector (feature flag) and an API client.
2) Replace Store data access with Yandex endpoints for categories, statements, quickes, and inited.
3) Replace Firebase Functions calls (createStatement, importFromGlobal) with HTTP.
4) Add long polling and WebSocket listeners; keep Firebase listeners as fallback.
5) Send bearer token (default via /v1/auth) in Authorization header for API calls.
6) Use https://tts.linka.su/tts for audio and https://tts.linka.su/voices for voice list (or keep a compatibility proxy if needed).

Android KMP (linkatype-android):
1) Introduce a Repository interface with FirebaseRepository and YandexRepository.
2) Implement YandexRepository using REST + long polling/WebSocket.
3) Add a local cache (Room or SQLDelight) to preserve offline behavior once Firebase is bypassed.
4) Keep Firebase Auth for sign-in; include ID token in API calls.
5) Update account deletion to call the new backend endpoint.
6) Use https://tts.linka.su/tts for audio and https://tts.linka.su/voices for voice list (or keep a compatibility proxy if needed).

## Phase 7: Gradual rollout
1) Deploy backend in shadow mode with dual-write and sync enabled.
2) Enable Yandex backend for a small cohort (hash user_id).
3) Monitor latency, error rate, and sync lag; compare data parity for sampled users.
4) Expand cohorts while keeping Firebase as a fallback.

## Phase 8: Cutover and decommission
1) Switch YDB to source of truth; keep Firebase read-only for legacy clients.
2) Keep sync worker running until legacy usage is negligible.
3) Remove Firebase dependencies in later client releases.
4) Decommission Firebase services after final parity checks.
